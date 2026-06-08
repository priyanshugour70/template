# test/

Integration and end-to-end tests.

| Folder | Purpose |
|--------|---------|
| `unit/` | Standalone unit tests that don't need network/DB. Run with `go test ./...`. (Module-local `*_test.go` files live next to the code they test — this folder is for tests that span multiple modules.) |
| `integration/` | Tests that hit PostgreSQL, Redis, or external partners. Guarded by `//go:build integration` so they don't run on every `go test`. Trigger with `make test-integration`. |

## Integration test pattern

```go
//go:build integration

package integration

import "testing"

func TestSomething(t *testing.T) { /* … */ }
```

Run integration tests:

```bash
make test-integration            # all
go test -tags=integration ./test/integration/...
```

Provide a docker-compose snippet (or `testcontainers-go`) so the tests can spin up their own DB/Redis if devs run them locally.
