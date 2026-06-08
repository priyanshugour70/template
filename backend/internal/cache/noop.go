package cache

import (
	"context"
	"time"
)

type NoopCache struct{}

func (NoopCache) Get(_ context.Context, _ string) (string, error)                      { return "", ErrMiss }
func (NoopCache) Set(_ context.Context, _ string, _ interface{}, _ time.Duration) error { return nil }
func (NoopCache) Delete(_ context.Context, _ string) error                              { return nil }
func (NoopCache) Exists(_ context.Context, _ string) (bool, error)                      { return false, nil }
