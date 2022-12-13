package cmd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCmdMeta(t *testing.T) {
	command := NewRoot()
	assert.Equal(t, "gaw", command.Use)
}
