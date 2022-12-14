package pkg

type Workflow struct {
	Name string
	Jobs map[string]Job

	// extra fields
	GitRepository string
}

type Job struct {
	Name   string
	RunsOn string `yaml:"runs-on"`
	Steps  []Step
}

type Step struct {
	Name string
	Uses string
	Env  map[string]string
	With map[string]string
	Run  string
	ID   string

	// extra fields
	Image string
}
