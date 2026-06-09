package dashboard

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/your-org/your-service/internal/pkg/response"
)

type Handler struct {
	svc *Service
	log *zap.Logger
}

func NewHandler(svc *Service, log *zap.Logger) *Handler { return &Handler{svc: svc, log: log} }

type PermissionFunc func(perm string) gin.HandlerFunc

// Routes registers the single dashboard summary endpoint. No permission
// gate beyond authentication — every authenticated user can see the
// home-page rollup for their own org; the service layer scopes the data.
func (h *Handler) Routes(g *gin.RouterGroup, auth gin.HandlerFunc, _ PermissionFunc) {
	dash := g.Group("/dashboard", auth)
	{
		dash.GET("/summary", h.summary)
	}
}

func (h *Handler) summary(c *gin.Context) {
	out, err := h.svc.GetSummary(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}
