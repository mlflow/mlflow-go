package service

import (
	"context"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/protos"
)

func (ts TrackingService) SetTraceTag(
	ctx context.Context, input *protos.SetTraceTag,
) (*protos.SetTraceTag_Response, *contract.Error) {
	if err := ts.Store.SetTraceTag(
		ctx, input.GetRequestId(), input.GetKey(), input.GetValue(),
	); err != nil {
		return nil, contract.NewErrorWith(protos.ErrorCode_INTERNAL_ERROR, "failed to create trace_tag", err)
	}

	return &protos.SetTraceTag_Response{}, nil
}
