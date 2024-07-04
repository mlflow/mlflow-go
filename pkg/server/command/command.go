package command

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/mlflow/mlflow-go/pkg/config"
)

func LaunchCommand(ctx context.Context, cfg *config.Config) error {
	//nolint:gosec
	cmd, err := newProcessGroupCommand(
		exec.CommandContext(ctx, cfg.PythonCommand[0], cfg.PythonCommand[1:]...),
	)
	if err != nil {
		return fmt.Errorf("failed to create process group command: %w", err)
	}

	cmd.Env = os.Environ()
	cmd.Stdout = logrus.StandardLogger().Writer()
	cmd.Stderr = logrus.StandardLogger().Writer()
	cmd.WaitDelay = 5 * time.Second //nolint:mnd

	logrus.Debugf("Launching command: %v", cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to launch command: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command exited with error: %w", err)
	}

	return nil
}
