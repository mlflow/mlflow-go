package models

import (
	"database/sql"

	"github.com/mlflow/mlflow-go/pkg/entities"
	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/utils"
)

const (
	StageNone       = "None"
	StageStaging    = "Staging"
	StageProduction = "Production"
	StageArchived   = "Archived"
)

const (
	ModelVersionStatusReady = "READY"
)

// ModelVersion mapped from table <model_versions>.
//
//revive:disable:exported
type ModelVersion struct {
	Name            string            `db:"name"                              gorm:"column:name;primaryKey"`
	Version         int32             `db:"version"                           gorm:"column:version;primaryKey"`
	CreationTime    int64             `db:"creation_time"                     gorm:"column:creation_time"`
	LastUpdatedTime int64             `db:"last_updated_time"                 gorm:"column:last_updated_time"`
	Description     sql.NullString    `db:"description"                       gorm:"column:description"`
	UserID          string            `db:"user_id"                           gorm:"column:user_id"`
	CurrentStage    ModelVersionStage `db:"current_stage"                     gorm:"column:current_stage"`
	Source          string            `db:"source"                            gorm:"column:source"`
	RunID           sql.NullString    `db:"run_id"                            gorm:"column:run_id"`
	Status          string            `db:"status"                            gorm:"column:status"`
	StatusMessage   sql.NullString    `db:"status_message"                    gorm:"column:status_message"`
	RunLink         string            `db:"run_link"                          gorm:"column:run_link"`
	StorageLocation string            `db:"storage_location"                  gorm:"column:storage_location"`
	Tags            []ModelVersionTag `gorm:"foreignKey:Name;references:Name"`
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
		Description:          &mv.Description.String,
		Source:               &mv.Source,
		RunId:                &mv.RunID.String,
		Status:               status,
		StatusMessage:        &mv.StatusMessage.String,
		RunLink:              &mv.RunLink,
	}
}

func (mv ModelVersion) ToEntity() *entities.ModelVersion {
	modelVersion := entities.ModelVersion{
		Tags:            make([]*entities.ModelVersionTag, 0, len(mv.Tags)),
		Name:            mv.Name,
		Version:         mv.Version,
		CreationTime:    mv.CreationTime,
		LastUpdatedTime: mv.LastUpdatedTime,
		Description:     mv.Description.String,
		UserID:          mv.UserID,
		CurrentStage:    mv.CurrentStage.String(),
		Source:          mv.Source,
		Status:          mv.Status,
		StatusMessage:   mv.StatusMessage.String,
		RunLink:         mv.RunLink,
		StorageLocation: mv.StorageLocation,
	}

	if mv.RunID.Valid {
		modelVersion.RunID = &mv.RunID.String
	}

	for _, tag := range mv.Tags {
		modelVersion.Tags = append(modelVersion.Tags, tag.ToEntity())
	}

	return &modelVersion
}
