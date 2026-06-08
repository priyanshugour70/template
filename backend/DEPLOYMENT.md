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

The repository uses immutable tags. The workflow checks whether the Git SHA tag already exists and reuses it instead of pushing twice.

### Secrets Manager

Runtime environment variables come from AWS Secrets Manager, not committed `.env` files.

Required secrets:

```text
your-org/your-service/staging
your-org/your-service/prod
```

Each secret should contain the keys listed in `.env.example`. Store secret values raw (no quotes). Passwords containing `$` are supported by the EC2 compose file through `env_file.format: raw`.

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

Both EC2 instances must appear online in:

```text
AWS Systems Manager → Fleet Manager
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

1. Installs missing prerequisites (`awscli`, `jq`, `curl`, `docker.io`, Docker Compose plugin, `nginx`, `certbot`, `python3-certbot-nginx`).
2. Writes deployment files into `/opt/your-org/your-service`.
3. Runs `scripts/deploy-ec2.sh`.

The EC2 deploy script:

1. Selects environment-specific values (secret/domain/project name).
2. Fetches the Secrets Manager secret into `runtime.env`.
3. Logs in to ECR and pulls the image.
4. Starts Redis.
5. Runs database migrations (`docker compose run --rm migrate`).
6. Restarts API and worker.
7. Polls `http://127.0.0.1:8080/health/ready`.
8. Configures Nginx and TLS for the environment domain.

## Docker runtime

EC2 uses `deployments/docker/docker-compose.ec2.yaml` with services: `redis`, `api`, `worker`, `migrate`.

The same ECR image contains all binaries (built from `Dockerfile.app` which compiles every `cmd/*`).

Optional integrations (Meilisearch, S3) are external; configure via env.

## Nginx and TLS

`scripts/configure-nginx.sh` installs a reverse-proxy config and runs `certbot --nginx` for Let's Encrypt. Before TLS can issue:

1. DNS A records must point to the EC2 public IP.
2. Security groups must allow `80` and `443` inbound.
3. Certbot must be able to validate via HTTP-01.

## Common failures

- **SSM `Failed` with empty output** — open *Systems Manager → Run Command*, find the invocation, and look for: missing IAM permissions on Secrets Manager, ECR pull failure, migration error, or a `/health/ready` timeout.
- **SSM `Undeliverable`** — instance is not reachable. Check Fleet Manager **Ping status**, instance ID, region, IAM profile, SSM agent, and outbound 443.
- **`AccessDeniedException` (Secrets Manager)** — EC2 role lacks `secretsmanager:GetSecretValue` for the secret.
- **ECR tag already exists** — expected with immutable tags; workflow reuses existing Git-SHA tags.
- **MariaDB access denied** — credentials or grants in Secrets Manager are wrong.
- **Docker Compose `$…` warning** — secrets containing `$` need `env_file.format: raw`.

## Manual rerun

```text
GitHub → Actions → Deploy Backend → Re-run jobs
```

To deploy a new commit, push to `staging` or `main`.
