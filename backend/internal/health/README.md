# internal/health/

Comprehensive health check system for production observability.

| Endpoint | Purpose |
|----------|---------|
| `GET /health` | Full health report: MariaDB (version, pool stats, latency), Redis (latency, pool, memory), system (goroutines, memory, CPU, GC). Returns `503` if any required dependency is down. |
| `GET /health/live` | Kubernetes liveness probe (always 200 if the process is alive). |
| `GET /health/ready` | Kubernetes readiness probe (200 only when MariaDB + Redis are reachable). |

Endpoints registered at both `/health` (infra) and `/api/v1/health` (versioned API).

To add a dependency check (e.g., Meilisearch, a partner API), extend `Checker` and the `Components` map.
