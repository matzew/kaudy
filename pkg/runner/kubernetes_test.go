package runner

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestKubernetesRunnerDryRun(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	kr := &KubernetesRunner{}
	err := kr.Run(&RunOptions{
		Image:       "quay.io/matzew/kaudy:latest",
		SkillImages: []string{"quay.io/matzew/agent-skills"},
		DryRun:      true,
	})

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	yaml := buf.String()

	if !strings.Contains(yaml, "apiVersion: v1") {
		t.Error("output should contain apiVersion")
	}
	if !strings.Contains(yaml, "quay.io/matzew/agent-skills") {
		t.Error("output should contain skill image")
	}
	if !strings.Contains(yaml, "initContainers") {
		t.Error("output should contain initContainers when skills are present")
	}
}

func TestKubernetesRunnerRequiresDryRun(t *testing.T) {
	kr := &KubernetesRunner{}
	err := kr.Run(&RunOptions{
		Image:  "quay.io/matzew/kaudy:latest",
		DryRun: false,
	})
	if err == nil {
		t.Error("expected error when not using --dry-run")
	}
}
