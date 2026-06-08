package user

import (
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/your-org/your-service/internal/cache"
)

type Module struct {
	Handler *Handler
	Service *Service
	Repo    *Repository
}

func New(db *gorm.DB, log *zap.Logger, cacheSvc cache.Cache) *Module {
	repo := NewRepository(db)
	svc := NewService(repo, log, cacheSvc)
	h := NewHandler(svc, log)
	return &Module{Handler: h, Service: svc, Repo: repo}
}
