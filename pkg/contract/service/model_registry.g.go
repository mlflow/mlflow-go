// Code generated by mlflow/go/cmd/generate/main.go. DO NOT EDIT.

package service

import (
	"context"
	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/contract"
)

type ModelRegistryService interface {
	contract.Destroyer
	RenameRegisteredModel(ctx context.Context, input *protos.RenameRegisteredModel) (*protos.RenameRegisteredModel_Response, *contract.Error)
	UpdateRegisteredModel(ctx context.Context, input *protos.UpdateRegisteredModel) (*protos.UpdateRegisteredModel_Response, *contract.Error)
	DeleteRegisteredModel(ctx context.Context, input *protos.DeleteRegisteredModel) (*protos.DeleteRegisteredModel_Response, *contract.Error)
	GetLatestVersions(ctx context.Context, input *protos.GetLatestVersions) (*protos.GetLatestVersions_Response, *contract.Error)
}
