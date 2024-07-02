package main

import "C"

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"unsafe"

	"github.com/sirupsen/logrus"

	"github.com/mlflow/mlflow-go/pkg/config"
	"github.com/mlflow/mlflow-go/pkg/server"
)

//export LaunchServer
func LaunchServer(configData unsafe.Pointer, configSize C.int) int64 {
	logger := logrus.StandardLogger()

	var config config.Config
	//nolint:nlreturn
	if err := json.Unmarshal(C.GoBytes(configData, configSize), &config); err != nil {
		logger.Error("Failed to parse JSON config: ", err)

		return -1
	}

	logLevel, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		logger.Error("Failed to parse log level: ", err)

		return -1
	}

	logger.SetLevel(logLevel)

	logger.Debugf("Loaded config: %#v", config)

	ctx, cancel := context.WithCancel(context.Background())

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigint)

	go func() {
		<-sigint
		logger.Info("Shutting down MLflow Go server")
		cancel()
	}()

	logger.Infof("Starting MLflow Go server on http://%s", config.Address)

	if err := server.Launch(ctx, logger, &config); err != nil {
		logger.Error("Failed to launch server: ", err)

		return -1
	}

	return 0
}
