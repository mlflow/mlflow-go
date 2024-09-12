package models

import (
	"github.com/mlflow/mlflow-go/pkg/entities"
	"github.com/mlflow/mlflow-go/pkg/protos"
)

// InputTag mapped from table <input_tags>.
type InputTag struct {
	Key     string `gorm:"column:name;primaryKey"`
	Value   string `gorm:"column:value;not null"`
	InputID string `gorm:"column:input_uuid;primaryKey"`
}

func (i *InputTag) ToEntity() *entities.InputTag {
	return &entities.InputTag{
		Key:   i.Key,
		Value: i.Value,
	}
}

func (i *InputTag) ToProto() *protos.InputTag {
	return &protos.InputTag{
		Key:   &i.Key,
		Value: &i.Value,
	}
}
