package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/your-org/your-service/internal/pkg/correlation"
)

const (
	// HeaderCorrelationID is the preferred client header for distributed tracing.
	HeaderCorrelationID = "X-Correlation-ID"
	// HeaderRequestID is accepted as an alias; echoed on responses.
	HeaderRequestID = "X-Request-ID"
)

const (
	ctxCorrelationID = "correlation_id"
	ctxRequestLogger = "request_logger"
)

// CorrelationID ensures every request has a correlation ID: reuse client headers or generate one.
// It sets response headers, stores a request-scoped Zap logger on the Gin context, and attaches
// the ID to the OpenTelemetry span when present.
func CorrelationID(root *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := strings.TrimSpace(c.GetHeader(HeaderCorrelationID))
		if id == "" {
			id = strings.TrimSpace(c.GetHeader(HeaderRequestID))
		}
		if id == "" {
			id = uuid.New().String()
		}

		c.Writer.Header().Set(HeaderCorrelationID, id)
		c.Writer.Header().Set(HeaderRequestID, id)
		c.Set(ctxCorrelationID, id)

		reqLog := root.With(
			zap.String("correlation_id", id),
			zap.String("request_id", id),
		)
		c.Set(ctxRequestLogger, reqLog)

		c.Request = c.Request.WithContext(correlation.WithID(c.Request.Context(), id))

		if span := trace.SpanFromContext(c.Request.Context()); span.SpanContext().IsValid() {
			span.SetAttributes(attribute.String("correlation.id", id))
		}

		c.Next()
	}
}

func CorrelationIDFromGin(c *gin.Context) string {
	if v, ok := c.Get(ctxCorrelationID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// RequestLogger returns the request-scoped logger (with correlation fields), or a no-op logger.
func RequestLogger(c *gin.Context) *zap.Logger {
	if v, ok := c.Get(ctxRequestLogger); ok {
		if l, ok := v.(*zap.Logger); ok {
			return l
		}
	}
	return zap.NewNop()
}

// RequestLoggerOrFallback returns the request logger when CorrelationID ran; otherwise root.
func RequestLoggerOrFallback(c *gin.Context, root *zap.Logger) *zap.Logger {
	if v, ok := c.Get(ctxRequestLogger); ok {
		if l, ok := v.(*zap.Logger); ok {
			return l
		}
	}
	return root
}
