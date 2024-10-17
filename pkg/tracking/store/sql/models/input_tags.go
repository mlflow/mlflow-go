package models

import (
	"github.com/mlflow/mlflow-go/pkg/entities"
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

func NewInputTagFromEntity(inputID string, tag *entities.InputTag) *InputTag {
	return &InputTag{
		InputID: inputID,
		Key:     tag.Key,
		Value:   tag.Value,
	}
}
