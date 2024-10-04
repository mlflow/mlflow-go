package models

import (
	"github.com/mlflow/mlflow-go/pkg/entities"
)

// Param mapped from table <params>.
type Param struct {
	Key   string `db:"key"      gorm:"column:key;primaryKey"`
	Value string `db:"value"    gorm:"column:value;not null"`
	RunID string `db:"run_uuid" gorm:"column:run_uuid;primaryKey"`
}

func (p Param) ToEntity() *entities.Param {
	return &entities.Param{
		Key:   p.Key,
		Value: p.Value,
	}
}

func NewParamFromEntity(runID string, param *entities.Param) Param {
	return Param{
		Key:   param.Key,
		Value: param.Value,
		RunID: runID,
	}
}
