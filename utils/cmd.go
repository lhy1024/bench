package utils

import (
	"bytes"
	"os/exec"

	"github.com/pingcap/log"
	"go.uber.org/zap"
)

// Command directly run command
type Command struct {
	path string
	args []string
}

// NewCommand returns Command
func NewCommand(path string, args ...string) *Command {
	return &Command{path: path, args: args}
}

// Run run command and return result
func (command *Command) Run() (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Cmd{Path: command.path, Args: command.args}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	log.Info(cmd.Path, zap.Strings("cmd", cmd.Args),
		zap.String("stdout", stdout.String()), zap.String("stderr", stderr.String()))
	return stdout.String(), err
}
