package service

import (
	"context"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/entities"
	"github.com/mlflow/mlflow-go/pkg/protos"
)

func (ts TrackingService) LogInputs(
	ctx context.Context, input *protos.LogInputs,
) (*protos.LogInputs_Response, *contract.Error) {
	if len(input.GetDatasets()) == 0 {
		return &protos.LogInputs_Response{}, nil
	}

	datasets := make([]*entities.DatasetInput, 0, len(input.GetDatasets()))
	for _, d := range input.GetDatasets() {
		datasets = append(datasets, entities.NewDatasetInputFromProto(d))
	}

	if err := ts.Store.LogInputs(ctx, input.GetRunId(), datasets); err != nil {
		return nil, err
	}

	return &protos.LogInputs_Response{}, nil
}
