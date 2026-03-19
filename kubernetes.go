package kaudy

import (
	"fmt"
	"os"
	"strings"
	"text/template"
)

// KubernetesRunner implements Runner by rendering a Pod YAML.
// Currently only supports --dry-run mode.
type KubernetesRunner struct{}

const podTemplate = `apiVersion: v1
kind: Pod
metadata:
  name: kaudy
  labels:
    app: kaudy
spec:{{if .SkillImages}}
  initContainers:{{range $i, $img := .SkillImages}}
  - name: skill-{{$i}}
    image: {{$img}}
    command: ["cp", "-r", "/skills/.", "/opt/skills/"]
    volumeMounts:
    - name: skills
      mountPath: /opt/skills{{end}}{{end}}
  containers:
  - name: claude
    image: {{.Image}}
    workingDir: /workspace
    command: [{{.Command}}]
    env:{{range .EnvVars}}
    - name: {{.}}
      value: ""{{end}}
    volumeMounts:
    - name: workspace
      mountPath: /workspace{{if .SkillImages}}
    - name: skills
      mountPath: /opt/skills{{end}}
  volumes:
  - name: workspace
    emptyDir: {}{{if .SkillImages}}
  - name: skills
    emptyDir: {}{{end}}
`

type podTemplateData struct {
	Image       string
	SkillImages []string
	EnvVars     []string
	Command     string
}

func (r *KubernetesRunner) Run(opts *RunOptions) error {
	if !opts.DryRun {
		return fmt.Errorf("kubernetes mode only supports --dry-run currently")
	}

	command := `"claude", "--dangerously-skip-permissions"`
	if len(opts.SkillImages) > 0 {
		script := SkillSymlinkScript(opts.ClaudeArgs)
		command = fmt.Sprintf(`"bash", "-c", %q`, script)
	} else {
		for _, a := range opts.ClaudeArgs {
			command += fmt.Sprintf(`, %q`, a)
		}
	}

	data := podTemplateData{
		Image:       opts.Image,
		SkillImages: opts.SkillImages,
		EnvVars:     PassthroughEnvVars,
		Command:     command,
	}

	tmpl, err := template.New("pod").Parse(podTemplate)
	if err != nil {
		return fmt.Errorf("parsing pod template: %w", err)
	}

	var sb strings.Builder
	if err := tmpl.Execute(&sb, data); err != nil {
		return fmt.Errorf("rendering pod template: %w", err)
	}

	fmt.Fprint(os.Stdout, sb.String())
	return nil
}
