package entities

import (
	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/utils"
)

type RunInfo struct {
	RunID          string
	RunUUID        string
	RunName        string
	ExperimentID   int32
	UserID         string
	Status         string
	StartTime      int64
	EndTime        *int64
	ArtifactURI    string
	LifecycleStage string
}

func (ri RunInfo) ToProto() *protos.RunInfo {
	return &protos.RunInfo{
		RunId:          &ri.RunID,
		RunUuid:        &ri.RunID,
		RunName:        &ri.RunName,
		ExperimentId:   utils.ConvertInt32PointerToStringPointer(&ri.ExperimentID),
		UserId:         &ri.UserID,
		Status:         RunStatusToProto(ri.Status),
		StartTime:      &ri.StartTime,
		EndTime:        ri.EndTime,
		ArtifactUri:    &ri.ArtifactURI,
		LifecycleStage: utils.PtrTo(ri.LifecycleStage),
	}
}
