# Architecture

## Overview

KaudY is a Go CLI that launches Claude Code inside a container with optional OCI skill images mounted as libraries.

```
kaudy run [flags] [-- claude-args...]
         |
         v
   +-----------+
   |  RunOptions|
   +-----------+
         |
    NewRunner(mode)
         |
    +----+----+
    |         |
 Podman   Kubernetes
 Runner     Runner
    |         |
 syscall   Pod YAML
  .Exec    (dry-run)
```

## Skill Mounting

### Podman

Each skill image is mounted read-only via `--mount type=image,src=<img>,dst=/opt/skills-<N>`. Before launching claude, a bash script symlinks `/opt/skills-*/skills/*/` into `$HOME/.claude/skills/`.

### Kubernetes (future)

Skill images become init containers that copy `/skills/` into a shared `emptyDir` volume. The main container then symlinks from `/opt/skills/` into `$HOME/.claude/skills/`.

## Environment Variables

Forwarded into the container: `ANTHROPIC_API_KEY`, `CLAUDE_CODE_USE_VERTEX`, `CLOUD_ML_REGION`, `ANTHROPIC_VERTEX_PROJECT_ID`, and IDE-related vars.

## Agentic Workflow

This project is designed for agentic development:

1. **AGENTS.md** provides context for AI agents to understand the codebase
2. **CI pipeline** validates every change automatically
3. **`--dry-run`** enables agents to verify container commands without execution
4. Agents can iterate: modify code, run `make test`, check CI, fix issues
