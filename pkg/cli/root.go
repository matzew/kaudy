package cli

import (
	"fmt"
	"os"

	"github.com/matzew/kaudy/pkg/runner"
	"github.com/spf13/cobra"
)

const defaultImage = "quay.io/matzew/kaudy:latest"

// NewRootCommand creates the top-level cobra command with the "run" subcommand.
func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "kaudy",
		Short: "Run Claude Code in a container with optional OCI skill images",
	}

	root.AddCommand(newRunCommand())
	return root
}

func newRunCommand() *cobra.Command {
	opts := &runner.RunOptions{}

	cmd := &cobra.Command{
		Use:   "run [flags] [-- claude-args...]",
		Short: "Launch Claude Code in a container",
		Long:  "Start a containerised Claude Code session with optional skill image mounts.",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.ClaudeArgs = args

			if opts.Workdir == "" {
				wd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("getting working directory: %w", err)
				}
				opts.Workdir = wd
			}

			r := runner.NewRunner(opts.Mode)
			return r.Run(opts)
		},
	}

	f := cmd.Flags()
	f.StringArrayVarP(&opts.SkillImages, "skill-image", "s", nil, "OCI skill image to mount (repeatable)")
	f.StringVar(&opts.Mode, "mode", "podman", `runner mode: "podman" or "kubernetes"`)
	f.StringVar(&opts.Image, "image", defaultImage, "base container image")
	f.StringVar(&opts.Workdir, "workdir", "", "project directory to mount (default: $PWD)")
	f.BoolVar(&opts.DryRun, "dry-run", false, "print command or YAML without executing")
	f.BoolVar(&opts.Rebuild, "rebuild", false, "force rebuild base image before running")

	return cmd
}
