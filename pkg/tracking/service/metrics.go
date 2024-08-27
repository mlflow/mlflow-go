package service

import (
	"context"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/protos"
)

func (ts TrackingService) LogMetric(
	ctx context.Context,
	input *protos.LogMetric,
) (*protos.LogMetric_Response, *contract.Error) {
	metric := &protos.Metric{
		Key:       input.Key,
		Value:     input.Value,
		Timestamp: input.Timestamp,
		Step:      input.Step,
	}

	if err := ts.Store.LogMetric(ctx, input.GetRunId(), metric); err != nil {
		return nil, err
	}

	return &protos.LogMetric_Response{}, nil
}
