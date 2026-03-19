package kaudy

import "fmt"

// SkillSymlinkScript returns a bash snippet that symlinks skill directories
// from /opt/skills-* into $HOME/.claude/skills/ and then exec's claude.
func SkillSymlinkScript(claudeArgs []string) string {
	// Build quoted arg list for the exec line
	quoted := ""
	for _, a := range claudeArgs {
		quoted += fmt.Sprintf(" %q", a)
	}

	return fmt.Sprintf(`mkdir -p $HOME/.claude/skills
for d in /opt/skills-*/skills/*/; do
  ln -sfn "$d" "$HOME/.claude/skills/$(basename "$d")"
done
exec claude --dangerously-skip-permissions%s`, quoted)
}
