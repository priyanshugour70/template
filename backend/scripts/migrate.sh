#!/usr/bin/env bash
# Applies SQL migrations using the same .env / config as the API (RDS, Docker, or local).
# Requires: run from repo root (or: make migrate).
set -euo pipefail
cd "$(dirname "$0")/.."
exec go run ./cmd/migrate
