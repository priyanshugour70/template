package comm

import (
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/your-org/your-service/internal/modules/user"
)

// Module is the composition root for the communication module — created in
// bootstrap. Same shape as other modules so the dependency wiring lives in
// one obvious place.
type Module struct {
	Handler *Handler
	Service *Service
	Repo    *Repository
}

// New wires the repository, service, and handler. The notification + rbac
// ports are optional from a type perspective (the handler tolerates a nil
// rbac port and treats the caller as a non-moderator); in real bootstrap
// both should always be non-nil.
func New(
	db *gorm.DB,
	userSvc *user.Service,
	notif NotificationPort,
	rbacPort RBACPort,
	log *zap.Logger,
) *Module {
	repo := NewRepository(db)
	svc := NewService(repo, userSvc, notif, log)
	h := NewHandler(svc, rbacPort, log)
	return &Module{Handler: h, Service: svc, Repo: repo}
}
