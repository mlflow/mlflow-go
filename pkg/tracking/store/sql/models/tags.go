package models

import (
	"github.com/mlflow/mlflow-go/pkg/entities"
)

// Tag mapped from table <tags>.
type Tag struct {
	Key   string `db:"key"      gorm:"column:key;primaryKey"`
	Value string `db:"value"    gorm:"column:value"`
	RunID string `db:"run_uuid" gorm:"column:run_uuid;primaryKey"`
}

func (t Tag) ToEntity() *entities.RunTag {
	return &entities.RunTag{
		Key:   t.Key,
		Value: t.Value,
	}
}

func NewTagFromEntity(runID string, entity *entities.RunTag) Tag {
	tag := Tag{
		Key:   entity.Key,
		Value: entity.Value,
	}
	if runID != "" {
		tag.RunID = runID
	}

	return tag
}
