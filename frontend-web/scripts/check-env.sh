#!/usr/bin/env bash
# Fails the build when API_URL is missing — same check Amplify enforces.
set -euo pipefail

if [ -z "${API_URL:-}" ]; then
  echo "ERROR: API_URL is required (set it in .env.local for dev or Amplify env for prod)." >&2
  exit 1
fi

echo "Environment OK (API_URL=$API_URL)"
