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
SKILL_IMAGE=${SKILL_IMAGE:-"quay.io/matzew/agent-skills"}
CLOUD_ML_REGION=${CLOUD_ML_REGION:-"us-east5"}

# Vertex AI project is required
if [ -z "${ANTHROPIC_VERTEX_PROJECT_ID}" ]; then
    echo "Error: ANTHROPIC_VERTEX_PROJECT_ID environment variable is not set"
    echo "  export ANTHROPIC_VERTEX_PROJECT_ID='your-gcp-project-id'"
    exit 1
fi

# Resolve GCP credentials: prefer explicit SA key file, fall back to ADC
CREDENTIALS_FILE=""
if [ -n "${VERTEX_SA_KEY_FILE}" ]; then
    if [ ! -f "${VERTEX_SA_KEY_FILE}" ]; then
        echo "Error: VERTEX_SA_KEY_FILE points to a file that does not exist: ${VERTEX_SA_KEY_FILE}"
        exit 1
    fi
    CREDENTIALS_FILE="${VERTEX_SA_KEY_FILE}"
else
    ADC_FILE="${HOME}/.config/gcloud/application_default_credentials.json"
    if [ -f "${ADC_FILE}" ]; then
        CREDENTIALS_FILE="${ADC_FILE}"
        echo "Using application default credentials from ${ADC_FILE}"
    else
        echo "Error: No GCP credentials found."
        echo "  Either set VERTEX_SA_KEY_FILE to a service account key JSON file:"
        echo "    export VERTEX_SA_KEY_FILE='/path/to/sa-key.json'"
        echo "  Or run 'gcloud auth application-default login' first."
        exit 1
    fi
fi

header_text "Building kaudy container image"
podman build -t "${IMAGE}" .

header_text "Loading image into kind"
kind load docker-image "${IMAGE}"

if [ -n "${SKILL_IMAGE}" ]; then
  header_text "Loading skill image into kind"
  kind load docker-image "${SKILL_IMAGE}"
fi

header_text "Creating vertex-credentials secret"
kubectl delete secret vertex-credentials --ignore-not-found
kubectl create secret generic vertex-credentials \
  --from-file=service-account.json="${CREDENTIALS_FILE}"

header_text "Deploying kaudy pod"
kubectl delete pod kaudy --ignore-not-found --wait=true 2>/dev/null || true
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
    command: ["bash", "-c"]
    args:
    - |
      mkdir -p \$HOME/.claude/skills
      for d in /opt/skills-*/skills/*/; do
        ln -sfn "\$d" "\$HOME/.claude/skills/\$(basename "\$d")"
      done
      exec claude --dangerously-skip-permissions
    env:
    - name: CLAUDE_CODE_USE_VERTEX
      value: "1"
    - name: ANTHROPIC_VERTEX_PROJECT_ID
      value: "${ANTHROPIC_VERTEX_PROJECT_ID}"
    - name: CLOUD_ML_REGION
      value: "${CLOUD_ML_REGION}"
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: "/var/run/secrets/gcloud/service-account.json"
    workingDir: /workspace
    volumeMounts:
    - name: workspace
      mountPath: /workspace
    - name: vertex-credentials
      mountPath: /var/run/secrets/gcloud
      readOnly: true
    - name: skills
      mountPath: /opt/skills-0
    stdin: true
    tty: true
  volumes:
  - name: workspace
    emptyDir: {}
  - name: vertex-credentials
    secret:
      secretName: vertex-credentials
  - name: skills
    image:
      reference: ${SKILL_IMAGE}
      pullPolicy: IfNotPresent
EOF

header_text "Waiting for kaudy pod to be ready"
kubectl wait --for=condition=Ready pod/kaudy --timeout=180s

header_text "kaudy pod is running!"
kubectl get pods -l app=kaudy

echo ""
echo "To exec into the kaudy pod:"
echo "  kubectl exec -it kaudy -- claude --dangerously-skip-permissions"
echo ""
echo "Claude Code connects to Vertex AI (project: ${ANTHROPIC_VERTEX_PROJECT_ID}, region: ${CLOUD_ML_REGION})"
