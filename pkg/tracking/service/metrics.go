package service

import (
	"context"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/entities"
	"github.com/mlflow/mlflow-go/pkg/protos"
)

func (ts TrackingService) LogMetric(
	ctx context.Context,
	input *protos.LogMetric,
) (*protos.LogMetric_Response, *contract.Error) {
	if err := ts.Store.LogMetric(ctx, input.GetRunId(), entities.MetricFromLogMetricProtoInput(input)); err != nil {
		return nil, err
	}

	return &protos.LogMetric_Response{}, nil
}
