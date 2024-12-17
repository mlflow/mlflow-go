package models

import "github.com/mlflow/mlflow-go/pkg/entities"

// TraceRequestMetadata mapped from table <trace_request_metadata>.
type TraceRequestMetadata struct {
	Key       string `gorm:"column:key;primaryKey"`
	Value     string `gorm:"column:value"`
	RequestID string `gorm:"column:request_id;primaryKey"`
}

func (trm TraceRequestMetadata) TableName() string {
	return "trace_request_metadata"
}

func (trm TraceRequestMetadata) ToEntity() *entities.TraceRequestMetadata {
	return &entities.TraceRequestMetadata{
		Key:       trm.Key,
		Value:     trm.Value,
		RequestID: trm.RequestID,
	}
}

func NewTraceRequestMetadataFromEntity(
	requestID string, metadata *entities.TraceRequestMetadata,
) TraceRequestMetadata {
	return TraceRequestMetadata{
		Key:       metadata.Key,
		Value:     metadata.Value,
		RequestID: requestID,
	}
}
