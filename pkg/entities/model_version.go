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
	Aliases         []string
}

func (mv ModelVersion) ToProto() *protos.ModelVersion {
	modelVersion := protos.ModelVersion{
		Name:                 &mv.Name,
		Version:              utils.PtrTo(strconv.Itoa(int(mv.Version))),
		Description:          &mv.Description,
		CurrentStage:         &mv.CurrentStage,
		CreationTimestamp:    &mv.CreationTime,
		LastUpdatedTimestamp: &mv.LastUpdatedTime,
		UserId:               &mv.UserID,
		Source:               &mv.Source,
		RunId:                &mv.RunID,
		Status:               utils.PtrTo(protos.ModelVersionStatus(protos.ModelVersionStatus_value[mv.Status])),
		StatusMessage:        &mv.StatusMessage,
		RunLink:              &mv.RunLink,
		Aliases:              mv.Aliases,
	}

	for _, tag := range mv.Tags {
		modelVersion.Tags = append(modelVersion.Tags, tag.ToProto())
	}

	return &modelVersion
}
