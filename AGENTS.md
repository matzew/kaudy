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

- `Runner` interface in `container.go` with `PodmanRunner` and `KubernetesRunner` implementations
- `RunOptions` struct holds all CLI flags
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
