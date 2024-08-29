package service

import (
	"context"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/utils"
)

func (ts TrackingService) LogMetric(
	ctx context.Context,
	input *protos.LogMetric,
) (*protos.LogMetric_Response, *contract.Error) {
	metric := &protos.Metric{
		Key:       utils.PtrTo(input.GetKey()),
		Value:     utils.PtrTo(input.GetValue()),
		Timestamp: utils.PtrTo(input.GetTimestamp()),
		Step:      utils.PtrTo(input.GetStep()),
	}

	if err := ts.Store.LogMetric(ctx, input.GetRunId(), metric); err != nil {
		return nil, err
	}

	return &protos.LogMetric_Response{}, nil
}
