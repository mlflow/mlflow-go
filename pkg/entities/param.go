package entities

import "github.com/mlflow/mlflow-go/pkg/protos"

type Param struct {
	Key   string
	Value string
}

func (p Param) ToProto() *protos.Param {
	return &protos.Param{
		Key:   &p.Key,
		Value: &p.Value,
	}
}

func ParamFromProto(proto *protos.Param) *Param {
	return &Param{
		Key:   *proto.Key,
		Value: *proto.Value,
	}
}
