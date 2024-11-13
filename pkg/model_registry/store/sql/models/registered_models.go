package models

import (
	"database/sql"

	"github.com/mlflow/mlflow-go/pkg/entities"
)

// RegisteredModel mapped from table <registered_models>.
type RegisteredModel struct {
	Name            string               `gorm:"column:name;primaryKey"`
	Tags            []RegisteredModelTag `gorm:"foreignKey:Name;references:Name"`
	Description     sql.NullString       `gorm:"column:description"`
	CreationTime    int64                `gorm:"column:creation_time"`
	LastUpdatedTime int64                `gorm:"column:last_updated_time"`
}

func (m *RegisteredModel) ToEntity() *entities.RegisteredModel {
	model := entities.RegisteredModel{
		Name:            m.Name,
		Tags:            make([]*entities.RegisteredModelTag, 0, len(m.Tags)),
		CreationTime:    m.CreationTime,
		LastUpdatedTime: m.LastUpdatedTime,
	}

	if m.Description.Valid {
		model.Description = &m.Description.String
	}

	for _, tag := range m.Tags {
		model.Tags = append(model.Tags, tag.ToEntity())
	}

	return &model
}
