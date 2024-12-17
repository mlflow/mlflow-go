package models

import (
	"database/sql"

	"github.com/mlflow/mlflow-go/pkg/entities"
	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/utils"
)

const StageDeletedInternal = "Deleted_Internal"

// ModelVersion mapped from table <model_versions>.
//
//revive:disable:exported
type ModelVersion struct {
	Name            string                 `gorm:"column:name;primaryKey"`
	Version         int32                  `gorm:"column:version;primaryKey"`
	CreationTime    int64                  `gorm:"column:creation_time"`
	LastUpdatedTime int64                  `gorm:"column:last_updated_time"`
	Description     sql.NullString         `gorm:"column:description"`
	UserID          sql.NullString         `gorm:"column:user_id"`
	CurrentStage    ModelVersionStage      `gorm:"column:current_stage"`
	Source          string                 `gorm:"column:source"`
	RunID           string                 `gorm:"column:run_id"`
	Status          string                 `gorm:"column:status"`
	StatusMessage   sql.NullString         `gorm:"column:status_message"`
	RunLink         string                 `gorm:"column:run_link"`
	StorageLocation string                 `gorm:"column:storage_location"`
	Tags            []ModelVersionTag      `gorm:"foreignKey:Name,Version"`
	Aliases         []RegisteredModelAlias `gorm:"-"`
}

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
		UserId:               &mv.UserID.String,
		CurrentStage:         utils.PtrTo(mv.CurrentStage.String()),
		Description:          &mv.Description.String,
		Source:               &mv.Source,
		RunId:                &mv.RunID,
		Status:               status,
		StatusMessage:        &mv.StatusMessage.String,
		RunLink:              &mv.RunLink,
	}
}

func (mv ModelVersion) ToEntity() *entities.ModelVersion {
	modelVersion := entities.ModelVersion{
		Name:            mv.Name,
		Version:         mv.Version,
		CreationTime:    mv.CreationTime,
		LastUpdatedTime: mv.LastUpdatedTime,
		Description:     mv.Description.String,
		UserID:          mv.UserID.String,
		CurrentStage:    mv.CurrentStage.String(),
		Source:          mv.Source,
		RunID:           mv.RunID,
		Status:          mv.Status,
		StatusMessage:   mv.StatusMessage.String,
		RunLink:         mv.RunLink,
		StorageLocation: mv.StorageLocation,
		Tags:            make([]*entities.ModelVersionTag, 0, len(mv.Tags)),
		Aliases:         make([]*entities.RegisteredModelAlias, 0, len(mv.Aliases)),
	}

	for _, tag := range mv.Tags {
		modelVersion.Tags = append(modelVersion.Tags, tag.ToEntity())
	}

	for _, alias := range mv.Aliases {
		modelVersion.Aliases = append(modelVersion.Aliases, alias.ToEntity())
	}

	return &modelVersion
}
