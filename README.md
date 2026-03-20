# KaudY

Run [Claude Code](https://docs.anthropic.com/en/docs/claude-code) in a container with optional OCI skill images mounted as libraries.

Starts with podman locally, designed to support Kubernetes Pods later.

## Quick Start

```bash
# Build
make build

# Run Claude Code in a container (uses $PWD as the project directory)
kaudy run

# Mount a skill image
kaudy run -s quay.io/matzew/agent-skills

# Multiple skills + claude args
kaudy run -s quay.io/matzew/agent-skills -s quay.io/other/skills -- -p "fix tests"

# Preview the podman command without executing
kaudy run --dry-run

# Render a Kubernetes Pod YAML
kaudy run --mode kubernetes --dry-run -s quay.io/matzew/agent-skills
```

## Install

```bash
go install github.com/matzew/kaudy/cmd/kaudy@latest
```

Or build from source:

```bash
git clone git@github.com:matzew/kaudy.git
cd kaudy
make build    # binary at ./kaudy
make install  # go install
```

## Base Container Image

Build the base image that includes Fedora, Node.js, and Claude Code:

```bash
make container
```

This builds `quay.io/matzew/kaudy:latest` using the included `Containerfile`.

## CLI Reference

```
kaudy run [flags] [-- claude-args...]

Flags:
  -s, --skill-image string   OCI skill image to mount (repeatable)
      --mode string           Runner mode: "podman" or "kubernetes" (default "podman")
      --image string          Base container image (default "quay.io/matzew/kaudy:latest")
      --workdir string        Project directory to mount (default $PWD)
      --dry-run               Print command or YAML without executing
      --rebuild               Force rebuild base image before running
```

## Skill Images

Skill images are OCI images containing Claude Code skills. They follow a simple layout:

```
/skills/<skill-name>/
```

Build a skill image with a `Containerfile` like:

```dockerfile
FROM scratch
COPY my-skill/ /skills/my-skill/
```

When mounted, skills are symlinked into `$HOME/.claude/skills/` before Claude starts.

## How It Works

1. `kaudy run` constructs a `podman run` command with volume mounts for the project directory, Claude config, and optional credential directories
2. Skill images are mounted read-only via `--mount type=image`
3. The Go process replaces itself with podman via `syscall.Exec` for clean TTY and signal handling
4. Environment variables (`ANTHROPIC_API_KEY`, Vertex AI config, IDE vars) are forwarded into the container

## Kubernetes Mode

`--mode kubernetes` renders a Pod YAML with skill images as init containers. Currently `--dry-run` only.

```bash
kaudy run --mode kubernetes --dry-run -s quay.io/matzew/agent-skills | kubectl apply -f -
```

## Testing with kind

You can run kaudy in a local [kind](https://kind.sigs.k8s.io/) cluster for testing.

### Prerequisites

- [kind](https://kind.sigs.k8s.io/) installed
- `podman` available (used as the container engine and kind provider on Linux)
- `LITELLM_API_BASE` — base URL of your OpenAI-compatible model endpoint
- `LITELLM_API_KEY` — API key for the model endpoint

### Create the cluster

```bash
./scripts/00-kind-setup.sh
```

This creates a kind cluster using podman, waits for kube-system pods, and patches CoreDNS to use `8.8.8.8`.

### Build and deploy

```bash
export LITELLM_API_BASE='https://your-model-endpoint/v1'
export LITELLM_API_KEY='your-model-api-key'
./scripts/01-deploy-kaudy.sh
```

This deploys a pod with two containers:

- **kaudy** — Claude Code, configured to talk to LiteLLM at `http://localhost:4000`
- **litellm** — sidecar proxy that translates Anthropic Messages API to OpenAI Chat Completions API and forwards requests to your model endpoint

The model endpoint URL and API key are stored in a Kubernetes secret (`litellm-env`), not in the ConfigMap.

You can override the model name with `LITELLM_MODEL_NAME` (default: `mistral-small-24b-w8a8`).

### Interact with the pod

```bash
kubectl get pods -l app=kaudy
kubectl exec -it kaudy -- claude --dangerously-skip-permissions --model mistral-small-24b-w8a8
```

### Cleanup

```bash
kind delete cluster
```

## Pushing the Container Image

```bash
make push
```

This builds the image (via the `container` target) and pushes it to the registry configured in `IMAGE` (default `quay.io/matzew/kaudy:latest`).

## License

Apache-2.0
