package kaudy

// Runner builds and executes a container command for running Claude Code.
type Runner interface {
	// Run executes the container with the given options.
	// For podman mode this replaces the process via syscall.Exec.
	// For kubernetes mode with dry-run it prints Pod YAML.
	Run(opts *RunOptions) error
}

// NewRunner returns a Runner for the given mode.
func NewRunner(mode string) Runner {
	switch mode {
	case "kubernetes":
		return &KubernetesRunner{}
	default:
		return &PodmanRunner{}
	}
}

// PassthroughEnvVars are environment variables forwarded into the container.
var PassthroughEnvVars = []string{
	"ANTHROPIC_API_KEY",
	"CLAUDE_CODE_USE_VERTEX",
	"CLOUD_ML_REGION",
	"ANTHROPIC_VERTEX_PROJECT_ID",
	"CLAUDE_CODE_SSE_PORT",
	"CLAUDE_CODE_ENTRYPOINT",
	"ENABLE_IDE_INTEGRATION",
	"TERMINAL_EMULATOR",
	"CLAUDECODE",
}

// OptionalVolumeMounts maps host paths to container paths; only mounted when
// the host path exists.
type OptionalMount struct {
	HostPath string
	ReadOnly bool
}

// OptionalMounts returns volume mounts that are only added when the host path exists.
func OptionalMounts(home string) []OptionalMount {
	return []OptionalMount{
		{HostPath: home + "/bin", ReadOnly: false},
		{HostPath: "/home/linuxbrew/.linuxbrew", ReadOnly: false},
		{HostPath: home + "/.config/gcloud", ReadOnly: true},
		{HostPath: home + "/.config/gh", ReadOnly: true},
		{HostPath: home + "/.config/JetBrains", ReadOnly: true},
	}
}
