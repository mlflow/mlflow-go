package service

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/tracking/store/sql/models"
)

// CreateExperiment implements TrackingService.
func (ts TrackingService) CreateExperiment(ctx context.Context, input *protos.CreateExperiment) (
	*protos.CreateExperiment_Response, *contract.Error,
) {
	if input.GetArtifactLocation() != "" {
		artifactLocation := strings.TrimRight(input.GetArtifactLocation(), "/")

		// We don't check the validation here as this was already covered in the validator.
		url, _ := url.Parse(artifactLocation)
		switch url.Scheme {
		case "file", "":
			path, err := filepath.Abs(url.Path)
			if err != nil {
				return nil, contract.NewError(
					protos.ErrorCode_INVALID_PARAMETER_VALUE,
					fmt.Sprintf("error getting absolute path: %v", err),
				)
			}

			if runtime.GOOS == "windows" {
				url.Scheme = "file"
				path = "/" + strings.ReplaceAll(path, "\\", "/")
			}

			url.Path = path
			artifactLocation = url.String()
		}

		input.ArtifactLocation = &artifactLocation
	}

	experimentID, err := ts.Store.CreateExperiment(ctx, input)
	if err != nil {
		return nil, err
	}

	response := protos.CreateExperiment_Response{
		ExperimentId: &experimentID,
	}

	return &response, nil
}

// GetExperiment implements TrackingService.
func (ts TrackingService) GetExperiment(ctx context.Context, input *protos.GetExperiment) (*protos.GetExperiment_Response, *contract.Error) {
	experiment, err := ts.Store.GetExperiment(ctx, input.GetExperimentId())
	if err != nil {
		return nil, err
	}

	response := protos.GetExperiment_Response{
		Experiment: experiment,
	}

	return &response, nil
}

func (ts TrackingService) DeleteExperiment(
	ctx context.Context, input *protos.DeleteExperiment,
) (*protos.DeleteExperiment_Response, *contract.Error) {
	err := ts.Store.DeleteExperiment(ctx, input.GetExperimentId())
	if err != nil {
		return nil, err
	}

	return &protos.DeleteExperiment_Response{}, nil
}

func (ts TrackingService) RestoreExperiment(
	ctx context.Context, input *protos.RestoreExperiment,
) (*protos.RestoreExperiment_Response, *contract.Error) {
	err := ts.Store.RestoreExperiment(ctx, input.GetExperimentId())
	if err != nil {
		return nil, err
	}
	return &protos.RestoreExperiment_Response{}, nil
}

func (ts TrackingService) UpdateExperiment(
	ctx context.Context, input *protos.UpdateExperiment,
) (*protos.UpdateExperiment_Response, *contract.Error) {
	experiment, err := ts.Store.GetExperiment(ctx, input.GetExperimentId())
	if err != nil {
		return nil, err
	}
	if *experiment.LifecycleStage != string(models.LifecycleStageActive) {
		return nil, contract.NewError(
			protos.ErrorCode_INVALID_STATE,
			"Cannot rename a non-active experiment.",
		)
	}
	if input.NewName != nil {
		experiment.Name = input.NewName
		if err := ts.Store.RenameExperiment(ctx, experiment); err != nil {
			return nil, err
		}
	}

	return &protos.UpdateExperiment_Response{}, nil
}

func (ts TrackingService) GetExperimentByName(
	ctx context.Context, input *protos.GetExperimentByName,
) (*protos.GetExperimentByName_Response, *contract.Error) {
	experiment, err := ts.Store.GetExperimentByName(ctx, input.GetExperimentName())
	if err != nil {
		return nil, err
	}

	response := protos.GetExperimentByName_Response{
		Experiment: experiment,
	}

	return &response, nil
}
