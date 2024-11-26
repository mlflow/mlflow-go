package entities

import (
	"strconv"

	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/utils"
)

type ModelVersion struct {
	Name            string
	Version         int32
	CreationTime    int64
	LastUpdatedTime int64
	Description     string
	UserID          string
	CurrentStage    string
	Source          string
	RunID           string
	Status          string
	StatusMessage   string
	RunLink         string
	StorageLocation string
}

func (mv ModelVersion) ToProto() *protos.ModelVersion {
	return &protos.ModelVersion{
		Version:      utils.PtrTo(strconv.Itoa(int(mv.Version))),
		CurrentStage: utils.PtrTo(mv.CurrentStage),
	}
}
