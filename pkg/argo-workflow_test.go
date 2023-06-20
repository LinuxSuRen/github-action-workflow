package pkg

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestWorkflow_ConvertToArgoWorkflow(t *testing.T) {
	tests := []struct {
		name          string
		githubActions string
		argoWorkflows string
		wantErr       bool
	}{{
		name:          "simple",
		githubActions: "testdata/github-actions.yaml",
		argoWorkflows: "testdata/argo-workflows.yaml",
	}, {
		name:          "with image",
		githubActions: "testdata/github-action-image.yaml",
		argoWorkflows: "testdata/argo-workflows-image.yaml",
	}, {
		name:          "complex event",
		githubActions: "testdata/github-action-complex-event.yaml",
		argoWorkflows: "testdata/argo-workflows-complex-event.yaml",
	}, {
		name:          "with concurrency",
		githubActions: "testdata/github-actions-concurrency.yaml",
		argoWorkflows: "testdata/argo-workflows-concurrency.yaml",
	}, {
		name:          "with schedule",
		githubActions: "testdata/github-actions-schedule.yaml",
		argoWorkflows: "testdata/argo-workflows-schedule.yaml",
	}, {
		name:          "with retry event",
		githubActions: "testdata/github-actions-retry-event.yaml",
		argoWorkflows: "testdata/argo-workflows-retry-event.yaml",
	}, {
		name:          "with raw",
		githubActions: "testdata/github-actions-raw.yaml",
		argoWorkflows: "testdata/argo-workflows-raw.yaml",
	}, {
		name:          "with template referemce",
		githubActions: "testdata/github-actions-template-ref.yaml",
		argoWorkflows: "testdata/argo-workflows-template-ref.yaml",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Workflow{GitRepository: "https://gitee.com/LinuxSuRen/yaml-readme"}
			data, err := os.ReadFile(tt.githubActions)
			assert.Nil(t, err)
			err = yaml.Unmarshal(data, w)
			assert.Nil(t, err)

			gotOutput, err := w.ConvertToArgoWorkflow(false)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToArgoWorkflow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			wantData, err := os.ReadFile(tt.argoWorkflows)
			assert.Nil(t, err)
			assert.Equalf(t, string(wantData), gotOutput, gotOutput)
		})
	}

	// workflow name is empty
	wf := &Workflow{}
	result, err := wf.ConvertToArgoWorkflow(false)
	assert.Equal(t, "", result)
	assert.Nil(t, err)
}

func Test_getProjectName(t *testing.T) {
	type args struct {
		projectName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "gaw1",
			args: args{
				projectName: "https://github.com/LinuxSuRen/github-action-workflow.git",
			},
			want: "LinuxSuRen/github-action-workflow",
		},
		{
			name: "gaw2",
			args: args{
				projectName: "git@github.com:LinuxSuRen/github-action-workflow.git",
			},
			want: "LinuxSuRen/github-action-workflow",
		},
		{
			name: "gaw3",
			args: args{
				projectName: "git@github.com:group0/group1/LinuxSuRen/github-action-workflow.git",
			},
			want: "group0/group1/LinuxSuRen/github-action-workflow",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getProjectName(tt.args.projectName); got != tt.want {
				t.Errorf("getProjectName() = %v, want %v", got, tt.want)
			}
		})
	}
}
