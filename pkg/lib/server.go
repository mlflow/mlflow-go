package main

import "C"

import (
	"unsafe"

	"github.com/sirupsen/logrus"

	"github.com/mlflow/mlflow-go/pkg/config"
	"github.com/mlflow/mlflow-go/pkg/server"
)

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
