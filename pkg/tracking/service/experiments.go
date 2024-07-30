package service

import (
	"fmt"
	"github.com/mlflow/mlflow-go/pkg/utils"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/protos"
)

// CreateExperiment implements TrackingService.
func (ts TrackingService) CreateExperiment(input *protos.CreateExperiment) (
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

	experimentID, err := ts.Store.CreateExperiment(input)
	if err != nil {
		return nil, err
	}

	response := protos.CreateExperiment_Response{
		ExperimentId: &experimentID,
	}

	return &response, nil
}

// GetExperiment implements TrackingService.
func (ts TrackingService) GetExperiment(input *protos.GetExperiment) (*protos.GetExperiment_Response, *contract.Error) {
	experiment, err := ts.Store.GetExperimentByID(input.GetExperimentId())
	if err != nil {
		return nil, err
	}

	response := protos.GetExperiment_Response{
		Experiment: experiment,
	}

	return &response, nil
}

func (ts TrackingService) DeleteExperiment(
	input *protos.DeleteExperiment,
) (*protos.DeleteExperiment_Response, *contract.Error) {
	err := ts.Store.DeleteExperiment(input.GetExperimentId())
	if err != nil {
		return nil, err
	}

	return &protos.DeleteExperiment_Response{}, nil
}

func (ts TrackingService) RestoreExperiment(
	input *protos.RestoreExperiment,
) (*protos.RestoreExperiment_Response, *contract.Error) {
	err := ts.Store.RestoreExperiment(input.GetExperimentId())
	if err != nil {
		return nil, err
	}
	return &protos.RestoreExperiment_Response{}, nil
}

func (ts TrackingService) UpdateExperiment(
	input *protos.UpdateExperiment,
) (*protos.UpdateExperiment_Response, *contract.Error) {
	experiment, err := ts.Store.GetExperimentByID(input.GetExperimentId())
	if err != nil {
		return nil, err
	}
	if input.NewName != nil {
		experiment.Name = utils.PtrTo(input.GetNewName())
		if err := ts.Store.UpdateExperiment(experiment); err != nil {
			return nil, err
		}
	}

	return &protos.UpdateExperiment_Response{}, nil
}

func (ts TrackingService) GetExperimentByName(
	input *protos.GetExperimentByName,
) (*protos.GetExperimentByName_Response, *contract.Error) {
	experiment, err := ts.Store.GetExperimentByName(input.GetExperimentName())
	if err != nil {
		return nil, err
	}

	response := protos.GetExperimentByName_Response{
		Experiment: experiment,
	}

	return &response, nil
}
