## 4.1. Readiness probe

The Deployment runs `ping-app` and `log-app` in one Pod. The images are published to Google Artifact Registry in `europe-north1`.

- `ping-app` uses an exec readiness probe: `/server --check-db`. It runs `SELECT 1` against PostgreSQL using `DATABASE_URL`, with a two-second timeout. Failure exits with status `1`.
- `log-app` uses an HTTP readiness probe at `/healthz` on port `8080`.
- Both containers must be ready before the Pod becomes ready and its Service endpoint accepts traffic.

### Deploy without the database

Run these commands from the repository root. Create the namespace, then apply the application configuration, networking, and Deployment. Leave `statefulset.yaml` unapplied during this phase.

```sh
kubectl create namespace exercise-4-1
kubectl apply -n exercise-4-1 \
  -f chapter-5/exercise-4.1/manifests/configmap.yaml \
  -f chapter-5/exercise-4.1/manifests/service.yaml \
  -f chapter-5/exercise-4.1/manifests/gateway.yaml \
  -f chapter-5/exercise-4.1/manifests/route.yaml \
  -f chapter-5/exercise-4.1/manifests/deployment.yaml
kubectl wait -n exercise-4-1 --for=jsonpath='{.status.phase}'=Running \
  pod -l app=ping-app --timeout=120s
sleep 60
kubectl get pods -n exercise-4-1
kubectl get events -n exercise-4-1 --field-selector=reason=Unhealthy
```

Verified on GKE on 25 October 2026: after more than a minute, `log-app` was ready but `ping-app` was not. The Pod remained `1/2` with `Ready=False`, zero restarts, and a Service endpoint with `ready=false`. The exec probe reported that `ping-db-svc` could not be resolved because the database Service did not exist yet.

### Deploy the database and wait for readiness

Apply the PostgreSQL StatefulSet and its headless Service, then wait for the existing application Pod to become ready:

```sh
kubectl apply -n exercise-4-1 \
  -f chapter-5/exercise-4.1/manifests/statefulset.yaml
kubectl wait -n exercise-4-1 --for=condition=Ready \
  pod -l app=ping-app --timeout=120s
kubectl get pods -n exercise-4-1
kubectl exec -n exercise-4-1 deployment/ping-app-dep \
  -c ping-app -- /server --check-db
```

The same application Pod became `2/2 Ready` within 31 seconds of deploying PostgreSQL, with zero restarts. The database Pod was `1/1`, the exec probe succeeded, and the application Service endpoint changed to `ready=true`.

The Gateway routes `/pingpong` to `ping-app` on port `8081` through `ping-app-svc`.

### Current limitation

The probe checks the database through a separate process. If the running application exhausts its startup retries before PostgreSQL arrives, it keeps using its in-memory counter even after the probe succeeds. Readiness recovery demonstrates database availability; reconnecting the running application's database client requires an additional change.
