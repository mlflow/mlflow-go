package store

import (
	"context"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/protos"
)

//go:generate mockery
type TrackingStore interface {
	RunTrackingStore
	MetricTrackingStore
	ExperimentTrackingStore
}

type (
	RunTrackingStore interface {
		GetRun(ctx context.Context, runID string) (*protos.Run, *contract.Error)
		CreateRun(ctx context.Context, input *protos.CreateRun) (*protos.Run, *contract.Error)
		UpdateRun(ctx context.Context, run *protos.Run) *contract.Error
		DeleteRun(ctx context.Context, runID string) *contract.Error
		RestoreRun(ctx context.Context, runID string) *contract.Error

		GetRunTag(ctx context.Context, runID, tagKey string) (*protos.RunTag, *contract.Error)
	}
	MetricTrackingStore interface {
		LogBatch(
			ctx context.Context,
			runID string,
			metrics []*protos.Metric,
			params []*protos.Param,
			tags []*protos.RunTag) *contract.Error

		LogMetric(ctx context.Context, runID string, metric *protos.Metric) *contract.Error
	}
)

type ExperimentTrackingStore interface {
	// GetExperiment returns experiment by the experiment ID.
	// The experiment should contain the linked tags.
	GetExperiment(ctx context.Context, id string) (*protos.Experiment, *contract.Error)
	GetExperimentByName(ctx context.Context, name string) (*protos.Experiment, *contract.Error)

	CreateExperiment(ctx context.Context, input *protos.CreateExperiment) (string, *contract.Error)
	RestoreExperiment(ctx context.Context, id string) *contract.Error
	RenameExperiment(ctx context.Context, experiment *protos.Experiment) *contract.Error

	SearchRuns(
		ctx context.Context,
		experimentIDs []string,
		filter string,
		runViewType protos.ViewType,
		maxResults int,
		orderBy []string,
		pageToken string,
	) (pagedList *PagedList[*protos.Run], err *contract.Error)

	DeleteExperiment(ctx context.Context, id string) *contract.Error
}

type PagedList[T any] struct {
	Items         []T
	NextPageToken *string
}
