package correlation

import (
	"context"
	"testing"
)

func TestFromContext_empty(t *testing.T) {
	if got := FromContext(context.Background()); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestWithID_roundTrip(t *testing.T) {
	ctx := WithID(context.Background(), "abc-123")
	if got := FromContext(ctx); got != "abc-123" {
		t.Fatalf("got %q", got)
	}
}
