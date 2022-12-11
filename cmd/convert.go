package cmd

import (
	"github.com/linuxsuren/github-action-workflow/pkg"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v2"
	"os"
	"strings"
)

func newConvertCmd() (c *cobra.Command) {
	opt := &convertOption{}
	c = &cobra.Command{
		Use:     "convert",
		Example: "gaw convert .github/workflows/build.yaml",
		Short:   "Convert GitHub Actions workflow file to Argo Workflows",
		Args:    cobra.MinimumNArgs(1),
		RunE:    opt.runE,
	}

	flags := c.Flags()
	flags.StringToStringVarP(&opt.env, "env", "e", nil,
		"Environment variables for all steps")
	return
}

func (o *convertOption) runE(cmd *cobra.Command, args []string) (err error) {
	gh := &pkg.Workflow{}
	var data []byte
	if data, err = os.ReadFile(args[0]); err == nil {
		if err = yaml.Unmarshal(data, gh); err != nil {
			return
		}
	}

	for i, job := range gh.Jobs {
		for j, step := range job.Steps {
			if step.Env == nil {
				gh.Jobs[i].Steps[j].Env = o.env
			} else {
				maps.Copy(gh.Jobs[i].Steps[j].Env, o.env)
			}
		}
	}

	var result string
	if result, err = gh.ConvertToArgoWorkflow(); err == nil {
		cmd.Println(strings.TrimSpace(result))
	}
	return
}

type convertOption struct {
	env map[string]string
}
