package entities

import (
	"github.com/mlflow/mlflow-go/pkg/protos"
)

type Metric struct {
	Key       string
	Value     float64
	Timestamp int64
	Step      int64
}

func (m Metric) ToProto() *protos.Metric {
	return &protos.Metric{
		Key:       &m.Key,
		Value:     &m.Value,
		Timestamp: &m.Timestamp,
		Step:      &m.Step,
	}
}

func MetricFromProto(proto *protos.Metric) *Metric {
	return &Metric{
		Key:       proto.GetKey(),
		Value:     proto.GetValue(),
		Timestamp: proto.GetTimestamp(),
		Step:      proto.GetStep(),
	}
}

func MetricFromLogMetricProtoInput(input *protos.LogMetric) *Metric {
	return &Metric{
		Key:       input.GetKey(),
		Value:     input.GetValue(),
		Timestamp: input.GetTimestamp(),
		Step:      input.GetStep(),
	}
}
