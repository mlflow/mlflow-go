package models

import (
	"database/sql"

	"github.com/mlflow/mlflow-go/pkg/entities"
)

const (
	TraceInfoStatusUnspecified = "TRACE_STATUS_UNSPECIFIED"
	TraceInfoStatusOk          = "OK"
	TraceInfoStatusError       = "ERROR"
	TraceInfoStatusInProgress  = "IN_PROGRESS"
)

// TraceInfo mapped from table <trace_info>.
type TraceInfo struct {
	RequestID            string                 `gorm:"column:request_id;primaryKey"`
	ExperimentID         string                 `gorm:"column:experiment_id"`
	TimestampMS          int64                  `gorm:"column:timestamp_ms"`
	ExecutionTimeMS      sql.NullInt64          `gorm:"column:execution_time_ms"`
	Status               string                 `gorm:"column:status"`
	Tags                 []TraceTag             `gorm:"foreignKey:RequestID"`
	TraceRequestMetadata []TraceRequestMetadata `gorm:"foreignKey:RequestID"`
}

func (ti TraceInfo) TableName() string {
	return "trace_info"
}

func (ti TraceInfo) ToEntity() *entities.TraceInfo {
	traceInfo := entities.TraceInfo{
		RequestID:            ti.RequestID,
		Status:               ti.Status,
		ExperimentID:         ti.ExperimentID,
		TimestampMS:          ti.TimestampMS,
		Tags:                 make([]*entities.TraceTag, 0, len(ti.Tags)),
		TraceRequestMetadata: make([]*entities.TraceRequestMetadata, 0, len(ti.TraceRequestMetadata)),
	}

	if ti.ExecutionTimeMS.Valid {
		traceInfo.ExecutionTimeMS = &ti.ExecutionTimeMS.Int64
	}

	for _, tag := range ti.Tags {
		traceInfo.Tags = append(traceInfo.Tags, tag.ToEntity())
	}

	for _, metadata := range ti.TraceRequestMetadata {
		traceInfo.TraceRequestMetadata = append(traceInfo.TraceRequestMetadata, metadata.ToEntity())
	}

	return &traceInfo
}
