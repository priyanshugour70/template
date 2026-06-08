# internal/jobs/sample-job/

Reference scheduled job. Runs `RunOnce` on an interval driven by `cfg.SampleJob.IntervalMinutes` (TODO: add to config).

Replace the body of `RunOnce` with your job's work (sync external prices, refresh caches, reconcile data, …).
