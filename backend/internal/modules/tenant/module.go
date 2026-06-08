package tenant

import (
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/your-org/your-service/internal/cache"
)

// Module bundles the tenant module's wiring so bootstrap can compose modules
// with a single call. Repo and Service are exported so other modules (auth,
// user) can take typed dependencies on them.
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
