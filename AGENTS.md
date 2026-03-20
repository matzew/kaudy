# AGENTS.md

Instructions for AI agents working on this codebase.

## Project

KaudY runs Claude Code in a container with optional OCI skill image mounts. Podman locally, Kubernetes Pods later.

## Build & Test

```bash
make build     # go build ./cmd/kaudy/
make test      # go test ./...
make container # podman build -t quay.io/matzew/kaudy:latest .
```

## Architecture

- `pkg/runner/runner.go`: `Runner` interface, `RunOptions`, `NewRunner` factory, shared config
- `pkg/runner/podman.go`: `PodmanRunner` implementation
- `pkg/runner/kubernetes.go`: `KubernetesRunner` implementation
- `pkg/skills/skills.go`: `SkillSymlinkScript` helper
- `pkg/cli/root.go`: Cobra command wiring
- `cmd/kaudy/main.go`: entrypoint
- Dependency DAG (no cycles): `main → cli → runner → skills`
- Skill images are OCI images mounted via `--mount type=image` (podman) or init containers (k8s)
- `syscall.Exec` replaces the Go process with podman for clean TTY/signal handling

## Code Conventions

- Go module: `github.com/matzew/kaudy`
- CLI framework: cobra
- Container engine: podman (not docker)
- All commits must be signed off: `git commit -s`

## Testing

- Unit tests live next to the code they test (`*_test.go`)
- `--dry-run` is the primary way to verify command construction without running containers
- Tests should validate the generated podman args and Kubernetes YAML

## CI

GitHub Actions runs on every push and PR:
1. `go build ./cmd/kaudy/`
2. `go vet ./...`
3. `go test ./...`
