package models

import "github.com/mlflow/mlflow-go/pkg/entities"

// ModelVersionTag mapped from table <model_version_tags>.
//
//revive:disable:exported
type ModelVersionTag struct {
	Key     string `gorm:"column:key;primaryKey"`
	Value   string `gorm:"column:value"`
	Name    string `gorm:"column:name;primaryKey"`
	Version int32  `gorm:"column:version;primaryKey"`
}

func (mvt ModelVersionTag) ToEntity() *entities.ModelVersionTag {
	return &entities.ModelVersionTag{
		Key:   mvt.Key,
		Value: mvt.Value,
	}
}
