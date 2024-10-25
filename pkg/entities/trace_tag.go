package entities

import "github.com/mlflow/mlflow-go/pkg/protos"

type TraceTag struct {
	Key       string
	Value     string
	RequestID string
}

func (tt TraceTag) ToProto() *protos.TraceTag {
	return &protos.TraceTag{
		Key:   &tt.Key,
		Value: &tt.Value,
	}
}

func TagsFromStartTraceProtoInput(protoTags []*protos.TraceTag) []*TraceTag {
	entityTags := make([]*TraceTag, 0, len(protoTags))
	for _, tag := range protoTags {
		entityTags = append(entityTags, &TraceTag{
			Key:   tag.GetKey(),
			Value: tag.GetValue(),
		})
	}

	return entityTags
}
