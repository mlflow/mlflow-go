package models

import (
	"strconv"

	"github.com/mlflow/mlflow-go/pkg/entities"
)

// Experiment mapped from table <experiments>.
type Experiment struct {
	ID               int32          `gorm:"column:experiment_id;primaryKey;autoIncrement:true"`
	Name             string         `gorm:"column:name;not null"`
	ArtifactLocation string         `gorm:"column:artifact_location"`
	LifecycleStage   LifecycleStage `gorm:"column:lifecycle_stage"`
	CreationTime     int64          `gorm:"column:creation_time"`
	LastUpdateTime   int64          `gorm:"column:last_update_time"`
	Tags             []ExperimentTag
	Runs             []Run
}

func (e Experiment) ToEntity() *entities.Experiment {
	experiment := entities.Experiment{
		ExperimentID:     strconv.Itoa(int(e.ID)),
		Name:             e.Name,
		ArtifactLocation: e.ArtifactLocation,
		LifecycleStage:   e.LifecycleStage.String(),
		CreationTime:     e.CreationTime,
		LastUpdateTime:   e.LastUpdateTime,
		Tags:             make([]*entities.ExperimentTag, len(e.Tags)),
	}

	for i, tag := range e.Tags {
		experiment.Tags[i] = &entities.ExperimentTag{
			Key:   tag.Key,
			Value: tag.Value,
		}
	}

	return &experiment
}
