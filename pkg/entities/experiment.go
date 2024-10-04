package entities

import (
	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/utils"
)

type Experiment struct {
	Name             string
	ExperimentID     string
	ArtifactLocation string
	LifecycleStage   string
	LastUpdateTime   int64
	CreationTime     int64
	Tags             []*ExperimentTag
}

func (e Experiment) ToProto() *protos.Experiment {
	tags := make([]*protos.ExperimentTag, len(e.Tags))

	for i, tag := range e.Tags {
		tags[i] = tag.ToProto()
	}

	experiment := protos.Experiment{
		ExperimentId:     &e.ExperimentID,
		Name:             &e.Name,
		ArtifactLocation: &e.ArtifactLocation,
		LifecycleStage:   utils.PtrTo(e.LifecycleStage),
		CreationTime:     &e.CreationTime,
		LastUpdateTime:   &e.LastUpdateTime,
		Tags:             tags,
	}

	return &experiment
}
