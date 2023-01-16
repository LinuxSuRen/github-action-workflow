package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplateIsNotEmpty(t *testing.T) {
	assert.NotEmpty(t, cronWorkflowTemplate)
	assert.NotEmpty(t, argoworkflowTemplate)
	assert.NotEmpty(t, argoworkflowEventBinding)
	assert.NotEmpty(t, eventBindingRole)
	assert.NotEmpty(t, eventBindingSecret)
	assert.NotEmpty(t, eventBindingGitlabRoleBinding)
	assert.NotEmpty(t, eventBindingGitlabServiceAccount)
	assert.NotEmpty(t, eventBindingGitHubRoleBinding)
	assert.NotEmpty(t, eventBindingGitHubServiceAccount)
}
