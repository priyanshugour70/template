package queue

import "context"

type NoopProducer struct{}

func (NoopProducer) Publish(_ context.Context, _ string, _ interface{}) error { return nil }
