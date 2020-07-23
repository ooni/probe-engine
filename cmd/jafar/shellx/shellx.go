// Package shellx allows to run commands
package shellx

import (
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/apex/log"
	"github.com/google/shlex"
)

// Run executes the specified command with the specified args
func Run(name string, arg ...string) error {
	log.Infof("exec: %s %s", name, strings.Join(arg, " "))
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	log.Infof("exec result: %+v", err)
	return err
}

// RunCommandline is like Run but its only argument is a command
// line that will be splitted using the google/shlex package
func RunCommandline(cmdline string) error {
	args, err := shlex.Split(cmdline)
	if err != nil {
		return err
	}
	if len(args) < 1 {
		return errors.New("shellx: no command to execute")
	}
	return Run(args[0], args[1:]...)
}
