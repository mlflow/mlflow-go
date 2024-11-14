package service

import (
	"context"
	"fmt"

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

func (ts TrackingService) LogParam(
	ctx context.Context, input *protos.LogParam,
) (*protos.LogParam_Response, *contract.Error) {
	if err := ts.Store.LogParam(ctx, input.GetRunId(), entities.ParamFromLogMetricProtoInput(input)); err != nil {
		return nil, err
	}

	return &protos.LogParam_Response{}, nil
}

func (ts TrackingService) GetMetricHistory(
	ctx context.Context, input *protos.GetMetricHistory,
) (*protos.GetMetricHistory_Response, *contract.Error) {
	if input.PageToken != nil {
		//nolint:lll
		return nil, contract.NewError(
			protos.ErrorCode_INVALID_PARAMETER_VALUE,
			fmt.Sprintf(
				"The SQLAlchemyStore backend does not support pagination for the `get_metric_history` API. Supplied argument `page_token` '%s' must be `None`.",
				*input.PageToken,
			),
		)
	}

	runID := input.GetRunId()
	if input.RunUuid != nil {
		runID = input.GetRunUuid()
	}

	metrics, err := ts.Store.GetMetricHistory(ctx, runID, input.GetMetricKey())
	if err != nil {
		return nil, err
	}

	response := protos.GetMetricHistory_Response{
		Metrics: make([]*protos.Metric, len(metrics)),
	}

	for i, metric := range metrics {
		response.Metrics[i] = metric.ToProto()
	}

	return &response, nil
}
