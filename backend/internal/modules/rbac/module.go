package rbac

import (
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/your-org/your-service/internal/cache"
	"github.com/your-org/your-service/internal/queue"
)

type Module struct {
	Handler    *Handler
	Service    *Service
	Repo       *Repository
	Middleware *Middleware
}

func New(db *gorm.DB, log *zap.Logger, c cache.Cache, producer queue.Producer) *Module {
	repo := NewRepository(db)
	svc := NewService(repo, log, c, producer)
	mw := NewMiddleware(svc)
	h := NewHandler(svc, log)
	return &Module{Handler: h, Service: svc, Repo: repo, Middleware: mw}
}
