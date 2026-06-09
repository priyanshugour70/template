package billing

import (
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/your-org/your-service/internal/cache"
	"github.com/your-org/your-service/internal/modules/tenant"
	"github.com/your-org/your-service/internal/pkg/mail"
	"github.com/your-org/your-service/internal/pkg/storage"
	"github.com/your-org/your-service/internal/queue"
)

type Module struct {
	Handler    *Handler
	Service    *Service
	Repo       *Repository
	Middleware *Middleware
}

// New wires the billing module. Optional dependencies:
//   - s3 may be nil if the bucket isn't configured (PDF endpoints will 503).
//   - tenantSvc must be non-nil for "bill to" rendering on invoice PDFs.
//   - mailer may be nil — falls back to mail.NoopSender so dev environments
//     without SMTP don't crash.
func New(
	db *gorm.DB,
	log *zap.Logger,
	c cache.Cache,
	p queue.Producer,
	s3 *storage.S3,
	tenantSvc *tenant.Service,
	mailer mail.Sender,
) *Module {
	repo := NewRepository(db)
	svc := NewService(repo, log, c, p, s3, tenantSvc, mailer)
	mw := NewMiddleware(svc)
	h := NewHandler(svc, log)
	return &Module{Handler: h, Service: svc, Repo: repo, Middleware: mw}
}
