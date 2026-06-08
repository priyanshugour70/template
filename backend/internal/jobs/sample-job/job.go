// Package samplejob is a reference scheduled job. Copy + rename per use case.
package samplejob

import (
	"context"

	"go.uber.org/zap"
)

// RunOnce performs a single invocation of the job. Idempotent.
// Wire this into cmd/worker/main.go behind a timer or cron library.
func RunOnce(ctx context.Context, log *zap.Logger) (ok int, fail int) {
	select {
	case <-ctx.Done():
		return 0, 0
	default:
	}
	log.Info("sample-job: TODO implement work")
	return 0, 0
}
