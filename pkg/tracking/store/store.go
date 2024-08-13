package store

import (
	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/protos"
)

//go:generate mockery
type TrackingStore interface {
	// GetExperiment returns experiment by the experiment ID.
	// The experiment should contain the linked tags.
	GetExperiment(id string) (*protos.Experiment, *contract.Error)
	GetExperimentByName(name string) (*protos.Experiment, *contract.Error)

	CreateExperiment(input *protos.CreateExperiment) (string, *contract.Error)
	RestoreExperiment(id string) *contract.Error
	RenameExperiment(experiment *protos.Experiment) *contract.Error

	SearchRuns(
		experimentIDs []string,
		filter string,
		runViewType protos.ViewType,
		maxResults int,
		orderBy []string,
		pageToken string,
	) (pagedList *PagedList[*protos.Run], err *contract.Error)

	DeleteExperiment(id string) *contract.Error

	LogBatch(
		runID string,
		metrics []*protos.Metric,
		params []*protos.Param,
		tags []*protos.RunTag) *contract.Error

	CreateRun(input *protos.CreateRun) (*protos.Run, *contract.Error)

	LogMetric(runID string, metric *protos.Metric) *contract.Error
}

type PagedList[T any] struct {
	Items         []T
	NextPageToken *string
}
