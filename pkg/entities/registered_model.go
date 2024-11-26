package entities

import (
	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/utils"
)

type RegisteredModel struct {
	Name            string
	Tags            []*RegisteredModelTag
	Aliases         []*RegisteredModelAlias
	Versions        []*ModelVersion
	Description     *string
	CreationTime    int64
	LastUpdatedTime int64
}

func (m RegisteredModel) ToProto() *protos.RegisteredModel {
	registeredModel := protos.RegisteredModel{
		Name:                 utils.PtrTo(m.Name),
		Tags:                 make([]*protos.RegisteredModelTag, 0, len(m.Tags)),
		Description:          m.Description,
		CreationTimestamp:    utils.PtrTo(m.CreationTime),
		LastUpdatedTimestamp: utils.PtrTo(m.LastUpdatedTime),
	}

	for _, tag := range m.Tags {
		registeredModel.Tags = append(registeredModel.Tags, tag.ToProto())
	}

	for _, alias := range m.Aliases {
		registeredModel.Aliases = append(registeredModel.Aliases, alias.ToProto())
	}

	for _, version := range m.Versions {
		registeredModel.LatestVersions = append(registeredModel.LatestVersions, version.ToProto())
	}

	return &registeredModel
}
