# deployments/docker/

Docker configurations for local development and CI/CD.

| File | Purpose |
|------|---------|
| `docker-compose.yaml` | Local dev: Redis 7 + API + Worker (MariaDB external). |
| `docker-compose.ec2.yaml` | EC2 runtime stack using an ECR image. Services: `redis`, `api`, `worker`, `migrate`. |
| `Dockerfile.api` | Multi-stage build for the API binary alone (~5 MB). |
| `Dockerfile.worker` | Multi-stage build for the Worker binary alone. |
| `Dockerfile.app` | Multi-stage build that compiles **all** `cmd/*` binaries into one image. Used on EC2 so the same artifact runs API + worker + migrate. |

```bash
make docker-up    # starts redis + api + worker
make docker-down
```

MariaDB, Meilisearch, S3, and other secrets come from `../../.env`; compose only overrides `REDIS_ADDR` for the in-network Redis hostname.
