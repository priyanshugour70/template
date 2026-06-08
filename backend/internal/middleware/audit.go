package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/your-org/your-service/internal/pkg/appctx"
	"github.com/your-org/your-service/internal/queue"
)

const (
	// auditMaxBody caps captured request/response body size in bytes. Bodies
	// larger than this are skipped to keep audit_log rows reasonable.
	auditMaxBody = 64 << 10
)

var sensitiveBodyKeys = []string{
	"password", "current_password", "new_password", "old_password",
	"token", "refresh", "refresh_token", "access_token", "id_token",
	"secret", "client_secret", "api_key", "apikey",
	"otp", "code", "verification_code",
	"private_key", "ssh_key",
	"card_number", "cvv", "cvc",
}

var sensitiveHeaders = map[string]bool{
	"authorization":   true,
	"proxy-authorization": true,
	"cookie":          true,
	"set-cookie":      true,
	"x-api-key":       true,
	"x-auth-token":    true,
}

// AuditEvent is the payload published on queue.ChannelAudit. The worker
// (cmd/worker) decodes and writes to the audit_log table.
type AuditEvent struct {
	ID              uuid.UUID         `json:"id"`
	OccurredAt      time.Time         `json:"occurredAt"`
	CorrelationID   string            `json:"correlationId,omitempty"`
	Method          string            `json:"method"`
	Path            string            `json:"path"`
	Route           string            `json:"route,omitempty"`
	StatusCode      int               `json:"statusCode"`
	LatencyMs       int64             `json:"latencyMs"`
	IP              string            `json:"ip,omitempty"`
	UserAgent       string            `json:"userAgent,omitempty"`
	UserID          *uuid.UUID        `json:"userId,omitempty"`
	UserEmail       string            `json:"userEmail,omitempty"`
	TenantID        *uuid.UUID        `json:"tenantId,omitempty"`
	OrganizationID  *uuid.UUID        `json:"organizationId,omitempty"`
	Action          string            `json:"action,omitempty"`
	TargetType      string            `json:"targetType,omitempty"`
	TargetID        *uuid.UUID        `json:"targetId,omitempty"`
	ErrorCode       string            `json:"errorCode,omitempty"`
	RequestHeaders  map[string]string `json:"requestHeaders,omitempty"`
	RequestBody     json.RawMessage   `json:"requestBody,omitempty"`
	ResponseHeaders map[string]string `json:"responseHeaders,omitempty"`
	ResponseBody    json.RawMessage   `json:"responseBody,omitempty"`
}

// AuditConfig allows callers to skip noisy paths (healthchecks, swagger).
type AuditConfig struct {
	SkipPaths       []string
	SkipPathPrefix  []string
	CaptureRequest  bool
	CaptureResponse bool
}

func DefaultAuditConfig() AuditConfig {
	return AuditConfig{
		SkipPaths: []string{
			"/health", "/health/live", "/health/ready", "/",
		},
		SkipPathPrefix:  []string{"/swagger"},
		CaptureRequest:  true,
		CaptureResponse: true,
	}
}

type responseCapture struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseCapture) Write(b []byte) (int, error) {
	if w.body.Len()+len(b) <= auditMaxBody {
		w.body.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

// Audit returns a middleware that captures each request and publishes an
// AuditEvent on queue.ChannelAudit asynchronously. The publish never blocks
// the response.
func Audit(producer queue.Producer, log *zap.Logger, cfg AuditConfig) gin.HandlerFunc {
	skipSet := make(map[string]bool, len(cfg.SkipPaths))
	for _, p := range cfg.SkipPaths {
		skipSet[p] = true
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if skipSet[path] {
			c.Next()
			return
		}
		for _, prefix := range cfg.SkipPathPrefix {
			if strings.HasPrefix(path, prefix) {
				c.Next()
				return
			}
		}

		start := time.Now()

		var reqBody []byte
		if cfg.CaptureRequest && c.Request.Body != nil {
			ct := c.GetHeader("Content-Type")
			if isCapturableMime(ct) && c.Request.ContentLength > 0 && c.Request.ContentLength <= auditMaxBody {
				reqBody, _ = io.ReadAll(io.LimitReader(c.Request.Body, auditMaxBody+1))
				c.Request.Body = io.NopCloser(bytes.NewReader(reqBody))
			}
		}

		var capture *responseCapture
		if cfg.CaptureResponse {
			capture = &responseCapture{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
			c.Writer = capture
		}

		c.Next()

		evt := AuditEvent{
			ID:            uuid.New(),
			OccurredAt:    start,
			CorrelationID: CorrelationIDFromGin(c),
			Method:        c.Request.Method,
			Path:          path,
			Route:         c.FullPath(),
			StatusCode:    c.Writer.Status(),
			LatencyMs:     time.Since(start).Milliseconds(),
			IP:            c.ClientIP(),
			UserAgent:     c.Request.UserAgent(),
		}

		ctx := c.Request.Context()
		if uid := appctx.UserID(ctx); uid != uuid.Nil {
			id := uid
			evt.UserID = &id
		}
		if email := appctx.Email(ctx); email != "" {
			evt.UserEmail = email
		}
		if tid := appctx.TenantID(ctx); tid != uuid.Nil {
			id := tid
			evt.TenantID = &id
		}
		if oid := appctx.OrganizationID(ctx); oid != uuid.Nil {
			id := oid
			evt.OrganizationID = &id
		}

		if cfg.CaptureRequest {
			evt.RequestHeaders = scrubHeaders(c.Request.Header)
			if reqBody != nil && len(reqBody) <= auditMaxBody && isCapturableMime(c.GetHeader("Content-Type")) {
				if s := scrubJSONBytes(reqBody); s != nil {
					evt.RequestBody = s
				}
			}
		}
		if cfg.CaptureResponse && capture != nil {
			evt.ResponseHeaders = scrubHeaders(c.Writer.Header())
			if capture.body.Len() > 0 && capture.body.Len() <= auditMaxBody && isCapturableMime(c.Writer.Header().Get("Content-Type")) {
				if s := scrubJSONBytes(capture.body.Bytes()); s != nil {
					evt.ResponseBody = s
				}
			}
		}
		if last := c.Errors.Last(); last != nil {
			evt.ErrorCode = last.Error()
		}

		go func(payload AuditEvent) {
			pubCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := producer.Publish(pubCtx, queue.ChannelAudit, payload); err != nil {
				log.Warn("audit publish failed",
					zap.Error(err),
					zap.String("correlation_id", payload.CorrelationID),
					zap.String("path", payload.Path),
				)
			}
		}(evt)
	}
}

func isCapturableMime(ct string) bool {
	ct = strings.ToLower(ct)
	switch {
	case strings.HasPrefix(ct, "application/json"):
		return true
	case strings.HasPrefix(ct, "application/x-www-form-urlencoded"):
		return true
	case strings.HasPrefix(ct, "text/"):
		return true
	}
	return false
}

func scrubHeaders(h http.Header) map[string]string {
	out := make(map[string]string, len(h))
	for k, v := range h {
		if sensitiveHeaders[strings.ToLower(k)] {
			out[k] = "[REDACTED]"
			continue
		}
		if len(v) > 0 {
			out[k] = v[0]
		}
	}
	return out
}

func scrubJSONBytes(b []byte) json.RawMessage {
	if len(b) == 0 {
		return nil
	}
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		// Not JSON — store raw if short, else skip.
		if len(b) <= 4096 {
			return json.RawMessage(b)
		}
		return nil
	}
	scrubValue(v)
	out, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return out
}

func scrubValue(v interface{}) {
	switch t := v.(type) {
	case map[string]interface{}:
		for k, val := range t {
			if isSensitiveKey(k) {
				t[k] = "[REDACTED]"
				continue
			}
			scrubValue(val)
		}
	case []interface{}:
		for _, item := range t {
			scrubValue(item)
		}
	}
}

func isSensitiveKey(k string) bool {
	lk := strings.ToLower(k)
	for _, s := range sensitiveBodyKeys {
		if strings.Contains(lk, s) {
			return true
		}
	}
	return false
}
