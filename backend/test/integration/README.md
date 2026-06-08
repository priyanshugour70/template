# test/integration/

Integration tests that hit MariaDB, Redis, and/or external partners. Guarded by the `integration` build tag so `go test ./...` doesn't pick them up by default.

```bash
make test-integration
# or
go test -tags=integration ./test/integration/...
```
