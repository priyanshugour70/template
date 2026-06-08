// Package bootstrap is the composition root. Wire every module, client, and
// middleware here. This is the only package that should import concrete module
// and client packages together.
package bootstrap

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/your-org/your-service/internal/cache"
	"github.com/your-org/your-service/internal/config"
	"github.com/your-org/your-service/internal/health"
	"github.com/your-org/your-service/internal/middleware"
	"github.com/your-org/your-service/internal/pkg/logger"
	"github.com/your-org/your-service/internal/pkg/response"
	"github.com/your-org/your-service/internal/queue"
	"github.com/your-org/your-service/internal/repository"
)

const Version = "0.1.0"

func parseAllowedOrigins(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if o := strings.TrimSpace(p); o != "" {
			out = append(out, o)
		}
	}
	return out
}

// API is the bag of handles built by BootstrapAPI. Closed in main on shutdown.
type API struct {
	Config     *config.Config
	Log        *zap.Logger
	Router     *gin.Engine
	DB         *gorm.DB
	Redis      *redis.Client
	Cache      cache.Cache
	Producer   queue.Producer
	QueueRedis *redis.Client
}

func BootstrapAPI(ctx context.Context, cfg *config.Config, log *zap.Logger) (*API, error) {
	var err error
	if log == nil {
		log, err = logger.New(cfg.App.Env)
		if err != nil {
			return nil, err
		}
	}

	gin.SetMode(cfg.Server.Mode)
	router := gin.New()
	router.MaxMultipartMemory = 50 << 20

	origins := parseAllowedOrigins(cfg.CORS.AllowedOrigins)
	router.Use(cors.New(cors.Config{
		AllowOrigins: origins,
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Origin", "Content-Type", "Accept", "Authorization", "X-Api-Key",
			middleware.HeaderCorrelationID, middleware.HeaderRequestID,
		},
		AllowCredentials: true,
		ExposeHeaders:    []string{middleware.HeaderCorrelationID, middleware.HeaderRequestID},
	}))
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.Recovery(log), middleware.CorrelationID(log), middleware.Logger(log))

	// ── PostgreSQL ─────────────────────────────────────────────────────────
	var db *gorm.DB
	if cfg.Postgres.Host != "" {
		logLevel := "warn"
		if cfg.App.Env == "development" {
			logLevel = "info"
		}
		db, err = repository.OpenPostgres(ctx, repository.DBConfig{
			DSN:            cfg.Postgres.DSN(),
			MaxOpenConns:   cfg.Postgres.MaxOpenConns,
			MaxIdleConns:   cfg.Postgres.MaxIdleConns,
			ConnMaxLifeSec: cfg.Postgres.ConnMaxLife,
			LogLevel:       logLevel,
		}, log)
		if err != nil {
			log.Warn("PostgreSQL unavailable; API starting in degraded mode", zap.Error(err))
			db = nil
		}
	} else {
		log.Warn("POSTGRES_HOST not set; database features disabled")
	}

	// ── Redis ─────────────────────────────────────────────────────────────
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		log.Warn("Redis unavailable, cache and queue will no-op", zap.Error(err))
		_ = rdb.Close()
		rdb = nil
	}
	pingCancel()

	var cacheSvc cache.Cache
	var producer queue.Producer
	var queueRdb *redis.Client
	if rdb != nil {
		cacheSvc = cache.NewRedisCache(rdb)
		queueRdb = newRedisClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.QueueDB)
		queuePingCtx, queuePingCancel := context.WithTimeout(ctx, 5*time.Second)
		queuePingErr := queueRdb.Ping(queuePingCtx).Err()
		queuePingCancel()
		if queuePingErr != nil {
			log.Warn("Redis queue DB unavailable, pub/sub will no-op", zap.Error(queuePingErr))
			_ = queueRdb.Close()
			queueRdb = nil
			producer = &queue.NoopProducer{}
		} else {
			producer = queue.NewRedisProducer(queueRdb)
			log.Info("redis pub/sub producer ready", zap.Int("queue_db", cfg.Redis.QueueDB))
		}
	} else {
		cacheSvc = cache.NewMemoryCache()
		producer = &queue.NoopProducer{}
	}

	// ── Health + Swagger ──────────────────────────────────────────────────
	healthChecker := health.NewChecker(Version, cfg.App.Env, db, rdb)
	healthChecker.Register(router)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.DefaultModelsExpandDepth(-1),
	))
	router.GET("/", func(c *gin.Context) { c.Redirect(http.StatusFound, "/swagger/index.html") })

	api := router.Group("/api/v1")

	if db != nil {
		registerModules(api, db, log, cfg, producer, cacheSvc)
	} else {
		log.Warn("skipping module registration — no database connection")
		router.NoRoute(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/api/v1") {
				response.Fail(c, 503, "SERVICE_UNAVAILABLE",
					"API requires a database connection. Set POSTGRES_* and ensure PostgreSQL is reachable, then restart.",
					nil)
				return
			}
			c.JSON(404, gin.H{"success": false, "error": gin.H{"code": "NOT_FOUND", "message": "not found"}})
		})
	}

	return &API{
		Config:     cfg,
		Log:        log,
		Router:     router,
		DB:         db,
		Redis:      rdb,
		Cache:      cacheSvc,
		Producer:   producer,
		QueueRedis: queueRdb,
	}, nil
}

func newRedisClient(addr, password string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
}

// registerModules wires every domain module. Add new modules here in dependency
// order: repository → service → handler → handler.Register(group).
func registerModules(
	api *gin.RouterGroup,
	db *gorm.DB,
	log *zap.Logger,
	cfg *config.Config,
	producer queue.Producer,
	cacheSvc cache.Cache,
) {
	_ = api
	_ = db
	_ = log
	_ = cfg
	_ = producer
	_ = cacheSvc
	// TODO: add modules here as they are implemented.
	// Pattern:
	//   repo := mymodule.NewRepository(db)
	//   svc := mymodule.NewService(repo, log, cacheSvc)
	//   h := mymodule.NewHandler(svc)
	//   h.Register(api)
}

func (a *API) Close() {
	if a.DB != nil {
		if sqlDB, err := a.DB.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}
	if a.QueueRedis != nil {
		_ = a.QueueRedis.Close()
	}
	if a.Redis != nil {
		_ = a.Redis.Close()
	}
}
