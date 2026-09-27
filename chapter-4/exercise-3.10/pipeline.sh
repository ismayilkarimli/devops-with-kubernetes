#!/usr/bin/env bash
set -euo pipefail

: "${PROJECT_ID:?PROJECT_ID is required}"
: "${PROJECT_NUMBER:?PROJECT_NUMBER is required}"
: "${REGISTRY:?REGISTRY is required}"
: "${REPOSITORY:?REPOSITORY is required}"
: "${GITHUB_SHA:?GITHUB_SHA is required}"
: "${NAMESPACE:?NAMESPACE is required}"

exercise_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
app_image_tag="$REGISTRY/$PROJECT_ID/$REPOSITORY/todo-app:$GITHUB_SHA"
server_image_tag="$REGISTRY/$PROJECT_ID/$REPOSITORY/todo-server:$GITHUB_SHA"
storage_bucket="dwk-gke-storage"
storage_ksa="dwk-storage-sa"
app_deployment="todo-app-dep"
server_deployment="todo-app-server-dep"

echo "Building and publishing images for $exercise_dir"
docker build --platform linux/amd64 --tag "$app_image_tag" "$exercise_dir/todo-app"
docker build --platform linux/amd64 --tag "$server_image_tag" "$exercise_dir/todo-server"
docker push "$app_image_tag"
docker push "$server_image_tag"

echo "Deploying $exercise_dir to namespace $NAMESPACE"
kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
kubectl wait --for=jsonpath='{.status.phase}'=Active "namespace/$NAMESPACE" --timeout=60s
kubectl config set-context --current --namespace="$NAMESPACE"

cd "$exercise_dir/manifests"
kustomize edit set namespace "$NAMESPACE"
kustomize edit set image \
  APP/IMAGE="$app_image_tag" \
  SERVER/IMAGE="$server_image_tag"
kustomize build . | kubectl apply -f -
kubectl get serviceaccount "$storage_ksa"

storage_principal="principal://iam.googleapis.com/projects/${PROJECT_NUMBER}/locations/global/workloadIdentityPools/${PROJECT_ID}.svc.id.goog/subject/ns/${NAMESPACE}/sa/${storage_ksa}"
gcloud storage buckets add-iam-policy-binding "gs://${storage_bucket}" \
  --role=roles/storage.objectUser \
  --member="$storage_principal" \
  --condition=None

kubectl rollout status deployment "$app_deployment" --timeout=5m
kubectl rollout status deployment "$server_deployment" --timeout=5m
kubectl get services -o wide
