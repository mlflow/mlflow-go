package server

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/mlflow/mlflow-go/pkg/config"
	"github.com/mlflow/mlflow-go/pkg/server/command"
	"github.com/mlflow/mlflow-go/pkg/utils"
)

func Launch(ctx context.Context, cfg *config.Config) error {
	if len(cfg.PythonCommand) > 0 {
		return launchCommandAndServer(ctx, cfg)
	}

	return launchServer(ctx, cfg)
}

func launchCommandAndServer(ctx context.Context, cfg *config.Config) error {
	var errs []error

	logger := utils.GetLoggerFromContext(ctx)

	cmdCtx, cmdCancel := context.WithCancel(ctx)
	srvCtx, srvCancel := context.WithCancel(ctx)

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(1)

	go func() {
		defer waitGroup.Done()

		if err := command.LaunchCommand(cmdCtx, cfg); err != nil && cmdCtx.Err() == nil {
			errs = append(errs, err)
		}

		logger.Debug("Python server has exited")

		srvCancel()
	}()

	waitGroup.Add(1)

	go func() {
		defer waitGroup.Done()

		if err := launchServer(srvCtx, cfg); err != nil && srvCtx.Err() == nil {
			errs = append(errs, err)
		}

		logger.Debug("Go server has exited")

		cmdCancel()
	}()

	waitGroup.Wait()

	return errors.Join(errs...)
}

func LaunchWithSignalHandler(cfg *config.Config) error {
	logger := utils.NewLoggerFromConfig(cfg)

	logger.Debugf("Loaded config: %#v", cfg)

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigint)

	ctx, cancel := context.WithCancel(
		utils.NewContextWithLogger(context.Background(), logger))

	go func() {
		sig := <-sigint
		logger.Debugf("Received signal: %v", sig)

		cancel()
	}()

	return Launch(ctx, cfg)
}
