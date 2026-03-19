package kaudy

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// PodmanRunner implements Runner using podman.
type PodmanRunner struct{}

func (r *PodmanRunner) Run(opts *RunOptions) error {
	if opts.Rebuild {
		if err := r.rebuildImage(opts.Image); err != nil {
			return err
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	u, err := user.Current()
	if err != nil {
		return fmt.Errorf("getting current user: %w", err)
	}

	// Ensure host paths exist
	os.MkdirAll(filepath.Join(home, ".claude"), 0o755)
	touchFile(filepath.Join(home, ".claude.json"))

	tmpDir := fmt.Sprintf("/tmp/claude-%s", u.Uid)
	os.MkdirAll(tmpDir, 0o755)

	containerName := fmt.Sprintf("claude-%s-%s",
		sanitizeName(filepath.Base(opts.Workdir)),
		time.Now().Format("20060102-150405"))

	args := []string{
		"run", "--rm", "-it",
		"--tz=local",
		"--name", containerName,
		"--network=host",
		"--userns=keep-id",
		"--pid=host",
		"-v", fmt.Sprintf("%s:%s:z", opts.Workdir, opts.Workdir),
		"-v", fmt.Sprintf("%s/.claude.json:%s/.claude.json:rw,z", home, home),
		"-v", fmt.Sprintf("%s/.claude:%s/.claude:rw,z", home, home),
		"-v", fmt.Sprintf("%s:%s:rw,z", tmpDir, tmpDir),
	}

	// Optional volume mounts
	for _, m := range OptionalMounts(home) {
		if _, err := os.Stat(m.HostPath); err == nil {
			mode := "rw,z"
			if m.ReadOnly {
				mode = "ro,z"
			}
			// linuxbrew doesn't need SELinux relabel
			if m.HostPath == "/home/linuxbrew/.linuxbrew" {
				mode = strings.TrimSuffix(mode, ",z")
			}
			args = append(args, "-v", fmt.Sprintf("%s:%s:%s", m.HostPath, m.HostPath, mode))
		}
	}

	// Skill image mounts
	for i, img := range opts.SkillImages {
		args = append(args, "--mount",
			fmt.Sprintf("type=image,src=%s,dst=/opt/skills-%d,readwrite=false", img, i))
	}

	args = append(args, "-w", opts.Workdir)

	// Environment variables
	for _, env := range PassthroughEnvVars {
		args = append(args, "-e", env)
	}

	args = append(args, opts.Image)

	// Entrypoint: if skills present, use bash -c to symlink then exec claude
	if len(opts.SkillImages) > 0 {
		args = append(args, "bash", "-c", SkillSymlinkScript(opts.ClaudeArgs))
	} else {
		args = append(args, "claude", "--dangerously-skip-permissions")
		args = append(args, opts.ClaudeArgs...)
	}

	if opts.DryRun {
		fmt.Println("podman " + shelljoin(args))
		return nil
	}

	// Replace this process with podman
	podmanPath, err := exec.LookPath("podman")
	if err != nil {
		return fmt.Errorf("podman not found in PATH: %w", err)
	}

	return syscall.Exec(podmanPath, append([]string{"podman"}, args...), os.Environ())
}

func (r *PodmanRunner) rebuildImage(image string) error {
	fmt.Fprintf(os.Stderr, "Building %s...\n", image)
	cmd := exec.Command("podman", "build", "-t", image, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func touchFile(path string) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o644)
	if err == nil {
		f.Close()
	}
}

func sanitizeName(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}
	return b.String()
}

func shelljoin(args []string) string {
	parts := make([]string, len(args))
	for i, a := range args {
		if strings.ContainsAny(a, " \t\n\"'\\$`!#&|;(){}") {
			parts[i] = fmt.Sprintf("%q", a)
		} else {
			parts[i] = a
		}
	}
	return strings.Join(parts, " ")
}
