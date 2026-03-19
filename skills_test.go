package kaudy

import (
	"strings"
	"testing"
)

func TestSkillSymlinkScript(t *testing.T) {
	t.Run("no args", func(t *testing.T) {
		script := SkillSymlinkScript(nil)
		if !strings.Contains(script, "mkdir -p $HOME/.claude/skills") {
			t.Error("script should create skills directory")
		}
		if !strings.Contains(script, "exec claude --dangerously-skip-permissions") {
			t.Error("script should exec claude")
		}
	})

	t.Run("with args", func(t *testing.T) {
		script := SkillSymlinkScript([]string{"-p", "fix tests"})
		if !strings.Contains(script, `"-p"`) {
			t.Error("script should contain quoted -p arg")
		}
		if !strings.Contains(script, `"fix tests"`) {
			t.Error("script should contain quoted 'fix tests' arg")
		}
	})
}
