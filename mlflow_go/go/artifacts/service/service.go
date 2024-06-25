package service

import (
	"github.com/sirupsen/logrus"

	"github.com/mlflow/mlflow-go/mlflow_go/go/config"
)

type ArtifactsService struct {
	config *config.Config
}

func NewArtifactsService(_ *logrus.Logger, config *config.Config) (*ArtifactsService, error) {
	return &ArtifactsService{
		config: config,
	}, nil
}
