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

func NewInputTagFromProto(proto *protos.InputTag) *InputTag {
	return &InputTag{
		Key:   proto.GetKey(),
		Value: proto.GetValue(),
	}
}
