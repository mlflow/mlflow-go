package service

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/mlflow/mlflow-go/pkg/config"
	"github.com/mlflow/mlflow-go/pkg/model_registry/store"
	"github.com/mlflow/mlflow-go/pkg/model_registry/store/sql"
)

type ModelRegistryService struct {
	store  store.ModelRegistryStore
	config *config.Config
}

func NewModelRegistryService(logger *logrus.Logger, config *config.Config) (*ModelRegistryService, error) {
	store, err := sql.NewModelRegistrySQLStore(logger, config)
	if err != nil {
		return nil, fmt.Errorf("could not create new sql store: %w", err)
	}

	return &ModelRegistryService{
		store:  store,
		config: config,
	}, nil
}
