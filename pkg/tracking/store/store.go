package store

import (
	"context"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/entities"
	"github.com/mlflow/mlflow-go/pkg/protos"
)

//go:generate mockery
type TrackingStore interface {
	contract.Destroyer
	RunTrackingStore
	TraceTrackingStore
	MetricTrackingStore
	ExperimentTrackingStore
	InputTrackingStore
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
		DeleteTag(ctx context.Context, runID, key string) *contract.Error
		SetTag(ctx context.Context, runID, key, value string) *contract.Error
	}
	TraceTrackingStore interface {
		EndTrace(
			ctx context.Context,
			reqeustID string,
			timestampMS int64,
			status string,
			metadata []*entities.TraceRequestMetadata,
			tags []*entities.TraceTag,
		) (*entities.TraceInfo, error)
		GetTraceInfo(ctx context.Context, reqeustID string) (*entities.TraceInfo, *contract.Error)
		SetTraceTag(ctx context.Context, requestID, key, value string) error
		GetTraceTag(ctx context.Context, requestID, key string) (*entities.TraceTag, *contract.Error)
		DeleteTraceTag(ctx context.Context, tag *entities.TraceTag) *contract.Error
		DeleteTraces(
			ctx context.Context,
			experimentID string,
			maxTimestampMillis int64,
			maxTraces int32,
			requestIDs []string,
		) (int32, *contract.Error)
	}
	MetricTrackingStore interface {
		LogBatch(
			ctx context.Context,
			runID string,
			metrics []*entities.Metric,
			params []*entities.Param,
			tags []*entities.RunTag) *contract.Error

		LogMetric(ctx context.Context, runID string, metric *entities.Metric) *contract.Error
		LogParam(ctx context.Context, runID string, metric *entities.Param) *contract.Error
	}

	ExperimentTrackingStore interface {
		// GetExperiment returns experiment by the experiment ID.
		// The experiment should contain the linked tags.
		GetExperiment(ctx context.Context, id string) (*entities.Experiment, *contract.Error)
		GetExperimentByName(ctx context.Context, name string) (*entities.Experiment, *contract.Error)
		SearchExperiments(
			ctx context.Context,
			experimentViewType protos.ViewType,
			maxResults int64,
			filter string,
			orderBy []string,
			pageToken string,
		) ([]*entities.Experiment, string, *contract.Error)

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
		SetExperimentTag(ctx context.Context, experimentID, key, value string) *contract.Error
	}

	InputTrackingStore interface {
		LogInputs(ctx context.Context, runID string, datasets []*entities.DatasetInput) *contract.Error
	}
)
