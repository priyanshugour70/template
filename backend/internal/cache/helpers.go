package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// TTL presets for different data categories.
const (
	TTLStatic  = 30 * time.Minute // master data, statuses, roles, permissions
	TTLWarm    = 5 * time.Minute  // list queries with filters
	TTLHot     = 60 * time.Second // frequently changing data (counts, boards)
	TTLSession = 2 * time.Minute  // RBAC state, user sessions
)

// keyPrefix is prepended to every namespaced key. Change per service.
const keyPrefix = "app"

// Key builds a namespaced cache key: "<prefix>:{module}:{segments...}".
func Key(module string, segments ...string) string {
	parts := make([]string, 0, 2+len(segments))
	parts = append(parts, keyPrefix, module)
	parts = append(parts, segments...)
	return strings.Join(parts, ":")
}

// KeyWithParams builds a cache key including sorted query parameters.
// Deterministic regardless of map iteration order.
func KeyWithParams(module, method string, params map[string]string) string {
	if len(params) == 0 {
		return Key(module, method)
	}
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte('&')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(params[k])
	}
	h := sha256.Sum256([]byte(b.String()))
	return Key(module, method, hex.EncodeToString(h[:8]))
}

// DeleteByPrefix removes all keys matching "prefix*". Only works with RedisCache.
func DeleteByPrefix(ctx context.Context, c Cache, prefix string) error {
	rc, ok := c.(*RedisCache)
	if !ok {
		return nil
	}
	var cursor uint64
	for {
		keys, next, err := rc.client.Scan(ctx, cursor, prefix+"*", 100).Result()
		if err != nil {
			return fmt.Errorf("scan %s*: %w", prefix, err)
		}
		if len(keys) > 0 {
			if err := rc.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("del keys: %w", err)
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return nil
}

// singleFlight prevents cache stampede: only one goroutine fetches when
// multiple requests hit a cold key simultaneously.
var flights = &singleFlightGroup{calls: make(map[string]*call)}

type call struct {
	wg  sync.WaitGroup
	val string
	err error
}

type singleFlightGroup struct {
	mu    sync.Mutex
	calls map[string]*call
}

// GetOrLoad implements cache-aside with stampede protection.
// On cache miss, only one goroutine calls loader(); others wait.
func GetOrLoad[T any](ctx context.Context, c Cache, key string, ttl time.Duration, loader func() (*T, error)) (*T, error) {
	cached, err := GetJSON[T](ctx, c, key)
	if err == nil && cached != nil {
		return cached, nil
	}

	flights.mu.Lock()
	if existing, ok := flights.calls[key]; ok {
		flights.mu.Unlock()
		existing.wg.Wait()
		if existing.err != nil {
			return nil, existing.err
		}
		result, err := GetJSON[T](ctx, c, key)
		if err != nil || result == nil {
			return loader()
		}
		return result, nil
	}
	cl := &call{}
	cl.wg.Add(1)
	flights.calls[key] = cl
	flights.mu.Unlock()

	result, loadErr := loader()
	if loadErr == nil && result != nil {
		_ = SetJSON(ctx, c, key, result, ttl)
	}
	cl.err = loadErr
	cl.wg.Done()

	flights.mu.Lock()
	delete(flights.calls, key)
	flights.mu.Unlock()

	return result, loadErr
}
