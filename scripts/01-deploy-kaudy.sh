#!/usr/bin/env bash

set -e

# Turn colors in this script off by setting the NO_COLOR variable in your
# environment to any value:
#
# $ NO_COLOR=1 ./scripts/01-deploy-kaudy.sh
NO_COLOR=${NO_COLOR:-""}
if [ -z "$NO_COLOR" ]; then
  header=$'\e[1;33m'
  reset=$'\e[0m'
else
  header=''
  reset=''
fi

function header_text {
  echo "$header$*$reset"
}

IMAGE=${IMAGE:-"quay.io/matzew/kaudy:latest"}

header_text "Building kaudy container image"
podman build -t "${IMAGE}" .

header_text "Loading image into kind"
kind load docker-image "${IMAGE}"

if [ -z "${ANTHROPIC_API_KEY}" ]; then
    echo "Error: ANTHROPIC_API_KEY environment variable is not set"
    exit 1
fi

header_text "Creating kaudy-env secret"
kubectl delete secret kaudy-env --ignore-not-found
kubectl create secret generic kaudy-env --from-literal=ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY}"

header_text "Deploying kaudy pod"
kubectl apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: kaudy
  labels:
    app: kaudy
spec:
  containers:
  - name: kaudy
    image: ${IMAGE}
    imagePullPolicy: Never
    envFrom:
    - secretRef:
        name: kaudy-env
    workingDir: /workspace
    volumeMounts:
    - name: workspace
      mountPath: /workspace
    stdin: true
    tty: true
  volumes:
  - name: workspace
    emptyDir: {}
EOF

header_text "Waiting for kaudy pod to be ready"
kubectl wait --for=condition=Ready pod/kaudy --timeout=120s

header_text "kaudy pod is running!"
kubectl get pods -l app=kaudy

echo ""
echo "To exec into the kaudy pod:"
echo "  kubectl exec -it kaudy -- claude --dangerously-skip-permissions"
