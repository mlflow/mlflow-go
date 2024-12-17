package models

import "github.com/mlflow/mlflow-go/pkg/entities"

// TraceTag mapped from table <trace_tags>.
type TraceTag struct {
	Key       string `gorm:"column:key;primaryKey"`
	Value     string `gorm:"column:value"`
	RequestID string `gorm:"column:request_id;primaryKey"`
}

func (t TraceTag) ToEntity() *entities.TraceTag {
	return &entities.TraceTag{
		Key:       t.Key,
		Value:     t.Value,
		RequestID: t.RequestID,
	}
}

func NewTraceTagFromEntity(requestID string, entity *entities.TraceTag) TraceTag {
	return TraceTag{
		Key:       entity.Key,
		Value:     entity.Value,
		RequestID: requestID,
	}
}
