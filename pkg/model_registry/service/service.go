package service

import (
	"context"
	"fmt"

	"github.com/mlflow/mlflow-go/pkg/config"
	"github.com/mlflow/mlflow-go/pkg/model_registry/store"
	"github.com/mlflow/mlflow-go/pkg/model_registry/store/sql"
)

type ModelRegistryService struct {
	store  store.ModelRegistryStore
	config *config.Config
}

func NewModelRegistryService(ctx context.Context, config *config.Config) (*ModelRegistryService, error) {
	store, err := sql.NewModelRegistrySQLStore(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create new sql store: %w", err)
	}

	return &ModelRegistryService{
		store:  store,
		config: config,
	}, nil
}

func (m *ModelRegistryService) Close() error {
	if err := m.store.Close(); err != nil {
		return fmt.Errorf("failed to close store: %w", err)
	}

	return nil
}
