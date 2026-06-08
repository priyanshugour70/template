#!/usr/bin/env bash
# Runs from CI. Sends the EC2 deployment command through AWS Systems Manager.
# Streams remote stdout/stderr while the SSM command executes so GitHub Actions
# logs show progress in near-real-time.
set -euo pipefail

ENVIRONMENT="${ENVIRONMENT:?ENVIRONMENT is required: staging or prod}"
INSTANCE_ID="${INSTANCE_ID:?INSTANCE_ID is required}"
AWS_REGION="${AWS_REGION:?AWS_REGION is required}"
AWS_ACCOUNT_ID="${AWS_ACCOUNT_ID:?AWS_ACCOUNT_ID is required}"
ECR_REPOSITORY="${ECR_REPOSITORY:?ECR_REPOSITORY is required}"
IMAGE_TAG="${IMAGE_TAG:?IMAGE_TAG is required}"
CERTBOT_EMAIL="${CERTBOT_EMAIL:-}"

BACKEND_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_DIR="${APP_DIR:-/opt/your-org/your-service}"

LAST_STDOUT_LEN=0
LAST_STDERR_LEN=0

stream_ssm_output() {
  local invocation_json="$1"
  local stdout stderr stdout_len stderr_len new_stdout new_stderr

  stdout="$(jq -r '.StandardOutputContent // ""' "$invocation_json")"
  stderr="$(jq -r '.StandardErrorContent // ""' "$invocation_json")"
  stdout_len="${#stdout}"
  stderr_len="${#stderr}"

  if [ "$stdout_len" -gt "$LAST_STDOUT_LEN" ]; then
    new_stdout="${stdout:$LAST_STDOUT_LEN}"
    if [ -n "$new_stdout" ]; then
      printf '%s' "$new_stdout"
      if [ "${new_stdout: -1}" != $'\n' ]; then
        printf '\n'
      fi
    fi
    LAST_STDOUT_LEN="$stdout_len"
  fi
  if [ "$stderr_len" -gt "$LAST_STDERR_LEN" ]; then
    new_stderr="${stderr:$LAST_STDERR_LEN}"
    if [ -n "$new_stderr" ]; then
      echo "--- remote stderr (live) ---" >&2
      printf '%s' "$new_stderr" >&2
      if [ "${new_stderr: -1}" != $'\n' ]; then
        printf '\n' >&2
      fi
    fi
    LAST_STDERR_LEN="$stderr_len"
  fi
}

# Optional: pre-flight checks against the EC2 instance and IAM role.
preflight_ssm_target() {
  echo "==> preflight: EC2 instance $INSTANCE_ID in $AWS_REGION ($ENVIRONMENT)"
  local ec2_state ping_status
  ec2_state="$(aws ec2 describe-instances --region "$AWS_REGION" --instance-ids "$INSTANCE_ID" \
    --query "Reservations[0].Instances[0].State.Name" --output text 2>/dev/null || echo Unknown)"
  if [ "$ec2_state" != "running" ]; then
    echo "ERROR: EC2 instance is '$ec2_state', not running." >&2
    exit 1
  fi
  ping_status="$(aws ssm describe-instance-information --region "$AWS_REGION" \
    --filters "Key=InstanceIds,Values=$INSTANCE_ID" \
    --query "InstanceInformationList[0].PingStatus" --output text 2>/dev/null || echo Unknown)"
  if [ "$ping_status" != "Online" ]; then
    echo "ERROR: SSM agent is not Online (status: ${ping_status})." >&2
    echo "Install/start amazon-ssm-agent and verify IAM AmazonSSMManagedInstanceCore is attached." >&2
    exit 1
  fi
}

preflight_ssm_target

compose_b64="$(base64 < "$BACKEND_ROOT/deployments/docker/docker-compose.ec2.yaml" | tr -d '\n')"
env_script_b64="$(base64 < "$BACKEND_ROOT/scripts/secrets-manager-to-env.sh" | tr -d '\n')"
nginx_script_b64="$(base64 < "$BACKEND_ROOT/scripts/configure-nginx.sh" | tr -d '\n')"
deploy_script_b64="$(base64 < "$BACKEND_ROOT/scripts/deploy-ec2.sh" | tr -d '\n')"

remote_script="$(mktemp)"
parameters_file="$(mktemp)"
trap 'rm -f "$remote_script" "$parameters_file" "${invocation_json:-}"' EXIT

cat > "$remote_script" <<REMOTE
set -euo pipefail
APP_DIR="$APP_DIR"
if [ "\$(id -u)" -eq 0 ]; then SUDO=""; else SUDO="sudo"; fi

install_prerequisites() {
  missing=""
  for bin in aws docker jq curl nginx; do
    if ! command -v "\$bin" >/dev/null 2>&1; then missing="\$missing \$bin"; fi
  done
  if [ -n "\$missing" ]; then
    if command -v apt-get >/dev/null 2>&1; then
      \$SUDO apt-get update
      \$SUDO env DEBIAN_FRONTEND=noninteractive apt-get install -y awscli jq curl docker.io ca-certificates nginx
      if ! command -v certbot >/dev/null 2>&1; then
        \$SUDO env DEBIAN_FRONTEND=noninteractive apt-get install -y certbot python3-certbot-nginx || true
      fi
    else
      echo "Missing tools and apt-get unavailable:\$missing" >&2; exit 1
    fi
  fi
  if ! docker compose version >/dev/null 2>&1; then
    arch="\$(uname -m)"
    case "\$arch" in x86_64|amd64) arch=x86_64;; aarch64|arm64) arch=aarch64;; esac
    \$SUDO mkdir -p /usr/local/lib/docker/cli-plugins
    \$SUDO curl -fsSL "https://github.com/docker/compose/releases/download/v2.32.4/docker-compose-linux-\$arch" \\
      -o /usr/local/lib/docker/cli-plugins/docker-compose
    \$SUDO chmod +x /usr/local/lib/docker/cli-plugins/docker-compose
  fi
  \$SUDO systemctl enable --now docker >/dev/null 2>&1 || \$SUDO service docker start >/dev/null 2>&1 || true
}

echo "==> install prerequisites"
install_prerequisites

echo "==> sync deploy files to \$APP_DIR"
\$SUDO mkdir -p "\$APP_DIR/deployments/docker" "\$APP_DIR/scripts"
printf '%s' '$compose_b64'      | base64 -d | \$SUDO tee "\$APP_DIR/deployments/docker/docker-compose.ec2.yaml" >/dev/null
printf '%s' '$env_script_b64'   | base64 -d | \$SUDO tee "\$APP_DIR/scripts/secrets-manager-to-env.sh" >/dev/null
printf '%s' '$nginx_script_b64' | base64 -d | \$SUDO tee "\$APP_DIR/scripts/configure-nginx.sh" >/dev/null
printf '%s' '$deploy_script_b64'| base64 -d | \$SUDO tee "\$APP_DIR/scripts/deploy-ec2.sh" >/dev/null
\$SUDO chmod 755 "\$APP_DIR/scripts/"*.sh

echo "==> run deploy-ec2.sh (ENVIRONMENT=$ENVIRONMENT IMAGE_TAG=$IMAGE_TAG)"
\$SUDO env \\
  ENVIRONMENT='$ENVIRONMENT' \\
  AWS_REGION='$AWS_REGION' \\
  AWS_ACCOUNT_ID='$AWS_ACCOUNT_ID' \\
  ECR_REPOSITORY='$ECR_REPOSITORY' \\
  IMAGE_TAG='$IMAGE_TAG' \\
  CERTBOT_EMAIL='$CERTBOT_EMAIL' \\
  APP_DIR="\$APP_DIR" \\
  bash "\$APP_DIR/scripts/deploy-ec2.sh"
REMOTE

jq -Rs '{commands: [.], executionTimeout: ["1800"]}' "$remote_script" > "$parameters_file"

command_id="$(
  aws ssm send-command \
    --region "$AWS_REGION" \
    --document-name "AWS-RunShellScript" \
    --instance-ids "$INSTANCE_ID" \
    --comment "Deploy backend $IMAGE_TAG to $ENVIRONMENT" \
    --parameters "file://$parameters_file" \
    --query "Command.CommandId" \
    --output text
)"
echo "SSM command id: $command_id"

invocation_json="$(mktemp)"
deadline=$((SECONDS + 1800))
status="Pending"
echo "==> streaming EC2 output"
while [ "$SECONDS" -lt "$deadline" ]; do
  aws ssm get-command-invocation \
    --region "$AWS_REGION" --command-id "$command_id" --instance-id "$INSTANCE_ID" \
    --output json > "$invocation_json" 2>/dev/null || true
  status="$(jq -r '.Status // "Pending"' "$invocation_json")"
  stream_ssm_output "$invocation_json"
  case "$status" in Success|Failed|Cancelled|TimedOut|Cancelling) break ;; *) sleep 3 ;; esac
done

aws ssm get-command-invocation --region "$AWS_REGION" --command-id "$command_id" --instance-id "$INSTANCE_ID" \
  --output json > "$invocation_json"
status="$(jq -r '.Status // "Unknown"' "$invocation_json")"
echo "SSM status: $status"
echo "--- remote stdout (final) ---"
jq -r '.StandardOutputContent // ""' "$invocation_json"
echo "--- remote stderr (final) ---"
jq -r '.StandardErrorContent // ""' "$invocation_json"

if [ "$status" != "Success" ]; then
  echo "Deploy failed. AWS Console: Systems Manager → Run Command → $command_id on $INSTANCE_ID" >&2
  exit 1
fi
