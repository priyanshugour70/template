# deployments/

Infrastructure deployment configuration.

| Folder | Purpose |
|--------|---------|
| **docker/** | Dockerfiles for API/worker/all-binaries image + `docker-compose.yaml` (local) and `docker-compose.ec2.yaml` (EC2 runtime via ECR image). |
| **k8s/** | Kubernetes Deployment + Service manifests with liveness/readiness probes, resource requests/limits, and secret references. |

See top-level [`DEPLOYMENT.md`](../DEPLOYMENT.md) for the end-to-end deploy flow.
