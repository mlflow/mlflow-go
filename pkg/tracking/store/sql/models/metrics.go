package models

import (
	"math"

	"github.com/mlflow/mlflow-go/pkg/entities"
)

// Metric mapped from table <metrics>.
type Metric struct {
	Key       string  `gorm:"column:key;primaryKey"`
	Value     float64 `gorm:"column:value;primaryKey"`
	Timestamp int64   `gorm:"column:timestamp;primaryKey"`
	RunID     string  `gorm:"column:run_uuid;primaryKey"`
	Step      int64   `gorm:"column:step;primaryKey"`
	IsNaN     bool    `gorm:"column:is_nan;primaryKey"`
}

func NewMetricFromEntity(runID string, metric *entities.Metric) *Metric {
	model := Metric{
		RunID:     runID,
		Key:       metric.Key,
		Timestamp: metric.Timestamp,
	}

	if metric.Step != 0 {
		model.Step = metric.Step
	}

	switch {
	case math.IsNaN(metric.Value):
		model.Value = 0
		model.IsNaN = true
	case math.IsInf(metric.Value, 0):
		// NB: SQL cannot represent Infs => We replace +/- Inf with max/min 64b float value
		if metric.Value > 0 {
			model.Value = math.MaxFloat64
		} else {
			model.Value = -math.MaxFloat64
		}
	default:
		model.Value = metric.Value
	}

	return &model
}

func (m Metric) ToEntity() *entities.Metric {
	return &entities.Metric{
		Key:       m.Key,
		Value:     m.Value,
		Timestamp: m.Timestamp,
		Step:      m.Step,
		IsNaN:     m.IsNaN,
	}
}

func (m Metric) NewLatestMetricFromProto() LatestMetric {
	return LatestMetric{
		RunID:     m.RunID,
		Key:       m.Key,
		Value:     m.Value,
		Timestamp: m.Timestamp,
		Step:      m.Step,
		IsNaN:     m.IsNaN,
	}
}
