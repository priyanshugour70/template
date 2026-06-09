package notification

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Module struct {
	Handler *Handler
	Service *Service
	Repo    *Repository
}

func New(db *gorm.DB, log *zap.Logger) *Module {
	repo := NewRepository(db)
	svc := NewService(repo, log)
	h := NewHandler(svc, log)
	return &Module{Handler: h, Service: svc, Repo: repo}
}
