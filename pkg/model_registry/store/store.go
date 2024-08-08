package store

import (
	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/protos"
)

type ModelRegistryStore interface {
	GetLatestVersions(name string, stages []string) ([]*protos.ModelVersion, *contract.Error)
}
