package store

import (
	"github.com/mlflow/mlflow-go/mlflow_go/go/contract"
	"github.com/mlflow/mlflow-go/mlflow_go/go/protos"
)

type ModelRegistryStore interface {
	GetLatestVersions(name string, stages []string) ([]*protos.ModelVersion, *contract.Error)
}
