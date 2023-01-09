package pkg

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
)

func k8sStyleName(name string) (result string) {
	result = strings.ToLower(name)
	result = strings.ReplaceAll(result, " ", "-")
	return
}

func (w *Workflow) GetWorkflowBindings() (wfbs []WorkflowEventBinding) {
	projectName := w.GitRepository
	if strings.Contains(projectName, "/") {
		projectName = strings.TrimSuffix(projectName, ".git")
		projectName = strings.Split(projectName, "/")[1]
	}

	for _, e := range w.GetEvent() {
		binding := WorkflowEventBinding{
			Selector:   fmt.Sprintf(`payload.object_kind == "%s" && payload.project.path_with_namespace endsWith "%s"`, e, projectName),
			Ref:        w.Name,
			Name:       fmt.Sprintf("%s-%s", w.Name, e),
			Parameters: map[string]string{},
		}
		// read more about the expression from https://github.com/antonmedv/expr
		if es, err := w.GetEventDetail(e); err == nil && es != nil {
			if len(es.Branches) > 0 {
				branchSelector := "("
				for _, b := range es.Branches {
					branchSelector = branchSelector + getBranchSelector(e, b)
				}
				branchSelector = strings.TrimSuffix(branchSelector, " || ") + ")"
				binding.Selector = binding.Selector + " && " + branchSelector
			}
		}

		switch e {
		case "push":
			binding.Parameters["branch"] = "payload.ref"
			binding.Parameters["pr"] = "-1"
		case "merge_request":
			binding.Parameters["branch"] = "payload.object_attributes.source_branch"
			binding.Parameters["pr"] = "payload.object_attributes.iid"
			binding.Selector = binding.Selector + ` && payload.object_attributes.state == "opened"`
		default:
			continue
		}
		wfbs = append(wfbs, binding)
	}
	return
}

func getBranchSelector(eventName, branch string) string {
	switch eventName {
	case "push":
		return fmt.Sprintf(`payload.ref == "refs/heads/%s" || `, branch)
	default: // it should be merge_request
		return fmt.Sprintf("payload.object_attributes.target_branch == %s || ", branch)
	}
}

func (w *Workflow) ConvertToArgoWorkflow() (output string, err error) {
	if w.Name == "" {
		// name is required
		return
	}

	// pre-handle
	defaultImage := "alpine"
	w.Name = k8sStyleName(w.Name)
	for i := range w.Jobs {
		job := w.Jobs[i]
		job.Name = k8sStyleName(job.Name)
		var newSteps []Step
		for j := range w.Jobs[i].Steps {
			w.Jobs[i].Steps[j].Name = k8sStyleName(w.Jobs[i].Steps[j].Name)

			if strings.HasPrefix(w.Jobs[i].Steps[j].Uses, "actions/checkout") {
				w.Jobs[i].Steps[j].Image = "alpine/git:v2.26.2"
				w.Jobs[i].Steps[j].Run = fmt.Sprintf(`branch=$(echo {{workflow.parameters.branch}} | sed -e 's/refs\/heads\///g')
git clone --branch $branch %s .
if [ {{workflow.parameters.pr}} != -1 ]; then
  git fetch origin merge-requests/{{workflow.parameters.pr}}/head:mr-{{workflow.parameters.pr}}
  git checkout  mr-{{workflow.parameters.pr}}
fi`, w.GitRepository)
			} else if strings.HasPrefix(w.Jobs[i].Steps[j].Uses, "actions/setup-go") {
				defaultImage = "golang:1.19"

				if ver, ok := w.Jobs[i].Steps[j].With["go-version"]; ok {
					defaultImage = fmt.Sprintf("golang:%s", ver)
				}
				continue
			} else if strings.HasPrefix(w.Jobs[i].Steps[j].Uses, "goreleaser/goreleaser-action") {
				w.Jobs[i].Steps[j].Image = "goreleaser/goreleaser:v1.13.1"
				w.Jobs[i].Steps[j].Run = "goreleaser " + w.Jobs[i].Steps[j].With["args"]
			} else if strings.HasPrefix(w.Jobs[i].Steps[j].Uses, "docker://") {
				w.Jobs[i].Steps[j].Image = strings.TrimPrefix(w.Jobs[i].Steps[j].Uses, "docker://")
				w.Jobs[i].Steps[j].Run = w.Jobs[i].Steps[j].With["args"]
			} else if w.Jobs[i].Steps[j].Uses != "" {
				// TODO not support yet, do nothing
				continue
			} else {
				w.Jobs[i].Steps[j].Image = defaultImage
			}
			w.Jobs[i].Steps[j].Run = strings.TrimSpace(w.Jobs[i].Steps[j].Run)
			newSteps = append(newSteps, w.Jobs[i].Steps[j])
		}

		// make sure a correct depends order
		for j := 1; j < len(newSteps); j++ {
			newSteps[j].Depends = newSteps[j-1].Name
		}
		(&job).Steps = newSteps
		w.Jobs[i] = job

		// TODO currently we can only handle one job
		break
	}

	// generate workflowTemplate
	var t *template.Template
	if t, err = template.New("argo").Funcs(sprig.GenericFuncMap()).Parse(argoworkflowTemplate); err != nil {
		return
	}
	data := bytes.NewBuffer([]byte{})
	if err = t.Execute(data, w); err == nil {
		output = strings.TrimSpace(data.String())
	}

	// generate workflowEventBinding
	for _, binding := range w.GetWorkflowBindings() {
		if t, err = template.New("argo").Funcs(sprig.GenericFuncMap()).Parse(argoworkflowEventBinding); err != nil {
			return
		}
		data := bytes.NewBuffer([]byte{})
		if err = t.Execute(data, binding); err == nil {
			output = output + "\n---\n" + strings.TrimSpace(data.String())
		}
	}

	// generate cronWorkflow
	var schedules []Schedule
	schedules, err = w.GetSchedules()
	for _, schedule := range schedules {
		if t, err = template.New("argo").Funcs(sprig.GenericFuncMap()).Parse(cronWorkflowTemplate); err != nil {
			return
		}
		data := bytes.NewBuffer([]byte{})
		w.Cron = schedule.Cron
		if err = t.Execute(data, w); err == nil {
			output = output + "\n---\n" + strings.TrimSpace(data.String())
		}
	}
	return
}

type WorkflowEventBinding struct {
	Name       string
	Ref        string
	Selector   string
	Parameters map[string]string
}

var argoworkflowEventBinding = `
apiVersion: argoproj.io/v1alpha1
kind: WorkflowEventBinding
metadata:
  name: {{.Name}}
spec:
  event:
    selector: {{.Selector}}
  submit:
    workflowTemplateRef:
      name: {{.Ref}}
   {{- if .Parameters}}
    arguments:
      parameters:
       {{- range $key, $val := .Parameters}}
        - name: {{$key}}
          valueFrom:
            event: "{{$val}}"
        {{- end}}
	{{- end}}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: submit-workflow-template
rules:
  - apiGroups:
      - argoproj.io
    resources:
      - workfloweventbindings
    verbs:
      - list
  - apiGroups:
      - argoproj.io
    resources:
      - workflowtemplates
    verbs:
      - get
  - apiGroups:
      - argoproj.io
    resources:
      - workflows
    verbs:
      - create
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: github.com
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: github.com
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: submit-workflow-template
subjects:
  - kind: ServiceAccount
    name: github.com
    namespace: default
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: gitlab.com
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: gitlab.com
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: submit-workflow-template
subjects:
  - kind: ServiceAccount
    name: gitlab.com
    namespace: default
---
kind: Secret
apiVersion: v1
metadata:
  name: argo-workflows-webhook-clients
stringData:
  bitbucket.org: |
    type: bitbucket
  bitbucketserver: |
    type: bitbucketserver
  github.com: |
    type: github
  gitlab.com: |
    type: gitlab`

var cronWorkflowTemplate = `
apiVersion: argoproj.io/v1alpha1
kind: CronWorkflow
metadata:
  name: {{.Name}}
spec:
  schedule: "{{.Cron}}"
  concurrencyPolicy: "Replace"
  startingDeadlineSeconds: 0
  workflowSpec:
    entrypoint: main
    templates:
      - name: main
        dag:
          tasks:
        {{- range $key, $job := .Jobs}}
        {{- range $i, $step := $job.Steps}}
        {{- if $step.Image}}
            - name: {{$step.Name}}
              template: {{$step.Name}}
              {{- if $step.Depends}}
              depends: {{$step.Depends}}
              {{- end}}
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
{{indent 12 $step.Run}}
        {{- end}}
        {{- end}}
        {{- end}}`

var argoworkflowTemplate = `
apiVersion: argoproj.io/v1alpha1
kind: WorkflowTemplate
metadata:
  name: {{.Name}}
spec:
  entrypoint: main
  arguments:
    parameters:
      - name: branch
        default: master
      - name: pr
        default: -1
  volumeClaimTemplates:
    - metadata:
        name: work
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 64Mi
  {{- if .Concurrency}}
  synchronization:
    mutex:
      name: {{.Concurrency}}
  {{- end}}
  templates:
    - name: main
      dag:
        tasks:
      {{- range $key, $job := .Jobs}}
      {{- range $i, $step := $job.Steps}}
      {{- if $step.Image}}
          - name: {{$step.Name}}
            template: {{$step.Name}}
            {{- if $step.Depends}}
            depends: {{$step.Depends}}
            {{- end}}
      {{- end}}
      {{- end}}
      {{- end}}

      {{- range $key, $job := .Jobs}}
      {{- range $i, $step := $job.Steps}}
      {{- if $step.Image}}
    - name: {{$step.Name}}
      {{- if $step.Secret}}
      volumes:
        - name: {{$step.Secret}}
          secret:
            defaultMode: 0400
            secretName: {{$step.Secret}}
      {{- end}}
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
{{indent 10 $step.Run}}
        volumeMounts:
          - mountPath: /work
            name: work
        {{- if $step.Secret}}
          - mountPath: /root/.ssh/
            name: {{$step.Secret}}
        {{- end}}
        workingDir: /work
      {{- end}}
      {{- end}}
      {{- end}}
`
