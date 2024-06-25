package service

import (
	"github.com/sirupsen/logrus"

	"github.com/mlflow/mlflow-go/mlflow_go/go/config"
)

type ModelRegistryService struct {
	config *config.Config
}

func NewModelRegistryService(_ *logrus.Logger, config *config.Config) (*ModelRegistryService, error) {
	return &ModelRegistryService{
		config: config,
	}, nil
}
