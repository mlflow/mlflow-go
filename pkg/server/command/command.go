package command

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/mlflow/mlflow-go/pkg/config"
	"github.com/mlflow/mlflow-go/pkg/utils"
)

func LaunchCommand(ctx context.Context, cfg *config.Config) error {
	logger := utils.GetLoggerFromContext(ctx)

	//nolint:gosec
	cmd, err := newProcessGroupCommand(
		ctx,
		exec.CommandContext(ctx, cfg.PythonCommand[0], cfg.PythonCommand[1:]...),
	)
	if err != nil {
		return fmt.Errorf("failed to create process group command: %w", err)
	}

	cmd.Env = append(os.Environ(), cfg.PythonEnv...)
	cmd.Stdout = logger.Writer()
	cmd.Stderr = logger.Writer()
	cmd.WaitDelay = 5 * time.Second //nolint:mnd

	logger.Debugf("Launching command: %v", cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to launch command: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command exited with error: %w", err)
	}

	return nil
}
