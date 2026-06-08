#!/usr/bin/env bash
set -euo pipefail

SECRET_ID="${1:?usage: secrets-manager-to-env.sh <secret-id> <output-file> [aws-region]}"
OUTPUT_FILE="${2:?usage: secrets-manager-to-env.sh <secret-id> <output-file> [aws-region]}"
AWS_REGION="${3:-${AWS_REGION:-us-east-1}}"

tmp_secret="$(mktemp)"
tmp_env="$(mktemp)"
trap 'rm -f "$tmp_secret" "$tmp_env"' EXIT

aws secretsmanager get-secret-value \
  --secret-id "$SECRET_ID" \
  --region "$AWS_REGION" \
  --query SecretString \
  --output text > "$tmp_secret"

if jq -e 'type == "object"' "$tmp_secret" >/dev/null 2>&1; then
  jq -r '
    to_entries[]
    | select(.value != null)
    | select(.key | test("^[A-Za-z_][A-Za-z0-9_]*$"))
    | "\(.key)=\(.value | tostring | gsub("\r"; "") | gsub("\n"; "\\n"))"
  ' "$tmp_secret" > "$tmp_env"
else
  cp "$tmp_secret" "$tmp_env"
fi

install -m 600 "$tmp_env" "$OUTPUT_FILE"
