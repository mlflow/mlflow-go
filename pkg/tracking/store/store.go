package store

import (
	"context"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/entities"
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
		GetRun(ctx context.Context, runID string) (*entities.Run, *contract.Error)
		CreateRun(
			ctx context.Context,
			experimentID string,
			userID string,
			startTime int64,
			tags []*entities.RunTag,
			runName string,
		) (*entities.Run, *contract.Error)
		UpdateRun(
			ctx context.Context,
			runID string,
			runStatus string,
			endTime *int64,
			runName string,
		) *contract.Error
		DeleteRun(ctx context.Context, runID string) *contract.Error
		RestoreRun(ctx context.Context, runID string) *contract.Error

		GetRunTag(ctx context.Context, runID, tagKey string) (*entities.RunTag, *contract.Error)
	}
	MetricTrackingStore interface {
		LogBatch(
			ctx context.Context,
			runID string,
			metrics []*entities.Metric,
			params []*entities.Param,
			tags []*entities.RunTag) *contract.Error

		LogMetric(ctx context.Context, runID string, metric *entities.Metric) *contract.Error
	}
)

type ExperimentTrackingStore interface {
	// GetExperiment returns experiment by the experiment ID.
	// The experiment should contain the linked tags.
	GetExperiment(ctx context.Context, id string) (*entities.Experiment, *contract.Error)
	GetExperimentByName(ctx context.Context, name string) (*entities.Experiment, *contract.Error)

	CreateExperiment(
		ctx context.Context,
		name string,
		artifactLocation string,
		tags []*entities.ExperimentTag,
	) (string, *contract.Error)
	RestoreExperiment(ctx context.Context, id string) *contract.Error
	RenameExperiment(ctx context.Context, experimentID, name string) *contract.Error

	SearchRuns(
		ctx context.Context,
		experimentIDs []string,
		filter string,
		runViewType protos.ViewType,
		maxResults int,
		orderBy []string,
		pageToken string,
	) ([]*entities.Run, string, *contract.Error)

	DeleteExperiment(ctx context.Context, id string) *contract.Error
}
