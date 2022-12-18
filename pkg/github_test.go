package pkg

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
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
}
