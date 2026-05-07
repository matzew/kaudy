package runner

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/matzew/kaudy/pkg/skills"
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
spec:
  containers:
  - name: claude
    image: {{.Image}}
    workingDir: /workspace
    command: [{{.Command}}]
    env:{{range .EnvVars}}
    - name: {{.Name}}
      value: "{{.Value}}"{{end}}
    volumeMounts:
    - name: workspace
      mountPath: /workspace{{if .VertexAuth}}
    - name: vertex-credentials
      mountPath: /var/run/secrets/gcloud
      readOnly: true{{end}}{{range $i, $img := .SkillImages}}
    - name: skill-{{$i}}
      mountPath: /opt/skills-{{$i}}{{end}}
    stdin: true
    tty: true
  volumes:
  - name: workspace
    emptyDir: {}{{if .VertexAuth}}
  - name: vertex-credentials
    secret:
      secretName: vertex-credentials{{end}}{{range $i, $img := .SkillImages}}
  - name: skill-{{$i}}
    image:
      reference: {{$img}}
      pullPolicy: IfNotPresent{{end}}
`

type envVar struct {
	Name  string
	Value string
}

type podTemplateData struct {
	Image       string
	SkillImages []string
	EnvVars     []envVar
	Command     string
	VertexAuth  bool
}

func (r *KubernetesRunner) Run(opts *RunOptions) error {
	if !opts.DryRun {
		return fmt.Errorf("kubernetes mode only supports --dry-run currently")
	}

	command := `"claude", "--dangerously-skip-permissions"`
	if len(opts.SkillImages) > 0 {
		script := skills.SkillSymlinkScript(opts.ClaudeArgs)
		command = fmt.Sprintf(`"bash", "-c", %q`, script)
	} else {
		for _, a := range opts.ClaudeArgs {
			command += fmt.Sprintf(`, %q`, a)
		}
	}

	vertexAuth := os.Getenv("CLAUDE_CODE_USE_VERTEX") == "1"

	var envVars []envVar
	for _, name := range PassthroughEnvVars {
		ev := envVar{Name: name}
		if vertexAuth && name == "GOOGLE_APPLICATION_CREDENTIALS" {
			ev.Value = "/var/run/secrets/gcloud/service-account.json"
		}
		envVars = append(envVars, ev)
	}

	data := podTemplateData{
		Image:       opts.Image,
		SkillImages: opts.SkillImages,
		EnvVars:     envVars,
		Command:     command,
		VertexAuth:  vertexAuth,
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
