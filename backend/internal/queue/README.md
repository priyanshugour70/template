# internal/queue/

Event publishing uses **Redis Pub/Sub** (`PUBLISH` / `SUBSCRIBE`). For at-least-once delivery, swap the producer/consumer to Redis Streams or Kafka — the interfaces let you do so without touching call sites.

## Channels

Defined in `channels.go`. Add new ones as your domain evolves.

| Channel | Publishers | Purpose |
|---------|------------|---------|
| `app:audit` | Audit middleware | Audit log events (after primary DB persist) |
| `app:notifications` | Domain services | Async in-app/email notifications |
| `app:search.sync` | Indexable modules | Trigger search index updates |

## Usage

- **API**: publishes via `queue.Producer.Publish(ctx, channel, payload)` on Redis DB `REDIS_QUEUE_DB` (default `3`).
- **Worker**: subscribes on the same DB with `queue.NewRedisConsumer(rdb).Consume(ctx, channel, handler)`.

See `channels.go` for constants and `DefaultChannels()` for the worker subscription list.
