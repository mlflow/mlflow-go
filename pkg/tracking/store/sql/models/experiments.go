package models

import (
	"strconv"
	"time"

	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/utils"
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

//nolint:exportloopref
func (e Experiment) ToProto() *protos.Experiment {
	experimentID := strconv.FormatInt(int64(e.ID), 10)
	tags := make([]*protos.ExperimentTag, len(e.Tags))

	for i, tag := range e.Tags {
		tags[i] = &protos.ExperimentTag{
			Key:   &tag.Key,
			Value: &tag.Value,
		}
	}

	return &protos.Experiment{
		ExperimentId:     &experimentID,
		Name:             &e.Name,
		ArtifactLocation: &e.ArtifactLocation,
		LifecycleStage:   utils.PtrTo(e.LifecycleStage.String()),
		CreationTime:     &e.CreationTime,
		LastUpdateTime:   &e.LastUpdateTime,
		Tags:             tags,
	}
}

func NewExperimentFromProto(proto *protos.CreateExperiment) Experiment {
	tags := make([]ExperimentTag, len(proto.GetTags()))
	for i, tag := range proto.GetTags() {
		tags[i] = ExperimentTag{
			Key:   tag.GetKey(),
			Value: tag.GetValue(),
		}
	}

	return Experiment{
		Name:             proto.GetName(),
		ArtifactLocation: proto.GetArtifactLocation(),
		LifecycleStage:   LifecycleStageActive,
		CreationTime:     time.Now().UnixMilli(),
		LastUpdateTime:   time.Now().UnixMilli(),
		Tags:             tags,
	}
}
