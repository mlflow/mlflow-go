package models

import (
	"database/sql"

	"github.com/mlflow/mlflow-go/pkg/entities"
)

// Param mapped from table <params>.
type Param struct {
	Key   string         `db:"key"      gorm:"column:key;primaryKey"`
	Value sql.NullString `db:"value"    gorm:"column:value;not null"`
	RunID string         `db:"run_uuid" gorm:"column:run_uuid;primaryKey"`
}

func (p Param) ToEntity() *entities.Param {
	return &entities.Param{
		Key:   p.Key,
		Value: &p.Value.String,
	}
}

func NewParamFromEntity(runID string, param *entities.Param) Param {
	p := Param{
		Key:   param.Key,
		RunID: runID,
	}

	if param.Value != nil {
		p.Value = sql.NullString{Valid: true, String: *param.Value}
	}

	return p
}
