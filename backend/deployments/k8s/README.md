# deployments/k8s/

Kubernetes deployment manifests.

| File | Purpose |
|------|---------|
| `api-deployment.yaml` | API Deployment (2 replicas) + ClusterIP Service with liveness/readiness probes. |
| `worker-deployment.yaml` | Worker Deployment (1 replica) for background job processing. |

Both reference a `Secret` named `app-secrets` for environment configuration. Create it with the keys from `.env.example`:

```bash
kubectl create secret generic app-secrets --from-env-file=./.env
```

For ingress, set up your cluster's ingress controller and create an `Ingress` resource pointing at the `api` Service.
