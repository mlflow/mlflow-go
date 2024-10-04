//go:build !windows

package command

import (
	"context"
	"os/exec"
	"syscall"

	"github.com/mlflow/mlflow-go/pkg/utils"
)

func newProcessGroupCommand(ctx context.Context, cmd *exec.Cmd) (*exec.Cmd, error) {
	logger := utils.GetLoggerFromContext(ctx)

	// Create the process in a new process group
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}

	// Terminate the process group
	cmd.Cancel = func() error {
		logger.Debug("Sending interrupt signal to command process group")

		return syscall.Kill(-cmd.Process.Pid, syscall.SIGINT)
	}

	return cmd, nil
}
