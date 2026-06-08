package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Server struct {
	http *http.Server
	log  *zap.Logger
}

// New serves the given Gin engine.
func New(router *gin.Engine, port int, readTimeout, writeTimeout, idleTimeoutSec, maxHeaderBytes int, log *zap.Logger) *Server {
	return NewWithHandler(router, port, readTimeout, writeTimeout, idleTimeoutSec, maxHeaderBytes, log)
}

// NewWithHandler serves any http.Handler (e.g. Gin engine or a wrapper).
// idleTimeoutSec and maxHeaderBytes may be 0 to use defaults (120s, 1 MiB).
func NewWithHandler(handler http.Handler, port int, readTimeout, writeTimeout, idleTimeoutSec, maxHeaderBytes int, log *zap.Logger) *Server {
	if idleTimeoutSec <= 0 {
		idleTimeoutSec = 120
	}
	if maxHeaderBytes <= 0 {
		maxHeaderBytes = 1 << 20
	}
	return &Server{
		http: &http.Server{
			Addr:           fmt.Sprintf(":%d", port),
			Handler:        handler,
			ReadTimeout:    time.Duration(readTimeout) * time.Second,
			WriteTimeout:   time.Duration(writeTimeout) * time.Second,
			IdleTimeout:    time.Duration(idleTimeoutSec) * time.Second,
			MaxHeaderBytes: maxHeaderBytes,
		},
		log: log,
	}
}

func (s *Server) Run() error {
	s.log.Info("server starting", zap.String("addr", s.http.Addr))
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}
