package entities

import "github.com/mlflow/mlflow-go/pkg/protos"

type RunTag struct {
	Key   string
	Value string
}

func NewTagFromProto(proto *protos.RunTag) *RunTag {
	return &RunTag{
		Key:   proto.GetKey(),
		Value: proto.GetValue(),
	}
}
