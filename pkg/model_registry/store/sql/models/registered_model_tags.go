package models

import "github.com/mlflow/mlflow-go/pkg/entities"

// RegisteredModelTag mapped from table <registered_model_tags>.
type RegisteredModelTag struct {
	Key   string `gorm:"column:key;primaryKey"`
	Name  string `gorm:"column:name;primaryKey"`
	Value string `gorm:"column:value"`
}

func (t RegisteredModelTag) ToEntity() *entities.RegisteredModelTag {
	return &entities.RegisteredModelTag{
		Key:   t.Key,
		Value: t.Value,
	}
}

func RegisteredModelTagFromEntity(name string, tag *entities.RegisteredModelTag) RegisteredModelTag {
	return RegisteredModelTag{
		Name:  name,
		Key:   tag.Key,
		Value: tag.Value,
	}
}
