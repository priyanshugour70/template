package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/your-org/your-service/internal/pkg/response"
)

func Recovery(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				l := RequestLoggerOrFallback(c, log)
				l.Error("panic_recovered",
					zap.Any("panic", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, response.Body{
					Success:   false,
					Error:     &response.ErrPayload{Code: "INTERNAL_ERROR", Message: "internal server error"},
					Timestamp: time.Now().UTC().Format(time.RFC3339),
				})
			}
		}()
		c.Next()
	}
}
