package models

import (
	"math"

	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/utils"
)

// Metric mapped from table <metrics>.
type Metric struct {
	Key       *string  `db:"key"       gorm:"column:key;primaryKey"`
	Value     *float64 `db:"value"     gorm:"column:value;primaryKey"`
	Timestamp *int64   `db:"timestamp" gorm:"column:timestamp;primaryKey"`
	RunID     *string  `db:"run_uuid"  gorm:"column:run_uuid;primaryKey"`
	Step      int64    `db:"step"      gorm:"column:step;primaryKey"`
	IsNan     *bool    `db:"is_nan"    gorm:"column:is_nan;primaryKey"`
}

func (m Metric) ToProto() *protos.Metric {
	return &protos.Metric{
		Key:       m.Key,
		Value:     m.Value,
		Timestamp: m.Timestamp,
		Step:      utils.PtrTo(m.Step),
	}
}

func NewMetricFromProto(runID string, metric *protos.Metric) *Metric {
	isNaN := math.IsNaN(metric.GetValue())

	var value float64

	switch {
	case isNaN:
		value = 0
	case math.IsInf(metric.GetValue(), 0):
		// NB: SQL cannot represent Infs => We replace +/- Inf with max/min 64b float value
		if metric.GetValue() > 0 {
			value = math.MaxFloat64
		} else {
			value = -math.MaxFloat64
		}
	default:
		value = metric.GetValue()
	}

	var step int64
	if metric.Step != nil {
		step = *metric.Step
	}

	return &Metric{
		RunID:     utils.PtrTo(runID),
		Key:       metric.Key,
		Value:     utils.PtrTo(value),
		Timestamp: metric.Timestamp,
		Step:      step,
		IsNan:     &isNaN,
	}
}

func (m Metric) NewLatestMetricFromProto() LatestMetric {
	return LatestMetric{
		RunID:     m.RunID,
		Key:       m.Key,
		Value:     m.Value,
		Timestamp: m.Timestamp,
		Step:      m.Step,
		IsNan:     m.IsNan,
	}
}
