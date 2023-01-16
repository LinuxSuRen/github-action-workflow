package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/linuxsuren/github-action-workflow/pkg"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestCmdConvert(t *testing.T) {
	command := newConvertCmd()
	assert.Equal(t, "convert", command.Use)
}

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

func Test_convertOption_convertWorkflowsFromFilePath(t *testing.T) {
	type fields struct {
		env map[string]string
	}
	type args struct {
		targetPath string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantResult string
		wantErr    assert.ErrorAssertionFunc
	}{{
		name:       "one github workflow",
		args:       args{targetPath: "data/github-workflow-*.yaml"},
		wantResult: simpleWorkflow,
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err)
			return true
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &convertOption{
				env: tt.fields.env,
			}
			gotResult, err := o.convertWorkflowsFromFilePath(tt.args.targetPath)
			if !tt.wantErr(t, err, fmt.Sprintf("convertWorkflowsFromFilePath(%v)", tt.args.targetPath)) {
				return
			}
			assert.Equalf(t, tt.wantResult, gotResult, "convertWorkflowsFromFilePath(%v)", tt.args.targetPath)
		})
	}
}

const simpleWorkflow = `
---
apiVersion: argoproj.io/v1alpha1
kind: WorkflowTemplate
metadata:
  name: build
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
  templates:
    - name: main
      dag:
        tasks:
          - name: test
            template: test
    - name: test
      script:
        image: alpine
        command: [sh]
        source: |
          go test ./... -coverprofile coverage.out
        volumeMounts:
          - mountPath: /work
            name: work
        workingDir: /work`

func Test_convertOption_runE(t *testing.T) {
	type fields struct {
		env map[string]string
	}
	type args struct {
		cmd  *cobra.Command
		args []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		prepare func(c *cobra.Command) *bytes.Buffer
		check   func(t assert.TestingT, buf *bytes.Buffer)
		wantErr assert.ErrorAssertionFunc
	}{{
		name: "simple",
		args: args{
			cmd:  &cobra.Command{},
			args: []string{"data/github-workflow-*.yaml"},
		},
		prepare: func(c *cobra.Command) *bytes.Buffer {
			buf := bytes.Buffer{}
			c.SetOut(&buf)
			return &buf
		},
		check: func(t assert.TestingT, buf *bytes.Buffer) {
			assert.Equal(t, simpleWorkflow+"\n", buf.String())
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err)
			return true
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &convertOption{
				env: tt.fields.env,
			}
			buf := tt.prepare(tt.args.cmd)
			err := o.runE(tt.args.cmd, tt.args.args)
			tt.check(t, buf)
			tt.wantErr(t, err, fmt.Sprintf("runE(%v, %v)", tt.args.cmd, tt.args.args))
		})
	}
}

func Test_convertOption_preRunE(t *testing.T) {
	type fields struct {
		env           map[string]string
		gitRepository string
	}
	type args struct {
		cmd  *cobra.Command
		args []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		check   func(t *testing.T, opt *convertOption)
		wantErr assert.ErrorAssertionFunc
	}{{
		name: "simple",
		check: func(t *testing.T, opt *convertOption) {
			assert.Contains(t, strings.ToLower(opt.gitRepository), "linuxsuren/github-action-workflow")
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err)
			return true
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &convertOption{
				env:           tt.fields.env,
				gitRepository: tt.fields.gitRepository,
			}
			err := o.preRunE(tt.args.cmd, tt.args.args)
			tt.check(t, o)
			tt.wantErr(t, err, fmt.Sprintf("preRunE(%v, %v)", tt.args.cmd, tt.args.args))
		})
	}
}

func TestPreHandle(t *testing.T) {
	opt := &convertOption{
		env: map[string]string{
			"prefix": "dev-",
		},
	}
	wfWithoutName := &pkg.Workflow{}
	opt.preHandle(wfWithoutName)
	assert.Empty(t, wfWithoutName.Name)

	wf := &pkg.Workflow{
		Name: "sample",
	}
	opt.preHandle(wf)
	assert.Equal(t, "dev-sample", wf.Name)

	// empty option
	emptyOpt := &convertOption{}
	sampleWf := &pkg.Workflow{
		Name: "sample",
	}
	emptyOpt.preHandle(sampleWf)
	assert.Equal(t, "sample", sampleWf.Name)
}

func TestParseEnv(t *testing.T) {
	opt := &convertOption{}
	os.Setenv("ARGOCD_ENV_prefix", "dev-")
	opt.parseEnv()
	assert.EqualValues(t, map[string]string{"prefix": "dev-"}, opt.env)
}
