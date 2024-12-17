package models

import (
	"database/sql"

	"github.com/mlflow/mlflow-go/pkg/entities"
)

// Param mapped from table <params>.
type Param struct {
	Key   string         `gorm:"column:key;primaryKey"`
	Value sql.NullString `gorm:"column:value;not null"`
	RunID string         `gorm:"column:run_uuid;primaryKey"`
}

func (p Param) ToEntity() *entities.Param {
	return &entities.Param{
		Key:   p.Key,
		Value: &p.Value.String,
	}
}

func NewParamFromEntity(runID string, entity *entities.Param) Param {
	param := Param{
		Key:   entity.Key,
		RunID: runID,
	}

	if entity.Value != nil {
		param.Value = sql.NullString{Valid: true, String: *entity.Value}
	}

	return param
}
