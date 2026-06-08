# internal/cache/

Cache abstraction with stampede protection.

## Interface

```go
type Cache interface {
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
}
```

Implementations: `RedisCache` (production), `MemoryCache` (testing / Redis-less startup), `NoopCache` (disabled).

## Key features

- `GetOrLoad[T]` — cache-aside with single-flight protection: only one goroutine fetches on cache miss; others wait and reuse.
- `GetJSON / SetJSON` — generic JSON serialization helpers.
- `DeleteByPrefix` — scan + delete for cache invalidation by namespace (Redis only).
- `KeyWithParams` — deterministic cache keys from sorted query parameters (SHA256 over `k=v&…`).

## TTL presets

| Constant | Duration | Use case |
|----------|----------|----------|
| `TTLStatic` | 30 min | Master data, statuses, roles, permissions |
| `TTLWarm` | 5 min | List queries with filters |
| `TTLHot` | 60 sec | Frequently changing data (counts, boards) |
| `TTLSession` | 2 min | RBAC state, user sessions |

## Key namespace

All keys follow: `app:{module}:{segments...}` — change the `"app"` prefix in `helpers.go` if you want a per-service prefix.
