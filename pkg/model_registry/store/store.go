package store

import (
	"context"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/protos"
)

type ModelRegistryStore interface {
	contract.Destroyer
	GetLatestVersions(ctx context.Context, name string, stages []string) ([]*protos.ModelVersion, *contract.Error)
}
