package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestEvent(t *testing.T) {
	// single event
	wf := &Workflow{}
	err := yaml.Unmarshal([]byte(`on: push`), wf)
	assert.Nil(t, err)
	assert.EqualValues(t, []string{"push"}, wf.GetEvent())

	// multiple events
	wf = &Workflow{}
	err = yaml.Unmarshal([]byte(`on: [push, fork]`), wf)
	assert.Nil(t, err)
	assert.EqualValues(t, []string{"push", "fork"}, wf.GetEvent())

	// push event
	wf = &Workflow{}
	err = yaml.Unmarshal([]byte(`
on:
  push:
    branches:
      - main
    tags:
      - 1.1
    paths:
      - /work
    paths-ignore:
      - /bin
    branches-ignore:
      - bugfix`), wf)
	assert.Nil(t, err)
	assert.EqualValues(t, []string{"push"}, wf.GetEvent())
	var es *EventSource
	es, err = wf.GetEventDetail("push")
	assert.Nil(t, err)
	assert.Equal(t, EventSource{
		Branches:       []string{"main"},
		Tags:           []string{"1.1"},
		Paths:          []string{"/work"},
		PathsIgnore:    []string{"/bin"},
		BranchesIgnore: []string{"bugfix"},
	}, *es)

	// schedule
	wf = &Workflow{}
	err = yaml.Unmarshal([]byte(`on:
  schedule:
    - cron: '30 5 * * 1,3'
    - cron: '30 5 * * 2,4'`), wf)
	assert.Nil(t, err)
	var schedules []Schedule
	schedules, err = wf.GetSchedules()
	assert.Nil(t, err)
	assert.EqualValues(t, []Schedule{{
		Cron: "30 5 * * 1,3",
	}, {
		Cron: "30 5 * * 2,4",
	}}, schedules)
}

func TestMergeYAML(t *testing.T) {
	tests := []struct {
		name     string
		origin   string
		sub      string
		expected string
		hasErr   bool
	}{{
		name:   "origin YAML is invalid",
		origin: "name=rick",
		hasErr: true,
	}, {
		name:   "sub YAML is invalid",
		sub:    "name=rick",
		hasErr: true,
	}, {
		name:   "simple without conflicts",
		origin: `name: rick`,
		sub:    `age: 12`,
		expected: `age: 12
name: rick
`,
	}, {
		name:   "have the same key",
		origin: `name: rick`,
		sub:    `name: linuxsuren`,
		expected: `name: linuxsuren
`,
	}, {
		name: "multiple levels",
		origin: `name: rick
works:
  age: 12`,
		sub: `works:
  term: 10
  subject: math`,
		expected: `name: rick
works:
  age: 12
  subject: math
  term: 10
`,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mergeYAML(tt.origin, tt.sub)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.hasErr, err != nil, err)
		})
	}
}
