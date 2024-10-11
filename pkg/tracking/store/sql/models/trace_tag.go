package models

import "github.com/mlflow/mlflow-go/pkg/entities"

// TraceTag mapped from table <trace_tags>.
type TraceTag struct {
	Key       string `db:"key"        gorm:"column:key;primaryKey"`
	Value     string `db:"value"      gorm:"column:value"`
	RequestID string `db:"request_id" gorm:"column:request_id"`
}

func (t TraceTag) ToEntity() *entities.TraceTag {
	return &entities.TraceTag{
		Key:       t.Key,
		Value:     t.Value,
		RequestID: t.RequestID,
	}
}
