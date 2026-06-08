# cmd/worker/

Background worker entrypoint.

Responsibilities:

1. Subscribes to Redis Pub/Sub channels (see `internal/queue/channels.go`).
2. Dispatches messages to handlers registered per channel.
3. Runs cron-like jobs from `internal/jobs/*` (e.g., periodic syncs).
4. Trapped SIGINT/SIGTERM cleanly cancels in-flight work.

Add new background work by:

- Creating a channel constant in `internal/queue/channels.go`.
- Publishing from the API via `queue.Producer.Publish(ctx, channel, payload)`.
- Subscribing in this `main.go` with a `queue.Handler` (or job goroutine).

## Run

```bash
make run-worker
# or
go run ./cmd/worker
```
