package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

func NewRoot() (c *cobra.Command) {
	c = &cobra.Command{
		Use:   "gaw",
		Short: "GitHub Actions workflow compatible tool",
	}

	c.SetOut(os.Stdout)
	c.AddCommand(newConvertCmd())
	return
}
