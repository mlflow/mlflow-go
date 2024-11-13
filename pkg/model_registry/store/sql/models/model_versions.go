package models

import (
	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/utils"
)

// ModelVersion mapped from table <model_versions>.
//
//revive:disable:exported
type ModelVersion struct {
	Name            string            `gorm:"column:name;primaryKey"`
	Version         int32             `gorm:"column:version;primaryKey"`
	CreationTime    int64             `gorm:"column:creation_time"`
	LastUpdatedTime int64             `gorm:"column:last_updated_time"`
	Description     string            `gorm:"column:description"`
	UserID          string            `gorm:"column:user_id"`
	CurrentStage    ModelVersionStage `gorm:"column:current_stage"`
	Source          string            `gorm:"column:source"`
	RunID           string            `gorm:"column:run_id"`
	Status          string            `gorm:"column:status"`
	StatusMessage   string            `gorm:"column:status_message"`
	RunLink         string            `gorm:"column:run_link"`
	StorageLocation string            `gorm:"column:storage_location"`
}

const StageDeletedInternal = "Deleted_Internal"

func (mv ModelVersion) ToProto() *protos.ModelVersion {
	var status *protos.ModelVersionStatus
	if s, ok := protos.ModelVersionStatus_value[mv.Status]; ok {
		status = utils.PtrTo(protos.ModelVersionStatus(s))
	}

	return &protos.ModelVersion{
		Name:                 &mv.Name,
		Version:              utils.ConvertInt32PointerToStringPointer(&mv.Version),
		CreationTimestamp:    &mv.CreationTime,
		LastUpdatedTimestamp: &mv.LastUpdatedTime,
		UserId:               &mv.UserID,
		CurrentStage:         utils.PtrTo(mv.CurrentStage.String()),
		Description:          &mv.Description,
		Source:               &mv.Source,
		RunId:                &mv.RunID,
		Status:               status,
		StatusMessage:        &mv.StatusMessage,
		RunLink:              &mv.RunLink,
	}
}
