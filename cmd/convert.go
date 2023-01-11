package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/linuxsuren/github-action-workflow/pkg"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v2"
)

func newConvertCmd() (c *cobra.Command) {
	opt := &convertOption{}
	c = &cobra.Command{
		Use:     "convert",
		Example: "gaw convert .github/workflows/build.yaml",
		Short:   "Convert GitHub Actions workflow file to Argo Workflows",
		Args:    cobra.MinimumNArgs(1),
		PreRunE: opt.preRunE,
		RunE:    opt.runE,
	}

	flags := c.Flags()
	flags.StringToStringVarP(&opt.env, "env", "e", nil,
		"Environment variables for all steps")
	return
}

func (o *convertOption) preRunE(cmd *cobra.Command, args []string) (err error) {
	var repo *git.Repository
	if repo, err = git.PlainOpenWithOptions(".", &git.PlainOpenOptions{
		DetectDotGit: true,
	}); err == nil {
		var remote *git.Remote
		if remote, err = repo.Remote(git.DefaultRemoteName); err == nil {
			o.gitRepository = remote.Config().URLs[0]
		}
	}
	return
}

func (o *convertOption) runE(cmd *cobra.Command, args []string) (err error) {
	var result string
	if result, err = o.convertWorkflowsFromFilePath(args[0]); err == nil {
		cmd.Println(result)
	}
	return
}

func (o *convertOption) convertWorkflowsFromFilePath(targetPath string) (result string, err error) {
	var ghs []*pkg.Workflow
	var files []string
	if files, err = filepath.Glob(targetPath); err == nil {
		for _, file := range files {
			var data []byte
			if data, err = os.ReadFile(file); err == nil {
				gh := &pkg.Workflow{GitRepository: o.gitRepository}
				if err = yaml.Unmarshal(data, gh); err == nil {
					ghs = append(ghs, gh)
					continue
				}
			}
			return
		}
		result, err = o.convertWorkflows(ghs)
	}
	return
}

func (o *convertOption) convertWorkflows(ghs []*pkg.Workflow) (output string, err error) {
	buf := bytes.Buffer{}

	var result string
	for i := range ghs {
		if result, err = o.convert(ghs[i]); err != nil || result == "" {
			return
		}

		buf.WriteString("\n---\n")
		buf.WriteString(strings.TrimSpace(result))
	}
	output = buf.String()
	return
}

func (o *convertOption) convert(gh *pkg.Workflow) (output string, err error) {
	for i, job := range gh.Jobs {
		for j, step := range job.Steps {
			if step.Env == nil {
				gh.Jobs[i].Steps[j].Env = o.env
			} else {
				maps.Copy(gh.Jobs[i].Steps[j].Env, o.env)
			}
		}
	}

	o.preHandle(gh)

	output, err = gh.ConvertToArgoWorkflow()
	return
}

func (o *convertOption) preHandle(gh *pkg.Workflow) {
	if gh.Name == "" {
		return
	}

	if prefix, ok := o.env["prefix"]; ok && prefix != "" {
		gh.Name = prefix + gh.Name
	}
}

type convertOption struct {
	env           map[string]string
	gitRepository string
}
