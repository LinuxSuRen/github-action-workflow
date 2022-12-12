package pkg

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
)

func k8sStyleName(name string) (result string) {
	result = strings.ToLower(name)
	result = strings.ReplaceAll(result, " ", "-")
	return
}

func (w *Workflow) ConvertToArgoWorkflow() (output string, err error) {
	// pre-handle
	defaultImage := "alpine"
	w.Name = k8sStyleName(w.Name)
	for i := range w.Jobs {
		job := w.Jobs[i]
		job.Name = k8sStyleName(job.Name)
		for j := range w.Jobs[i].Steps {
			w.Jobs[i].Steps[j].Name = k8sStyleName(w.Jobs[i].Steps[j].Name)

			if strings.HasPrefix(w.Jobs[i].Steps[j].Uses, "actions/checkout") {
				w.Jobs[i].Steps[j].Image = "alpine/git:v2.26.2"
				w.Jobs[i].Steps[j].Run = "git clone https://gitee.com/LinuxSuRen/yaml-readme ."
			} else if strings.HasPrefix(w.Jobs[i].Steps[j].Uses, "actions/setup-go") {
				defaultImage = "golang:1.19"

				if ver, ok := w.Jobs[i].Steps[j].With["go-version"]; ok {
					defaultImage = fmt.Sprintf("golang:%s", ver)
				}
			} else if strings.HasPrefix(w.Jobs[i].Steps[j].Uses, "goreleaser/goreleaser-action") {
				w.Jobs[i].Steps[j].Image = "goreleaser/goreleaser:v1.13.1"
				w.Jobs[i].Steps[j].Run = "goreleaser " + w.Jobs[i].Steps[j].With["args"]
			} else if w.Jobs[i].Steps[j].Uses != "" {
				// TODO not support yet, do nothing
			} else {
				w.Jobs[i].Steps[j].Image = defaultImage
			}
			w.Jobs[i].Steps[j].Run = strings.TrimSpace(w.Jobs[i].Steps[j].Run)
		}

		// TODO currently we can only handle one job
		break
	}

	var t *template.Template
	t, err = template.New("argo").Parse(argoworkflowTemplate)

	data := bytes.NewBuffer([]byte{})
	if err = t.Execute(data, w); err == nil {
		output = strings.TrimSpace(data.String())
	}
	return
}

var argoworkflowTemplate = `
apiVersion: argoproj.io/v1alpha1
kind: WorkflowTemplate
metadata:
  name: {{.Name}}
spec:
  entrypoint: main
  volumeClaimTemplates:
    - metadata:
        name: work
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 64Mi

  templates:
    - name: main
      dag:
        tasks:
      {{- range $key, $job := .Jobs}}
      {{- range $i, $step := $job.Steps}}
      {{- if $step.Image}}
          - name: {{$step.Name}}
            template: {{$step.Name}}
      {{- end}}
      {{- end}}
      {{- end}}

      {{- range $key, $job := .Jobs}}
      {{- range $i, $step := $job.Steps}}
      {{- if $step.Image}}
    - name: {{$step.Name}}
      script:
        image: {{$step.Image}}
        command: [sh]
        {{- if $step.Env}}
        env:
        {{- range $k, $v := $step.Env}}
          - name: {{$k}}
            value: {{$v}}
        {{- end}}
        {{- end}}
        source: |
          {{$step.Run}}
        volumeMounts:
          - mountPath: /work
            name: work
        workingDir: /work
      {{- end}}
      {{- end}}
      {{- end}}
`
