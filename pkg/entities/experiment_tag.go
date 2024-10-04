package entities

import "github.com/mlflow/mlflow-go/pkg/protos"

type ExperimentTag struct {
	Key   string
	Value string
}

func (et *ExperimentTag) ToProto() *protos.ExperimentTag {
	return &protos.ExperimentTag{
		Key:   &et.Key,
		Value: &et.Value,
	}
}

func NewExperimentTagFromProto(proto *protos.ExperimentTag) *ExperimentTag {
	return &ExperimentTag{
		Key:   proto.GetKey(),
		Value: proto.GetValue(),
	}
}
