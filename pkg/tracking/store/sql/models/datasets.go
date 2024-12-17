package models

import (
	"github.com/mlflow/mlflow-go/pkg/entities"
)

// Dataset mapped from table <datasets>.
type Dataset struct {
	ID           string `gorm:"column:dataset_uuid;not null"`
	ExperimentID int32  `gorm:"column:experiment_id;primaryKey"`
	Name         string `gorm:"column:name;primaryKey"`
	Digest       string `gorm:"column:digest;primaryKey"`
	SourceType   string `gorm:"column:dataset_source_type;not null"`
	Source       string `gorm:"column:dataset_source;not null"`
	Schema       string `gorm:"column:dataset_schema"`
	Profile      string `gorm:"column:dataset_profile"`
}

func (d *Dataset) ToEntity() *entities.Dataset {
	return &entities.Dataset{
		Name:       d.Name,
		Digest:     d.Digest,
		SourceType: d.SourceType,
		Source:     d.Source,
		Schema:     d.Schema,
		Profile:    d.Profile,
	}
}
