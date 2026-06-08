package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	_ "github.com/your-org/your-service/docs/swagger"
	"github.com/your-org/your-service/internal/bootstrap"
	"github.com/your-org/your-service/internal/config"
	"github.com/your-org/your-service/internal/pkg/logger"
	"github.com/your-org/your-service/internal/pkg/response"
	"github.com/your-org/your-service/internal/server"
)

// @title           Your Service API
// @version         0.1.0
// @description     HTTP API for Your Service.
// @host            api.example.com
// @schemes         https http
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log, err := logger.New(cfg.App.Env)
	if err != nil {
		panic(err)
	}

	// Bind HTTP immediately so Docker health / curl never see "connection reset" while bootstrap runs.
	stub := gin.New()
	stub.Use(gin.Recovery())
	stub.GET("/health/live", func(c *gin.Context) { c.Status(http.StatusOK) })
	stub.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/v1") {
			response.Fail(c, http.StatusServiceUnavailable, "SERVICE_STARTING",
				"API is still starting; retry shortly.", nil)
			return
		}
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   gin.H{"code": "SERVICE_STARTING", "message": "API is still starting."},
		})
	})

	var current atomic.Value
	current.Store(stub)
	root := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current.Load().(http.Handler).ServeHTTP(w, r)
	})

	srv := server.NewWithHandler(
		root,
		cfg.Server.Port,
		cfg.Server.ReadTimeout,
		cfg.Server.WriteTimeout,
		cfg.Server.IdleTimeoutSec,
		cfg.Server.MaxHeaderBytes,
		log,
	)

	var apiReady atomic.Pointer[bootstrap.API]
	ctx := context.Background()
	go func() {
		api, err := bootstrap.BootstrapAPI(ctx, cfg, log)
		if err != nil {
			log.Fatal("bootstrap failed", zap.Error(err))
		}
		current.Store(api.Router)
		apiReady.Store(api)
		log.Info("routes loaded")
	}()

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		if api := apiReady.Load(); api != nil {
			api.Close()
		}
	}()

	if err := srv.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("server error", zap.Error(err))
	}
}
