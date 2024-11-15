package service

import (
	"context"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/protos"
)

func (m *ModelRegistryService) GetLatestVersions(
	ctx context.Context, input *protos.GetLatestVersions,
) (*protos.GetLatestVersions_Response, *contract.Error) {
	latestVersions, err := m.store.GetLatestVersions(ctx, input.GetName(), input.GetStages())
	if err != nil {
		return nil, err
	}

	return &protos.GetLatestVersions_Response{
		ModelVersions: latestVersions,
	}, nil
}

func (m *ModelRegistryService) UpdateRegisteredModel(
	ctx context.Context, input *protos.UpdateRegisteredModel,
) (*protos.UpdateRegisteredModel_Response, *contract.Error) {
	registeredModel, err := m.store.UpdateRegisteredModel(ctx, input.GetName(), input.GetDescription())
	if err != nil {
		return nil, err
	}

	return &protos.UpdateRegisteredModel_Response{
		RegisteredModel: registeredModel.ToProto(),
	}, nil
}
