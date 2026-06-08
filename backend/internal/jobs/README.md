# internal/jobs/

Background jobs run by the worker. Use this folder for cron-like scheduled work and long-running goroutines that aren't a direct response to a queue message.

## Standard job layout

```
internal/jobs/<job>/
├── README.md      # What this job does, schedule, side-effects
├── job.go         # `RunOnce(ctx, ...)` does the actual work
├── schedule.go    # `ScheduleFromConfig(cfg)` returns the trigger spec
└── *_test.go
```

## Wiring

The worker's `main.go` instantiates each job in its own goroutine and uses `time.After` or a cron library to trigger `RunOnce`. Jobs publish to Redis Pub/Sub channels when their results affect other modules (e.g., trigger search re-index).

See [`sample-job/`](./sample-job) for the standard pattern.
