package models

import (
	"database/sql"

	"github.com/mlflow/mlflow-go/pkg/entities"
)

// RegisteredModel mapped from table <registered_models>.
type RegisteredModel struct {
	Name            string                 `gorm:"column:name;primaryKey"`
	Tags            []RegisteredModelTag   `gorm:"foreignKey:Name;references:Name"`
	Aliases         []RegisteredModelAlias `gorm:"foreignKey:Name;references:Name"`
	Versions        []ModelVersion         `gorm:"foreignKey:Name;references:Name"`
	Description     sql.NullString         `gorm:"column:description"`
	CreationTime    int64                  `gorm:"column:creation_time"`
	LastUpdatedTime int64                  `gorm:"column:last_updated_time"`
}

func (m *RegisteredModel) ToEntity() *entities.RegisteredModel {
	model := entities.RegisteredModel{
		Name:            m.Name,
		Tags:            make([]*entities.RegisteredModelTag, 0, len(m.Tags)),
		Aliases:         make([]*entities.RegisteredModelAlias, 0, len(m.Aliases)),
		Versions:        make([]*entities.ModelVersion, 0),
		CreationTime:    m.CreationTime,
		LastUpdatedTime: m.LastUpdatedTime,
	}

	if m.Description.Valid {
		model.Description = &m.Description.String
	}

	for _, tag := range m.Tags {
		model.Tags = append(model.Tags, tag.ToEntity())
	}

	for _, alias := range m.Aliases {
		model.Aliases = append(model.Aliases, alias.ToEntity())
	}

	latestVersionsByStage := map[string]*ModelVersion{}

	for _, currentVersion := range m.Versions {
		stage := currentVersion.CurrentStage.String()
		if stage != "Deleted_Internal" {
			if latestVersion, ok := latestVersionsByStage[stage]; !ok || latestVersion.Version < currentVersion.Version {
				latestVersionsByStage[stage] = &currentVersion
			}
		}
	}

	for _, latestVersion := range latestVersionsByStage {
		model.Versions = append(model.Versions, latestVersion.ToEntity())
	}

	return &model
}
