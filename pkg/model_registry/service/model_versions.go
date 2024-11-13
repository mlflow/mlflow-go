package service

import (
	"context"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/entities"
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

func (m *ModelRegistryService) CreateRegisteredModel(
	ctx context.Context, input *protos.CreateRegisteredModel,
) (*protos.CreateRegisteredModel_Response, *contract.Error) {
	name := input.GetName()
	if name == "" {
		return nil, contract.NewError(
			protos.ErrorCode_INVALID_PARAMETER_VALUE,
			"Registered model name cannot be empty.",
		)
	}

	tags := make([]*entities.RegisteredModelTag, 0, len(input.GetTags()))
	for _, tag := range input.GetTags() {
		tags = append(tags, entities.NewRegisteredModelTagFromProto(tag))
	}

	registeredModel, err := m.store.CreateRegisteredModel(ctx, input.GetName(), input.GetDescription(), tags)
	if err != nil {
		return nil, err
	}

	return &protos.CreateRegisteredModel_Response{
		RegisteredModel: registeredModel.ToProto(),
	}, nil
}
