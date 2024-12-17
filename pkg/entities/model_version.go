package entities

import (
	"strconv"

	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/utils"
)

type ModelVersion struct {
	Name            string
	Version         int32
	CreationTime    int64
	LastUpdatedTime int64
	Description     string
	UserID          string
	CurrentStage    string
	Source          string
	RunID           string
	Status          string
	StatusMessage   string
	RunLink         string
	StorageLocation string
	Tags            []*ModelVersionTag
	Aliases         []*RegisteredModelAlias
}

func (mv ModelVersion) ToProto() *protos.ModelVersion {
	modelVersion := protos.ModelVersion{
		Name:                 utils.PtrTo(mv.Name),
		Version:              utils.PtrTo(strconv.Itoa(int(mv.Version))),
		CurrentStage:         utils.PtrTo(mv.CurrentStage),
		CreationTimestamp:    utils.PtrTo(mv.CreationTime),
		LastUpdatedTimestamp: utils.PtrTo(mv.LastUpdatedTime),
		Description:          utils.PtrTo(mv.Description),
		UserId:               utils.PtrTo(mv.UserID),
		Source:               utils.PtrTo(mv.Source),
		Status:               utils.PtrTo(protos.ModelVersionStatus(protos.ModelVersionStatus_value[mv.Status])),
		Tags:                 make([]*protos.ModelVersionTag, 0, len(mv.Tags)),
		RunLink:              utils.PtrTo(mv.RunLink),
	}

	if mv.RunID != "" {
		modelVersion.RunId = utils.PtrTo(mv.RunID)
	}

	if mv.StatusMessage != "" {
		modelVersion.StatusMessage = utils.PtrTo(mv.StatusMessage)
	}

	for _, tag := range mv.Tags {
		modelVersion.Tags = append(modelVersion.Tags, tag.ToProto())
	}

	for _, alias := range mv.Aliases {
		modelVersion.Aliases = append(modelVersion.Aliases, alias.Alias)
	}

	return &modelVersion
}
