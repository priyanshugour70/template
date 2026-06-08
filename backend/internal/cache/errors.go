package cache

import "errors"

// ErrMiss is returned by Cache implementations when a key is absent.
// GetJSON treats ErrMiss and redis.Nil as a cache miss (returns nil, nil).
var ErrMiss = errors.New("cache miss")
