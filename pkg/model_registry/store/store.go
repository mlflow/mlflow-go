package store

import (
	"context"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/entities"
	"github.com/mlflow/mlflow-go/pkg/protos"
)

type ModelRegistryStore interface {
	contract.Destroyer
	GetLatestVersions(ctx context.Context, name string, stages []string) ([]*protos.ModelVersion, *contract.Error)
	UpdateRegisteredModel(ctx context.Context, name, description string) (*entities.RegisteredModel, *contract.Error)
	RenameRegisteredModel(ctx context.Context, name, newName string) (*entities.RegisteredModel, *contract.Error)
	DeleteRegisteredModel(ctx context.Context, name string) *contract.Error
}
