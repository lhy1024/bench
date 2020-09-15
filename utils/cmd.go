package utils

import (
	"bytes"
	"os/exec"

	"github.com/pingcap/log"
	"go.uber.org/zap"
)

type Command struct {
	path string
	args []string
}

func NewCommand(path string, args ...string) *Command {
	return &Command{path: path, args: args}
}

func (command *Command) Run() error {
	var stdout, stderr bytes.Buffer
	cmd := exec.Cmd{Path: command.path, Args: command.args}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	log.Debug(cmd.Path, zap.Strings("cmd", cmd.Args),
		zap.String("stdout", stdout.String()), zap.String("stderr", stderr.String()))
	return err
}
