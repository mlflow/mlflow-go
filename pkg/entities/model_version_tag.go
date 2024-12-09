package entities

import "github.com/mlflow/mlflow-go/pkg/protos"

type ModelVersionTag struct {
	Key     string
	Value   string
	Name    string
	Version int32
}

func (mvt ModelVersionTag) ToProto() *protos.ModelVersionTag {
	return &protos.ModelVersionTag{
		Key:   &mvt.Key,
		Value: &mvt.Value,
	}
}
