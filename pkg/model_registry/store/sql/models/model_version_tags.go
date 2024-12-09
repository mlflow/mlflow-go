package models

import "github.com/mlflow/mlflow-go/pkg/entities"

// ModelVersionTag mapped from table <model_version_tags>.
//
//revive:disable:exported
type ModelVersionTag struct {
	Key     string `db:"key"     gorm:"column:key;primaryKey"`
	Value   string `db:"value"   gorm:"column:value"`
	Name    string `db:"name"    gorm:"column:name;primaryKey"`
	Version int32  `db:"version" gorm:"column:version;primaryKey"`
}

func (mvt ModelVersionTag) ToEntity() *entities.ModelVersionTag {
	return &entities.ModelVersionTag{
		Key:     mvt.Key,
		Value:   mvt.Value,
		Name:    mvt.Name,
		Version: mvt.Version,
	}
}
