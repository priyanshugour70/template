// Package correlation propagates a correlation ID through context.Context for logging and tracing.
package correlation

import (
	"context"
)

type ctxKey struct{}

// WithID returns a child context that carries correlationID (non-empty).
func WithID(ctx context.Context, correlationID string) context.Context {
	if correlationID == "" {
		return ctx
	}
	return context.WithValue(ctx, ctxKey{}, correlationID)
}

// FromContext returns the correlation ID from ctx, or empty string.
func FromContext(ctx context.Context) string {
	v, _ := ctx.Value(ctxKey{}).(string)
	return v
}
