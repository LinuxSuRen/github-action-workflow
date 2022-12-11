package main

import (
	"github.com/linuxsuren/github-action-workflow/cmd"
	"os"
)

func main() {
	if err := cmd.NewRoot().Execute(); err != nil {
		os.Exit(1)
	}
}
