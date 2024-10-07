package service

import (
	"context"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/protos"
)

func (ts TrackingService) DeleteTag(
	ctx context.Context, input *protos.DeleteTag,
) (*protos.DeleteTag_Response, *contract.Error) {
	if err := ts.Store.DeleteTag(ctx, input.GetRunId(), input.GetKey()); err != nil {
		return nil, err
	}

	return &protos.DeleteTag_Response{}, nil
}
