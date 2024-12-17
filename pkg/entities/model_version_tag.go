package entities

import (
	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/utils"
)

type ModelVersionTag struct {
	Key   string
	Value string
}

func (mvt ModelVersionTag) ToProto() *protos.ModelVersionTag {
	return &protos.ModelVersionTag{
		Key:   utils.PtrTo(mvt.Key),
		Value: utils.PtrTo(mvt.Value),
	}
}
