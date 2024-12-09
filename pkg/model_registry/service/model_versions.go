package service

import (
	"context"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/entities"
	"github.com/mlflow/mlflow-go/pkg/protos"
)

func (m *ModelRegistryService) CreateModelVersion(
	ctx context.Context, input *protos.CreateModelVersion,
) (*protos.CreateModelVersion_Response, *contract.Error) {
	tags := make([]entities.ModelTag, 0, len(input.Tags))
	for _, tag := range input.Tags {
		tags = append(tags, entities.ModelTag{
			Key:   tag.GetKey(),
			Value: tag.GetValue(),
		})
	}

	modelVersion, err := m.store.CreateModelVersion(
		ctx,
		input.GetName(),
		input.GetSource(),
		input.GetRunId(),
		tags,
		input.GetRunLink(),
		input.GetDescription(),
	)
	if err != nil {
		return nil, err
	}

	return &protos.CreateModelVersion_Response{
		ModelVersion: modelVersion.ToProto(),
	}, nil
}

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

func (m *ModelRegistryService) RenameRegisteredModel(
	ctx context.Context, input *protos.RenameRegisteredModel,
) (*protos.RenameRegisteredModel_Response, *contract.Error) {
	newName := input.GetNewName()
	if newName == "" {
		return nil, contract.NewError(
			protos.ErrorCode_INVALID_PARAMETER_VALUE,
			"Registered model name cannot be empty",
		)
	}

	registeredModel, err := m.store.RenameRegisteredModel(ctx, input.GetName(), newName)
	if err != nil {
		return nil, err
	}

	return &protos.RenameRegisteredModel_Response{
		RegisteredModel: registeredModel.ToProto(),
	}, nil
}

func (m *ModelRegistryService) DeleteRegisteredModel(
	ctx context.Context, input *protos.DeleteRegisteredModel,
) (*protos.DeleteRegisteredModel_Response, *contract.Error) {
	if err := m.store.DeleteRegisteredModel(ctx, input.GetName()); err != nil {
		return nil, err
	}

	return &protos.DeleteRegisteredModel_Response{}, nil
}

func (m *ModelRegistryService) GetRegisteredModel(
	ctx context.Context, input *protos.GetRegisteredModel,
) (*protos.GetRegisteredModel_Response, *contract.Error) {
	registeredModel, err := m.store.GetRegisteredModel(ctx, input.GetName())
	if err != nil {
		return nil, err
	}

	return &protos.GetRegisteredModel_Response{
		RegisteredModel: registeredModel.ToProto(),
	}, nil
}
