package cmd

import (
	"fmt"
	"github.com/linuxsuren/github-action-workflow/pkg"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_convertOption_convert(t *testing.T) {
	type fields struct {
		env map[string]string
	}
	type args struct {
		gh *pkg.Workflow
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantOutput string
		wantErr    bool
	}{{
		name: "simple",
		args: args{
			gh: &pkg.Workflow{
				Name: "simple",
				Jobs: map[string]pkg.Job{
					"simple": {
						Name: "test",
						Steps: []pkg.Step{{
							Name: "test",
							Run:  "echo 1",
						}},
					},
				}},
		},
		wantOutput: "data/one-workflow.yaml",
	}, {
		name: "workflow with env",
		args: args{
			gh: &pkg.Workflow{
				Name: "simple",
				Jobs: map[string]pkg.Job{
					"simple": {
						Name: "test",
						Steps: []pkg.Step{{
							Name: "test",
							Run:  "echo 1",
							Env:  map[string]string{},
						}},
					},
				}},
		},
		fields:     fields{env: map[string]string{"key": "value"}},
		wantOutput: "data/one-workflow-env.yaml",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &convertOption{
				env: tt.fields.env,
			}
			gotOutput, err := o.convert(tt.args.gh)
			if (err != nil) != tt.wantErr {
				t.Errorf("convert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if data, err := os.ReadFile(tt.wantOutput); err == nil {
				tt.wantOutput = string(data)
			}
			assert.Equal(t, tt.wantOutput, gotOutput)
		})
	}
}

func Test_convertOption_convertWorkflows(t *testing.T) {
	type fields struct {
		env map[string]string
	}
	type args struct {
		ghs []*pkg.Workflow
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantOutput string
		wantErr    assert.ErrorAssertionFunc
	}{{
		name: "two workflows",
		args: args{
			ghs: []*pkg.Workflow{{
				Name: "simple",
				Jobs: map[string]pkg.Job{
					"simple": {
						Name: "test",
						Steps: []pkg.Step{{
							Name: "test",
							Run:  "echo 1",
							Env:  map[string]string{},
						}},
					},
				}}, {
				Name: "simple",
				Jobs: map[string]pkg.Job{
					"simple": {
						Name: "test",
						Steps: []pkg.Step{{
							Name: "test",
							Run:  "echo 1",
							Env:  map[string]string{},
						}},
					},
				}}},
		},
		wantOutput: "data/combine-workflow.yaml",
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			return true
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &convertOption{
				env: tt.fields.env,
			}
			gotOutput, err := o.convertWorkflows(tt.args.ghs)
			if !tt.wantErr(t, err, fmt.Sprintf("convertWorkflows(%v)", tt.args.ghs)) {
				return
			}
			if data, err := os.ReadFile(tt.wantOutput); err == nil {
				tt.wantOutput = string(data)
			}
			assert.Equal(t, tt.wantOutput, gotOutput)
		})
	}
}
