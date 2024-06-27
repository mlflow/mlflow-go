package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/mlflow/mlflow-go/mlflow_go/go/config"
	"github.com/mlflow/mlflow-go/mlflow_go/go/server"
)

func main() {
	var config config.Config

	loggerInstance := logrus.StandardLogger()

	if err := json.Unmarshal([]byte(os.Getenv("MLFLOW_GO_CONFIG")), &config); err != nil {
		loggerInstance.Fatal(err)
	}

	logLevel, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		loggerInstance.Fatal(err)
	}

	loggerInstance.SetLevel(logLevel)
	loggerInstance.Debugf("Loaded config: %#v", config)

	ctx, cancel := context.WithCancel(context.Background())

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigint)

	go func() {
		<-sigint
		loggerInstance.Info("Shutting down MLflow Go server")
		cancel()
	}()

	loggerInstance.Infof("Starting MLflow Go server on http://%s", config.Address)

	if err := server.Launch(ctx, loggerInstance, &config); err != nil {
		loggerInstance.Fatal(err)
	}
}
