// Code generated by mlflow/go/cmd/generate/main.go. DO NOT EDIT.

package service

import (
	"context"
	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/contract"
)

type TrackingService interface {
	GetExperimentByName(ctx context.Context, input *protos.GetExperimentByName) (*protos.GetExperimentByName_Response, *contract.Error)
	CreateExperiment(ctx context.Context, input *protos.CreateExperiment) (*protos.CreateExperiment_Response, *contract.Error)
	GetExperiment(ctx context.Context, input *protos.GetExperiment) (*protos.GetExperiment_Response, *contract.Error)
	DeleteExperiment(ctx context.Context, input *protos.DeleteExperiment) (*protos.DeleteExperiment_Response, *contract.Error)
	RestoreExperiment(ctx context.Context, input *protos.RestoreExperiment) (*protos.RestoreExperiment_Response, *contract.Error)
	UpdateExperiment(ctx context.Context, input *protos.UpdateExperiment) (*protos.UpdateExperiment_Response, *contract.Error)
	CreateRun(ctx context.Context, input *protos.CreateRun) (*protos.CreateRun_Response, *contract.Error)
	UpdateRun(ctx context.Context, input *protos.UpdateRun) (*protos.UpdateRun_Response, *contract.Error)
	LogMetric(ctx context.Context, input *protos.LogMetric) (*protos.LogMetric_Response, *contract.Error)
	GetRun(ctx context.Context, input *protos.GetRun) (*protos.GetRun_Response, *contract.Error)
	SearchRuns(ctx context.Context, input *protos.SearchRuns) (*protos.SearchRuns_Response, *contract.Error)
	LogBatch(ctx context.Context, input *protos.LogBatch) (*protos.LogBatch_Response, *contract.Error)
}
