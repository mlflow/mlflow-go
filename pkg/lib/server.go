package main

import "C"

import (
	"context"
	"unsafe"

	"github.com/sirupsen/logrus"

	"github.com/mlflow/mlflow-go/pkg/config"
	"github.com/mlflow/mlflow-go/pkg/server"
)

type serverInstance struct {
	cancel  context.CancelFunc
	errChan <-chan error
}

func (si serverInstance) Close() error {
	return nil
}

var serverInstances = newInstanceMap[serverInstance]()

//export LaunchServer
func LaunchServer(configData unsafe.Pointer, configSize C.int) int64 {
	cfg, err := config.NewConfigFromBytes(C.GoBytes(configData, configSize)) //nolint:nlreturn
	if err != nil {
		logrus.Error("Failed to read config: ", err)

		return -1
	}

	if err := server.LaunchWithSignalHandler(cfg); err != nil {
		logrus.Error("Failed to launch server: ", err)

		return -1
	}

	return 0
}

//export LaunchServerAsync
func LaunchServerAsync(configData unsafe.Pointer, configSize C.int) int64 {
	serverID := serverInstances.Create(
		func(ctx context.Context, cfg *config.Config) (serverInstance, error) {
			errChan := make(chan error, 1)

			ctx, cancel := context.WithCancel(ctx)

			go func() {
				errChan <- server.Launch(ctx, cfg)
			}()

			return serverInstance{
				cancel:  cancel,
				errChan: errChan,
			}, nil
		},
		C.GoBytes(configData, configSize), //nolint:nlreturn
	)

	return serverID
}

//export StopServer
func StopServer(serverID int64) int64 {
	instance, cErr := serverInstances.Get(serverID)
	if cErr != nil {
		logrus.Error("Failed to get instance: ", cErr)

		return -1
	}
	defer serverInstances.Destroy(serverID)

	instance.cancel()

	err := <-instance.errChan
	if err != nil {
		logrus.Error("Server has exited with error: ", err)

		return -1
	}

	return 0
}
