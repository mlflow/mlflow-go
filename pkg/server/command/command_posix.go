//go:build !windows

package command

import (
	"os/exec"
	"syscall"

	"github.com/sirupsen/logrus"
)

func newProcessGroupCommand(cmd *exec.Cmd) (*exec.Cmd, error) {
	// Create the process in a new process group
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}

	// Terminate the process group
	cmd.Cancel = func() error {
		logrus.Debug("Sending interrupt signal to command process group")

		return syscall.Kill(-cmd.Process.Pid, syscall.SIGINT)
	}

	return cmd, nil
}
