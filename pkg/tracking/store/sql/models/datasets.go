package models

import (
	"github.com/mlflow/mlflow-go/pkg/entities"
)

// Dataset mapped from table <datasets>.
type Dataset struct {
	ID           string `db:"dataset_uuid"        gorm:"column:dataset_uuid;not null"`
	ExperimentID int32  `db:"experiment_id"       gorm:"column:experiment_id;primaryKey"`
	Name         string `db:"name"                gorm:"column:name;primaryKey"`
	Digest       string `db:"digest"              gorm:"column:digest;primaryKey"`
	SourceType   string `db:"dataset_source_type" gorm:"column:dataset_source_type;not null"`
	Source       string `db:"dataset_source"      gorm:"column:dataset_source;not null"`
	Schema       string `db:"dataset_schema"      gorm:"column:dataset_schema"`
	Profile      string `db:"dataset_profile"     gorm:"column:dataset_profile"`
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
