# deployments/docker/

Optional Dockerfile variants beyond the default `Dockerfile` at repo root. Use this folder for:

- `Dockerfile.standalone` — `output: "standalone"` Next.js builds for the smallest possible image.
- `Dockerfile.dev` — dev-server image used by CI smoke tests.

Reference each variant from CI or `Makefile` targets as needed.
