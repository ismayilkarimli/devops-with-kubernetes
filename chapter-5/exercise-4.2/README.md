# 4.2. The project, step 21

Adds liveness and readiness probes to the Todo project, plus a **Break the app** button to demonstrate recovery through a container restart.

- The backend's `/live` endpoint returns `200` while `healthy` is true and `500` after the button sets it to false.
- The backend's exec readiness probe runs `/server --check-db`, executing `SELECT 1` against PostgreSQL with a two-second timeout.
- The frontend's `/live` and `/healthz` endpoints check the frontend itself. Its homepage checks backend `/live` and displays “Sorry we are having issues :(” when the backend is unavailable or unhealthy.

## Try it on GKE

Open [the deployed application](http://136.69.69.240/) and click **Break the app**. The homepage displays the Sorry message while the backend is unhealthy. Kubernetes restarts the backend container after its liveness probe fails, resetting `healthy` to true. Refresh the page after recovery to see the application again.

Watch the restart count and readiness from a terminal:

```sh
kubectl get pods -n exercise-4-2 -w
```

The liveness probes start 80 seconds after a container starts and run every five seconds, with three failures required for a restart. A liveness failure restarts the container inside the existing Pod.

Verified on GKE on 25 October 2026: the backend reported liveness failures with HTTP `500`, Kubernetes restarted it, its restart count increased to `1`, and all three Pods returned to `1/1 Ready`. The homepage and database check recovered. The application was also checked manually through the public URL.

## Deployment

Images are hosted in Google Artifact Registry. The Kustomize configuration targets namespace `exercise-4-2`:

```sh
kubectl create namespace exercise-4-2
kubectl apply -k chapter-5/exercise-4.2/manifests
```

This learning deployment omits container resource requests and limits and does not apply the inherited quota or LimitRange policies, allowing it to run on the existing small GKE node. Persistent storage requests remain configured. The backup CronJob requires separate bucket access; no new backup bucket permissions were granted for this namespace.
