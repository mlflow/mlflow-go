package entities

import "github.com/mlflow/mlflow-go/pkg/protos"

type TraceTag struct {
	Key       string
	Value     string
	RequestID string
}

func NewTraceTagFromProto(proto *protos.SetTraceTag) *TraceTag {
	return &TraceTag{
		Key:       proto.GetKey(),
		Value:     proto.GetValue(),
		RequestID: proto.GetRequestId(),
	}
}
