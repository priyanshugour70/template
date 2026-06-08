# internal/config/

Application configuration loading and business-config interfaces.

| File | Purpose |
|------|---------|
| `config.go` | Static configuration via Viper/godotenv — structs for App, Server, MariaDB, Redis, SMTP, CORS, Auth, OAuth, Assets, Meilisearch, OTEL, and one sample external partner. Loads `.env` → `.env.{APP_ENV}` → `.env.local`. Production validation enforces JWT secret length and CORS origins. |
| `biz_config.go` | `BizConfigReader` interface — typed, read-only access to dynamic business config stored in the database. Modules depend on this interface, not a concrete `bizconfig.Service`. Add typed accessors here when a module needs runtime-mutable config. |

## Adding new configuration

**Static (env-based):** Add a field to the appropriate struct in `config.go`, bind via the `mapstructure` tag, set a default in `Load()`. Add an env line to `.env.example`.

**Dynamic (DB-based):** Add a key constant + typed accessor method to `BizConfigReader` here. Implement it in your `bizconfig.Service` and surface it through a thin admin UI / API.

## Dependency rules

- `config` package has **no imports from `internal/modules/`** — it defines interfaces that modules implement.
- All modules may import `config` for static config structs and the `BizConfigReader` interface.
