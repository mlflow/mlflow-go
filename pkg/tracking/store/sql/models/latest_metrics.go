package models

import (
	"github.com/mlflow/mlflow-go/pkg/entities"
)

// LatestMetric mapped from table <latest_metrics>.
type LatestMetric struct {
	Key       string  `gorm:"column:key;primaryKey"`
	Value     float64 `gorm:"column:value;not null"`
	Timestamp int64   `gorm:"column:timestamp"`
	Step      int64   `gorm:"column:step;not null"`
	IsNaN     bool    `gorm:"column:is_nan;not null"`
	RunID     string  `gorm:"column:run_uuid;primaryKey"`
}

func (lm LatestMetric) ToEntity() *entities.Metric {
	return &entities.Metric{
		Key:       lm.Key,
		Value:     lm.Value,
		Timestamp: lm.Timestamp,
		Step:      lm.Step,
		IsNaN:     lm.IsNaN,
	}
}
