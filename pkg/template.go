package pkg

import _ "embed"

//go:embed data/cronWorkflow.yaml
var cronWorkflowTemplate string

//go:embed data/workflow.yaml
var argoworkflowTemplate string

//go:embed data/workflowEventBinding.yaml
var argoworkflowEventBinding string

//go:embed data/role.yaml
var eventBindingRole string

//go:embed data/secret.yaml
var eventBindingSecret string

//go:embed data/gitlab/rolebinding.yaml
var eventBindingGitlabRoleBinding string

//go:embed data/gitlab/serviceaccount.yaml
var eventBindingGitlabServiceAccount string

//go:embed data/gitlab/secret_gitlab.com.yaml
var eventBindingGitlabServiceAccountSecret string

//go:embed data/github/rolebinding.yaml
var eventBindingGitHubRoleBinding string

//go:embed data/github/serviceaccount.yaml
var eventBindingGitHubServiceAccount string

//go:embed data/github/secret_github.com.yaml
var eventBindingGitHubServiceAccountSecret string
