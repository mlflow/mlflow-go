package entities

import "github.com/mlflow/mlflow-go/pkg/protos"

type RunTag struct {
	Key   string
	Value string
}

func (t RunTag) ToProto() *protos.RunTag {
	return &protos.RunTag{
		Key:   &t.Key,
		Value: &t.Value,
	}
}

func NewTagFromProto(proto *protos.RunTag) *RunTag {
	return &RunTag{
		Key:   proto.GetKey(),
		Value: proto.GetValue(),
	}
}
