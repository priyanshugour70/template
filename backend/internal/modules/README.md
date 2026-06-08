# internal/modules/

Domain business modules — the modular monolith. Each module owns one bounded context. This folder is **empty by default**; add modules as you build features.

## Standard module layout

```
internal/modules/<module>/
├── README.md          # What this module owns + its public surface
├── handler.go         # Gin HTTP routes; parses input, calls service, writes response
├── service.go         # Use cases / business rules; orchestrates repo, queue, cache
├── repository.go      # GORM data access; SQL-free public methods
├── model.go           # GORM models + DTOs (request/response shapes)
├── middleware.go      # Per-module middleware (e.g. permission checks) — optional
└── *_test.go          # Table-driven tests for service / handler
```

## Rules

- **Handlers** parse and validate input, call services, and write JSON responses with `pkg/response`. **No business logic, no DB queries.**
- **Services** are the only place business rules live. They consume `Repository`, `Cache`, `Producer`, and client interfaces.
- **Repositories** wrap GORM. Public methods take `context.Context` and return typed values or `*errors.AppError`. Never return `gorm.ErrRecordNotFound` directly — wrap with `errors.New(errors.CodeNotFound, …)`.
- **Models** live with the module that owns them. If two modules need the same model, that's a sign one should call the other through its service interface.
- **Cross-module calls** go through service interfaces, never through repositories. `bootstrap/` wires the concrete service into the consumer.

## Adding a new module

1. Create `internal/modules/<your-module>/` with the files above.
2. Add a migration in `migrations/postgres/NNN_<your-module>.sql` for the schema.
3. Wire it into `internal/bootstrap/bootstrap.go::registerModules` (repo → service → handler).
4. Add tests in `<your-module>_test.go`.
