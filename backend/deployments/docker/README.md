# deployments/docker/

Docker configurations for local development and CI/CD.

| File | Purpose |
|------|---------|
| `docker-compose.yaml` | Local dev: PostgreSQL 16 + Redis 7 + API + Worker. |
| `docker-compose.ec2.yaml` | EC2 runtime stack using an ECR image. Services: `redis`, `api`, `worker`. (Postgres is RDS-managed in production.) |
| `Dockerfile.api` | Multi-stage build for the API binary alone (~5 MB). |
| `Dockerfile.worker` | Multi-stage build for the Worker binary alone. |
| `Dockerfile.app` | Multi-stage build that compiles **both** `cmd/api` and `cmd/worker` into one image. Bundles `migrations/` and `scripts/migrate.sh` so migrations can be applied via `docker compose exec api ./scripts/migrate.sh`. Used on EC2 so a single artifact runs API + worker. |

```bash
make docker-up    # starts postgres + redis + api + worker
make docker-down
```

Migrations are applied via `scripts/migrate.sh` (which shells out to `psql`) — there's no separate `migrate` binary. Run it inside the api container after deploy:

```bash
docker compose exec api ./scripts/migrate.sh
```
