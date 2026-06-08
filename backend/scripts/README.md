# scripts/

Helper shell scripts for development and operations. All use `set -euo pipefail` and auto-detect the project root.

| Script | Purpose |
|--------|---------|
| `run-api.sh` | Starts the API via `go run ./cmd/api`. |
| `run-worker.sh` | Starts the Worker via `go run ./cmd/worker`. |
| `migrate.sh` | Applies SQL migrations using the same config as the API. |
| `deploy-ec2.sh` | **Runs on EC2.** Pulls the ECR image, runs migrations, restarts API/worker, polls `/health/ready`, configures Nginx. |
| `dispatch-ssm-deploy.sh` | **Runs from CI.** Sends the EC2 deployment command through AWS Systems Manager and streams output back. |
| `configure-nginx.sh` | Installs a reverse-proxy config and runs `certbot --nginx` for Let's Encrypt. |
| `secrets-manager-to-env.sh` | Converts an AWS Secrets Manager JSON secret into a Docker env file (chmod 600). |
| `s3-cors-browser-upload.example.json` | Example S3 CORS policy for browser-direct multipart uploads. |
