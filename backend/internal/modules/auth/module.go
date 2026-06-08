package auth

import (
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/your-org/your-service/internal/cache"
	"github.com/your-org/your-service/internal/config"
	"github.com/your-org/your-service/internal/modules/rbac"
	"github.com/your-org/your-service/internal/modules/tenant"
	"github.com/your-org/your-service/internal/modules/user"
	"github.com/your-org/your-service/internal/pkg/jwt"
	"github.com/your-org/your-service/internal/queue"
)

type Module struct {
	Handler *Handler
	Service *Service
	Repo    *Repository
	Signer  *jwt.Signer
}

func New(
	db *gorm.DB,
	tenantSvc *tenant.Service,
	userSvc *user.Service,
	rbacSvc *rbac.Service,
	cfg *config.Config,
	c cache.Cache,
	producer queue.Producer,
	log *zap.Logger,
) *Module {
	repo := NewRepository(db)
	ttl := minutesToDuration(cfg.Auth.AccessTokenMinutes, 15)
	signer := jwt.NewSigner(cfg.Auth.JWTSecret, ttl, "your-service")
	svc := NewService(repo, tenantSvc, userSvc, rbacSvc, signer, c, producer, cfg, log)
	h := NewHandler(svc, log)
	return &Module{Handler: h, Service: svc, Repo: repo, Signer: signer}
}

func minutesToDuration(m int, fallback int) (d time.Duration) {
	if m <= 0 {
		m = fallback
	}
	return time.Duration(m) * time.Minute
}
