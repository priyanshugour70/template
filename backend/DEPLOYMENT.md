# Deployment

This document describes the end-to-end deployment flow used by this template. Adapt names and AWS resource IDs to your environment.

## Environments

| Branch | Environment | EC2 instance ID secret | Domain | AWS secret |
|---|---|---|---|---|
| `staging` | Staging | `STAGING_EC2_INSTANCE_ID` | `https://staging-api.example.com` | `your-org/your-service/staging` |
| `main` | Production | `PROD_EC2_INSTANCE_ID` | `https://api.example.com` | `your-org/your-service/prod` |

Pushes to any other branch do not deploy.

## AWS resources

### ECR

Docker images are pushed to a private ECR repository:

```text
<aws-account>.dkr.ecr.<region>.amazonaws.com/your-org/your-service:<git-sha>
```

The repository uses immutable tags. The workflow reuses an existing Git-SHA tag instead of pushing twice.

### Secrets Manager

Runtime environment variables come from AWS Secrets Manager, not committed `.env` files.

Required secrets:

```text
your-org/your-service/staging
your-org/your-service/prod
```

Each secret should contain the keys listed in `.env.example` (PostgreSQL credentials, Redis address, JWT secret, SMTP, S3, etc.). Store values raw (no quotes). Passwords containing `$` are supported by the EC2 compose file through `env_file.format: raw`.

### IAM

GitHub Actions uses an IAM user via GitHub environment secrets:

```text
AWS_ACCESS_KEY_ID
AWS_SECRET_ACCESS_KEY
AWS_ACCOUNT_ID
AWS_REGION
ECR_REPOSITORY
STAGING_EC2_INSTANCE_ID
PROD_EC2_INSTANCE_ID
CERTBOT_EMAIL
```

The EC2 instances use an EC2 IAM role that must allow:

```text
AmazonSSMManagedInstanceCore
secretsmanager:GetSecretValue   # for staging/prod secrets
ECR image pull
S3 read/write                   # if your service uploads assets
```

### PostgreSQL

PostgreSQL is **not** run inside the EC2 compose stack. Use Amazon RDS (or any managed Postgres) and put the connection details in the Secrets Manager secret (`POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DATABASE`, `POSTGRES_SSLMODE`).

Migrations run inside the api container during deploy via:

```bash
docker compose run --rm --no-deps api ./scripts/migrate.sh
```

## GitHub Actions flow

Workflow file:

```text
.github/workflows/backend-deploy.yml
```

On push to `staging` or `main`, GitHub Actions:

1. Checks out the repository.
2. Sets up Go (matching `go.mod`).
3. Runs `go test ./...`.
4. Logs in to Amazon ECR.
5. Builds the runtime image using `deployments/docker/Dockerfile.app`.
6. Pushes or reuses the ECR image tagged with the Git SHA.
7. Selects the target EC2 instance based on branch.
8. Sends a deploy command through AWS Systems Manager — no SSH key required.

## EC2 deploy flow

CI calls `scripts/dispatch-ssm-deploy.sh`, which sends a remote shell command through SSM. On EC2 it:

1. Installs missing prerequisites (`awscli`, `jq`, `curl`, `docker.io`, Docker Compose plugin, `nginx`, `certbot`).
2. Writes deployment files into `/opt/your-org/your-service`.
3. Runs `scripts/deploy-ec2.sh`.

The EC2 deploy script:

1. Fetches the Secrets Manager secret into `runtime.env`.
2. Logs in to ECR and pulls the image.
3. Starts Redis.
4. Runs `scripts/migrate.sh` **inside the api container** (so connection details from `runtime.env` are honoured and `psql` is on PATH inside the image).
5. Starts API + worker.
6. Polls `http://127.0.0.1:8080/health/ready`.
7. Configures Nginx and TLS for the environment domain.

## Docker runtime

EC2 uses `deployments/docker/docker-compose.ec2.yaml` with services: `redis`, `api`, `worker`. The same ECR image (built from `Dockerfile.app`) contains both binaries plus the migration script and SQL files.

## Nginx and TLS

`scripts/configure-nginx.sh` installs a reverse-proxy config and runs `certbot --nginx` for Let's Encrypt. Before TLS can issue:

1. DNS A records must point to the EC2 public IP.
2. Security groups must allow `80` and `443` inbound.
3. Certbot must be able to validate via HTTP-01.

## Common failures

- **SSM `Failed` with empty output** — open *Systems Manager → Run Command*, find the invocation, and check for: missing IAM permissions on Secrets Manager, ECR pull failure, migration error, or a `/health/ready` timeout.
- **SSM `Undeliverable`** — instance not reachable. Check Fleet Manager **Ping status**, instance ID, region, IAM profile, SSM agent, outbound 443.
- **`AccessDeniedException` (Secrets Manager)** — EC2 role lacks `secretsmanager:GetSecretValue`.
- **ECR tag already exists** — expected with immutable tags; workflow reuses existing Git-SHA tags.
- **PostgreSQL connection refused** — RDS security group must allow inbound from the EC2 SG.
- **Docker Compose `$…` warning** — secrets containing `$` need `env_file.format: raw`.

## Manual rerun

```text
GitHub → Actions → Deploy Backend → Re-run jobs
```

To deploy a new commit, push to `staging` or `main`.
