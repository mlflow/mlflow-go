package entities

import (
	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/utils"
)

type RegisteredModelTag struct {
	Key   string
	Value string
}

func (t RegisteredModelTag) ToProto() *protos.RegisteredModelTag {
	return &protos.RegisteredModelTag{
		Key:   utils.PtrTo(t.Key),
		Value: utils.PtrTo(t.Value),
	}
}

func NewRegisteredModelTagFromProto(proto *protos.RegisteredModelTag) *RegisteredModelTag {
	return &RegisteredModelTag{
		Key:   proto.GetKey(),
		Value: proto.GetValue(),
	}
}
