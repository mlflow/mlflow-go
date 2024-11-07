package service

import (
	"context"
	"fmt"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/entities"
	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/utils"
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

func (ts TrackingService) EndTrace(
	ctx context.Context, input *protos.EndTrace,
) (*protos.EndTrace_Response, *contract.Error) {
	traceInfo, err := ts.Store.EndTrace(
		ctx,
		input.GetRequestId(),
		input.GetTimestampMs(),
		input.GetStatus().String(),
		entities.TraceRequestMetadataFromStartTraceProtoInput(input.GetRequestMetadata()),
		entities.TagsFromStartTraceProtoInput(input.GetTags()),
	)
	if err != nil {
		return nil, contract.NewError(protos.ErrorCode_INTERNAL_ERROR, fmt.Sprintf("error ending trace: %v", err))
	}

	return &protos.EndTrace_Response{
		TraceInfo: traceInfo.ToProto(),
	}, nil
}

func (ts TrackingService) GetTraceInfo(
	ctx context.Context, input *protos.GetTraceInfo,
) (*protos.GetTraceInfo_Response, *contract.Error) {
	traceInfo, err := ts.Store.GetTraceInfo(ctx, input.GetRequestId())
	if err != nil {
		return nil, err
	}

	return &protos.GetTraceInfo_Response{
		TraceInfo: traceInfo.ToProto(),
	}, nil
}

func (ts TrackingService) DeleteTraces(
	ctx context.Context, input *protos.DeleteTraces,
) (*protos.DeleteTraces_Response, *contract.Error) {
	if input.MaxTimestampMillis == nil && len(input.RequestIds) == 0 {
		return nil, contract.NewError(
			protos.ErrorCode_INVALID_PARAMETER_VALUE,
			"Either `max_timestamp_millis` or `request_ids` must be specified.",
		)
	}

	if input.MaxTimestampMillis != nil && input.RequestIds != nil {
		return nil, contract.NewError(
			protos.ErrorCode_INVALID_PARAMETER_VALUE,
			"Only one of `max_timestamp_millis` and `request_ids` can be specified.",
		)
	}

	if input.RequestIds != nil && input.MaxTraces != nil {
		return nil, contract.NewError(
			protos.ErrorCode_INVALID_PARAMETER_VALUE,
			"`max_traces` can't be specified if `request_ids` is specified.",
		)
	}

	if input.MaxTraces != nil && *input.MaxTraces <= 0 {
		return nil, contract.NewError(
			protos.ErrorCode_INVALID_PARAMETER_VALUE,
			fmt.Sprintf("`max_traces` must be a positive integer, received %d.", *input.MaxTraces),
		)
	}

	result, err := ts.Store.DeleteTraces(
		ctx,
		input.GetExperimentId(),
		input.GetMaxTimestampMillis(),
		input.GetMaxTraces(),
		input.GetRequestIds(),
	)
	if err != nil {
		return nil, err
	}

	return &protos.DeleteTraces_Response{
		TracesDeleted: utils.PtrTo(result),
	}, nil
}
