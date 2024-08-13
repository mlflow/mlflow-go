package service

import (
	"github.com/gofiber/fiber/v2"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/protos"
)

func (m *ModelRegistryService) GetLatestVersions(
	_ *fiber.Ctx, input *protos.GetLatestVersions,
) (*protos.GetLatestVersions_Response, *contract.Error) {
	latestVersions, err := m.store.GetLatestVersions(input.GetName(), input.GetStages())
	if err != nil {
		return nil, err
	}

	return &protos.GetLatestVersions_Response{
		ModelVersions: latestVersions,
	}, nil
}
