package service

import (
	"context"
	"fmt"

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

func (ts TrackingService) DeleteTraceTag(
	ctx context.Context, input *protos.DeleteTraceTag,
) (*protos.DeleteTraceTag_Response, *contract.Error) {
	tag, err := ts.Store.GetTraceTag(ctx, input.GetRequestId(), input.GetKey())
	if err != nil {
		return nil, contract.NewError(protos.ErrorCode_INTERNAL_ERROR, fmt.Sprintf("error getting trace tag: %v", err))
	}

	if tag == nil {
		return nil, contract.NewError(
			protos.ErrorCode_RESOURCE_DOES_NOT_EXIST,
			fmt.Sprintf(
				"No trace tag with key '%s' for trace with request_id '%s'",
				input.GetKey(),
				input.GetRequestId(),
			),
		)
	}

	if err := ts.Store.DeleteTraceTag(ctx, tag); err != nil {
		return nil, err
	}

	return &protos.DeleteTraceTag_Response{}, nil
}
