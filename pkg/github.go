package pkg

import (
	"gopkg.in/yaml.v2"
)

type Workflow struct {
	Name        string
	On          interface{} // could be: string, []string, Event
	Jobs        map[string]Job
	Concurrency string
	Raw         string

	// extra fields
	GitRepository string
	Cron          string
}

func (w *Workflow) GetEventDetail(name string) (es *EventSource, err error) {
	if name != "schedule" {
		switch w.On.(type) {
		case map[interface{}]interface{}:
			raw := w.On.(map[interface{}]interface{})

			if val, ok := raw[name]; ok {
				var data []byte
				if data, err = yaml.Marshal(val); err == nil {
					es = &EventSource{}
					err = yaml.Unmarshal(data, es)
				}
			}
		}
	}
	return
}

func (w *Workflow) GetSchedules() (schedules []Schedule, err error) {
	switch w.On.(type) {
	case map[interface{}]interface{}:
		raw := w.On.(map[interface{}]interface{})

		if val, ok := raw["schedule"]; ok {
			var data []byte
			if data, err = yaml.Marshal(val); err == nil {
				schedules = []Schedule{}
				err = yaml.Unmarshal(data, &schedules)
			}
		}
	}
	return
}

func (w *Workflow) GetEvent() (result []string) {
	switch w.On.(type) {
	case string:
		result = []string{w.On.(string)}
	case []interface{}:
		for _, item := range w.On.([]interface{}) {
			result = append(result, item.(string))
		}
	case map[interface{}]interface{}:
		for key := range w.On.(map[interface{}]interface{}) {
			result = append(result, key.(string))
		}
	}
	return
}

type Event struct {
	Push        EventSource
	PullRequest EventSource `yaml:"pull_request"`
	Schedule    []string
}

type EventSource struct {
	Branches       []string
	Tags           []string
	Paths          []string
	PathsIgnore    []string `yaml:"paths-ignore"`
	BranchesIgnore []string `yaml:"branches-ignore"`
}

type Schedule struct {
	Cron string
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
	Image   string
	Depends string
	Secret  string
}

func mergeYAML(origin, sub string) (result string, err error) {
	originObj := map[interface{}]interface{}{}
	subObj := map[interface{}]interface{}{}

	if err = yaml.Unmarshal([]byte(origin), originObj); err != nil {
		return
	}
	if err = yaml.Unmarshal([]byte(sub), subObj); err != nil {
		return
	}

	var data []byte
	resultObj := mergeMaps(originObj, subObj)
	if data, err = yaml.Marshal(resultObj); err == nil {
		result = string(data)
	}
	return
}

func mergeMaps(a, b map[interface{}]interface{}) map[interface{}]interface{} {
	out := make(map[interface{}]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}

	for k, v := range b {
		switch tv := v.(type) {
		case map[interface{}]interface{}:
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[interface{}]interface{}); ok {
					out[k] = mergeMaps(bv, tv)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}
