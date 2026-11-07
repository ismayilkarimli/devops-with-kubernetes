## 4.3. Prometheus

Prometheus is installed with the `prometheus-community/prometheus` Helm chart, which also deploys kube-state-metrics. A sample StatefulSet with two replicas runs in the `prometheus` namespace so there are StatefulSet pods to count.

Run these commands from `chapter-5/exercise-4.3`.

### Install Prometheus

```sh
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm install prometheus prometheus-community/prometheus \
  -n prometheus --create-namespace \
  -f observability/prometheus-values.yml
```

### Deploy the sample StatefulSet

```sh
kubectl apply -k manifests
kubectl -n prometheus get pods
```

The `hello-world` image exits right after starting, so its pods restart repeatedly. They still exist and are reported by kube-state-metrics.

### Query

Forward the Prometheus server and open http://localhost:9090:

```sh
kubectl -n prometheus port-forward svc/prometheus-server 9090:80
```

Number of pods created by StatefulSets in the `prometheus` namespace:

```promql
count(kube_pod_info{created_by_kind="StatefulSet", namespace="prometheus"})
```
