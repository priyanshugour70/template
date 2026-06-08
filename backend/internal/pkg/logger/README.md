# internal/pkg/logger/

Structured logging with Zap and daily file rotation.

- Console encoder for development, JSON encoder for production.
- Daily rotating log files under `logs/YYYY-MM-DD/{purpose}.log` via lumberjack.
- Configurable max file size and backup count.
- Caller info and stack traces (on error+) included automatically.

```go
log, _ := logger.New(cfg.App.Env)                       // default "api" purpose
log, _ := logger.NewWithConfig(cfg.App.Env, logger.FileConfig{Purpose: "worker"})
log, _ := logger.NewNoFile(cfg.App.Env)                 // stderr only (tests)
```
