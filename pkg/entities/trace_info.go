package entities

import (
	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/utils"
)

type TraceInfo struct {
	RequestID            string
	Status               string
	ExperimentID         string
	TimestampMS          int64
	ExecutionTimeMS      *int64
	Tags                 []*TraceTag
	TraceRequestMetadata []*TraceRequestMetadata
}

func (ti TraceInfo) ToProto() *protos.TraceInfo {
	traceInfo := protos.TraceInfo{
		RequestId:       &ti.RequestID,
		ExperimentId:    &ti.ExperimentID,
		TimestampMs:     &ti.TimestampMS,
		ExecutionTimeMs: ti.ExecutionTimeMS,
		Status:          utils.PtrTo(protos.TraceStatus(protos.TraceStatus_value[ti.Status])),
		RequestMetadata: make([]*protos.TraceRequestMetadata, 0, len(ti.Tags)),
		Tags:            make([]*protos.TraceTag, 0, len(ti.Tags)),
	}

	for _, tag := range ti.Tags {
		traceInfo.Tags = append(traceInfo.Tags, tag.ToProto())
	}

	for _, metadata := range ti.TraceRequestMetadata {
		traceInfo.RequestMetadata = append(traceInfo.RequestMetadata, metadata.ToProto())
	}

	return &traceInfo
}
