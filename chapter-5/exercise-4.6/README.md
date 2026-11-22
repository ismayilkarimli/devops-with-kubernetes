## 4.6. The project, step 23

The Todo backend publishes `todo_save` after creating a todo and `todo_update` after marking one done. Publishing follows the database write. The backend responds without waiting for the HTTP notification.

Six Go broadcaster Pods subscribe to both subjects in the `broadcasters` queue group. Core NATS selects one connected group member per message. That Pod makes one HTTP POST attempt. Events can be lost if NATS, the broadcaster, or the webhook is unavailable.

The NATS Service gives the backend and broadcasters a stable address. The broadcasters connect to NATS and send HTTP requests out to `WEBHOOK_URL`, so their Deployment needs neither a Service nor an Ingress.

## Install NATS and enable metrics

From the repository root:

```sh
helm repo add nats https://nats-io.github.io/k8s/helm/charts/
helm upgrade --install my-nats nats/nats -n nats --create-namespace \
  -f chapter-5/exercise-4.6/observability/nats-values.yaml
```

Apply the exercise's `observability/prometheus-values.yaml` to your Prometheus Helm release. Its `nats` scrape job targets the exporter on port 7777. In Prometheus, query `up{job="nats"}`. A value of `1` confirms a successful scrape.

## Try it

Build and publish the images using the exercise pipeline, or publish the three `4.6` tags referenced by Kustomize before applying the manifests. Set `WEBHOOK_URL` in the broadcaster Deployment if you want to use a different receiver. The default, `https://httpbin.org/post`, echoes the body in the response. Use the broadcaster logs to see which Pod forwarded each event and what it sent:

```sh
kubectl logs -n exercise-4-6 -l app=broadcaster --prefix --tail=30
```

Create a todo and mark it done. Look for one successful forward on each subject, with these bodies respectively:

```json
{"user":"bot","message":"A todo was created"}
{"user":"bot","message":"A todo was marked as done"}
```
