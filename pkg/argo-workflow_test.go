package pkg

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"os"
	"testing"
)

func TestWorkflow_ConvertToArgoWorkflow(t *testing.T) {
	tests := []struct {
		name          string
		githubActions string
		argoWorkflows string
		wantErr       bool
	}{{
		name:          "simple",
		githubActions: "data/github-actions.yaml",
		argoWorkflows: "data/argo-workflows.yaml",
	}, {
		name:          "with image",
		githubActions: "data/github-action-image.yaml",
		argoWorkflows: "data/argo-workflows-image.yaml",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Workflow{}
			data, err := os.ReadFile(tt.githubActions)
			assert.Nil(t, err)
			err = yaml.Unmarshal(data, w)
			assert.Nil(t, err)

			gotOutput, err := w.ConvertToArgoWorkflow()
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToArgoWorkflow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			wantData, err := os.ReadFile(tt.argoWorkflows)
			assert.Nil(t, err)
			assert.Equal(t, string(wantData), gotOutput)
		})
	}
}
