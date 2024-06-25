// Code generated by mlflow/go/cmd/generate/main.go. DO NOT EDIT.

package service

import (
	"github.com/mlflow/mlflow-go/mlflow_go/go/protos"
	"github.com/mlflow/mlflow-go/mlflow_go/go/contract"
)

type TrackingService interface {
	GetExperimentByName(input *protos.GetExperimentByName) (*protos.GetExperimentByName_Response, *contract.Error)
	CreateExperiment(input *protos.CreateExperiment) (*protos.CreateExperiment_Response, *contract.Error)
	GetExperiment(input *protos.GetExperiment) (*protos.GetExperiment_Response, *contract.Error)
	DeleteExperiment(input *protos.DeleteExperiment) (*protos.DeleteExperiment_Response, *contract.Error)
	CreateRun(input *protos.CreateRun) (*protos.CreateRun_Response, *contract.Error)
	SearchRuns(input *protos.SearchRuns) (*protos.SearchRuns_Response, *contract.Error)
	LogBatch(input *protos.LogBatch) (*protos.LogBatch_Response, *contract.Error)
}
