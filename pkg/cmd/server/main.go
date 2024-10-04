package main

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/mlflow/mlflow-go/pkg/config"
	"github.com/mlflow/mlflow-go/pkg/server"
)

func main() {
	cfg, err := config.NewConfigFromString(os.Getenv("MLFLOW_GO_CONFIG"))
	if err != nil {
		logrus.Fatal("Failed to read config from MLFLOW_GO_CONFIG environment variable: ", err)
	}

	if err := server.LaunchWithSignalHandler(cfg); err != nil {
		logrus.Fatal("Failed to launch server: ", err)
	}
}
