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
LITELLM_MODEL_NAME=${LITELLM_MODEL_NAME:-"mistral-small-24b-w8a8"}

header_text "Building kaudy container image"
podman build -t "${IMAGE}" .

header_text "Loading image into kind"
kind load docker-image "${IMAGE}"

if [ -n "${SKILL_IMAGE}" ]; then
  header_text "Loading skill image into kind"
  kind load docker-image "${SKILL_IMAGE}"
fi

if [ -z "${LITELLM_BASE_URL}" ]; then
    echo "Error: LITELLM_BASE_URL environment variable is not set"
    echo "  export LITELLM_BASE_URL='https://your-model-endpoint/v1'"
    exit 1
fi

if [ -z "${LITELLM_API_KEY}" ]; then
    echo "Error: LITELLM_API_KEY environment variable is not set"
    echo "  export LITELLM_API_KEY='your-api-key'"
    exit 1
fi

header_text "Creating litellm-env secret"
kubectl delete secret litellm-env --ignore-not-found
kubectl create secret generic litellm-env \
  --from-literal=LITELLM_BASE_URL="${LITELLM_BASE_URL}" \
  --from-literal=LITELLM_API_KEY="${LITELLM_API_KEY}"

header_text "Creating litellm config"
kubectl delete configmap litellm-config --ignore-not-found
kubectl create configmap litellm-config --from-literal=config.yaml="$(cat <<YAML
model_list:
  - model_name: ${LITELLM_MODEL_NAME}
    litellm_params:
      model: openai/${LITELLM_MODEL_NAME}
      api_base: os.environ/LITELLM_BASE_URL
      api_key: os.environ/LITELLM_API_KEY

litellm_settings:
  master_key: sk-litellm
  use_chat_completions_url_for_anthropic_messages: true
  drop_params: true

router_settings:
  disable_cooldowns: true
YAML
)"

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
    - name: ANTHROPIC_BASE_URL
      value: "http://localhost:4000"
    - name: ANTHROPIC_AUTH_TOKEN
      value: "sk-litellm"
    - name: ANTHROPIC_API_KEY
      value: ""
    - name: ANTHROPIC_CUSTOM_MODEL_OPTION
      value: "${LITELLM_MODEL_NAME}"
    - name: CLAUDE_CODE_DISABLE_ADAPTIVE_THINKING
      value: "1"
    - name: CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS
      value: "1"
    - name: MAX_THINKING_TOKENS
      value: "0"
    workingDir: /workspace
    volumeMounts:
    - name: workspace
      mountPath: /workspace
    - name: skills
      mountPath: /opt/skills-0
    stdin: true
    tty: true
  - name: litellm
    image: ghcr.io/berriai/litellm:main-latest
    args: ["--config", "/etc/litellm/config.yaml", "--port", "4000"]
    ports:
    - containerPort: 4000
    envFrom:
    - secretRef:
        name: litellm-env
    volumeMounts:
    - name: litellm-config
      mountPath: /etc/litellm
      readOnly: true
  volumes:
  - name: workspace
    emptyDir: {}
  - name: skills
    image:
      reference: ${SKILL_IMAGE}
      pullPolicy: IfNotPresent
  - name: litellm-config
    configMap:
      name: litellm-config
EOF

header_text "Waiting for kaudy pod to be ready"
kubectl wait --for=condition=Ready pod/kaudy --timeout=180s

header_text "kaudy pod is running!"
kubectl get pods -l app=kaudy

echo ""
echo "To exec into the kaudy pod:"
echo "  kubectl exec -it kaudy -- claude --dangerously-skip-permissions --model ${LITELLM_MODEL_NAME}"
echo ""
echo "The LiteLLM sidecar proxies requests to your model endpoint."
echo "Claude Code talks to LiteLLM at http://localhost:4000 (Anthropic Messages API)."
echo "LiteLLM forwards to your model endpoint (OpenAI Chat Completions API)."
