package models

import (
	"strconv"

	"github.com/mlflow/mlflow-go/pkg/entities"
)

// RegisteredModelAlias mapped from table <registered_model_aliases>.
type RegisteredModelAlias struct {
	Name    string `db:"name"    gorm:"column:name;primaryKey"`
	Alias   string `db:"alias"   gorm:"column:alias;primaryKey"`
	Version int32  `db:"version" gorm:"column:version;not null"`
}

func (a RegisteredModelAlias) ToEntity() *entities.RegisteredModelAlias {
	return &entities.RegisteredModelAlias{
		Alias:   a.Alias,
		Version: strconv.Itoa(int(a.Version)),
	}
}
