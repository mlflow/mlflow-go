package entities

import "github.com/mlflow/mlflow-go/pkg/protos"

type InputTag struct {
	Key   string
	Value string
}

func (i InputTag) ToProto() *protos.InputTag {
	return &protos.InputTag{
		Key:   &i.Key,
		Value: &i.Value,
	}
}
