#!/usr/bin/env bash
# Runs on EC2 (invoked via SSM by dispatch-ssm-deploy.sh).
# Pulls the ECR image, applies migrations inside the api container, restarts API/worker,
# polls /health/ready, then configures Nginx.
set -euo pipefail

ENVIRONMENT="${ENVIRONMENT:?ENVIRONMENT is required: staging or prod}"
AWS_REGION="${AWS_REGION:?AWS_REGION is required}"
AWS_ACCOUNT_ID="${AWS_ACCOUNT_ID:?AWS_ACCOUNT_ID is required}"
ECR_REPOSITORY="${ECR_REPOSITORY:?ECR_REPOSITORY is required}"
IMAGE_TAG="${IMAGE_TAG:?IMAGE_TAG is required}"

APP_DIR="${APP_DIR:-/opt/your-org/your-service}"
COMPOSE_FILE="${COMPOSE_FILE:-$APP_DIR/deployments/docker/docker-compose.ec2.yaml}"
RUNTIME_ENV_FILE="${RUNTIME_ENV_FILE:-$APP_DIR/runtime.env}"
CERTBOT_EMAIL="${CERTBOT_EMAIL:-admin@example.com}"

case "$ENVIRONMENT" in
  staging)
    APP_ENV="${APP_ENV:-staging}"
    SECRET_ID="${SECRET_ID:-your-org/your-service/staging}"
    PROJECT_NAME="${PROJECT_NAME:-your-service-staging}"
    API_DOMAIN="${API_DOMAIN:-staging-api.example.com}"
    ;;
  prod|production)
    APP_ENV="${APP_ENV:-production}"
    SECRET_ID="${SECRET_ID:-your-org/your-service/prod}"
    PROJECT_NAME="${PROJECT_NAME:-your-service-prod}"
    API_DOMAIN="${API_DOMAIN:-api.example.com}"
    ;;
  *)
    echo "Unsupported ENVIRONMENT: $ENVIRONMENT" >&2
    exit 1
    ;;
esac

for bin in aws docker jq curl; do
  if ! command -v "$bin" >/dev/null 2>&1; then
    echo "$bin is required on this EC2 instance" >&2
    exit 1
  fi
done
if ! docker compose version >/dev/null 2>&1; then
  echo "docker compose plugin is required on this EC2 instance" >&2
  exit 1
fi

mkdir -p "$APP_DIR"
cd "$APP_DIR"

echo "==> loading secrets from $SECRET_ID"
bash "$APP_DIR/scripts/secrets-manager-to-env.sh" "$SECRET_ID" "$RUNTIME_ENV_FILE" "$AWS_REGION"

SERVER_PORT="$(awk -F= '$1 == "SERVER_PORT" { print $2 }' "$RUNTIME_ENV_FILE" | tail -n 1)"
SERVER_PORT="${SERVER_PORT:-8080}"
HOST_API_PORT="${HOST_API_PORT:-$SERVER_PORT}"

export APP_ENV
export HOST_API_PORT
export IMAGE_URI="$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/$ECR_REPOSITORY:$IMAGE_TAG"
export RUNTIME_ENV_FILE
export SERVER_PORT

compose() {
  docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" "$@"
}

echo "==> logging in to ECR ($ECR_REPOSITORY:$IMAGE_TAG)"
aws ecr get-login-password --region "$AWS_REGION" \
  | docker login --username AWS --password-stdin "$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com"

echo "==> pulling images"
compose pull

echo "==> starting redis"
compose up -d redis

echo "==> applying migrations (psql against managed Postgres)"
if ! compose run --rm --no-deps --entrypoint "" api ./scripts/migrate.sh; then
  echo "Migrations failed. Recent logs:" >&2
  compose logs --tail=80 api >&2 || true
  exit 1
fi

echo "==> starting api and worker"
compose up -d --remove-orphans api worker

health_url="http://127.0.0.1:${HOST_API_PORT}/health/ready"
for _ in $(seq 1 36); do
  if curl -fsS "$health_url" >/dev/null; then
    if [ -n "${API_DOMAIN:-}" ]; then
      DOMAIN="$API_DOMAIN" \
        UPSTREAM_PORT="$HOST_API_PORT" \
        CERTBOT_EMAIL="$CERTBOT_EMAIL" \
        SITE_NAME="your-service-backend" \
        bash "$APP_DIR/scripts/configure-nginx.sh"
    fi
    echo "Deployment healthy: $health_url"
    exit 0
  fi
  sleep 5
done

echo "API did not become ready: $health_url" >&2
compose logs --tail=120 api >&2 || true
exit 1
