package service

import (
	"github.com/mlflow/mlflow-go/mlflow_go/go/contract"
	"github.com/mlflow/mlflow-go/mlflow_go/go/protos"
)

func (m *ModelRegistryService) GetLatestVersions(
	input *protos.GetLatestVersions,
) (*protos.GetLatestVersions_Response, *contract.Error) {
	latestVersions, err := m.store.GetLatestVersions(input.GetName(), input.GetStages())
	if err != nil {
		return nil, err
	}

	return &protos.GetLatestVersions_Response{
		ModelVersions: latestVersions,
	}, nil
}
