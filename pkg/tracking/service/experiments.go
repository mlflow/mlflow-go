package service

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/entities"
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

	tags := make([]*entities.ExperimentTag, len(input.GetTags()))
	for i, tag := range input.GetTags() {
		tags[i] = entities.NewExperimentTagFromProto(tag)
	}

	experimentID, err := ts.Store.CreateExperiment(ctx, input.GetName(), input.GetArtifactLocation(), tags)
	if err != nil {
		return nil, err
	}

	return &protos.CreateExperiment_Response{
		ExperimentId: &experimentID,
	}, nil
}

// GetExperiment implements TrackingService.
func (ts TrackingService) GetExperiment(
	ctx context.Context, input *protos.GetExperiment,
) (*protos.GetExperiment_Response, *contract.Error) {
	experiment, err := ts.Store.GetExperiment(ctx, input.GetExperimentId())
	if err != nil {
		return nil, err
	}

	return &protos.GetExperiment_Response{
		Experiment: experiment.ToProto(),
	}, nil
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

	if experiment.LifecycleStage != string(models.LifecycleStageActive) {
		return nil, contract.NewError(
			protos.ErrorCode_INVALID_STATE,
			"Cannot rename a non-active experiment.",
		)
	}

	if name := input.GetNewName(); name != "" {
		if err := ts.Store.RenameExperiment(ctx, input.GetExperimentId(), input.GetNewName()); err != nil {
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

	return &protos.GetExperimentByName_Response{
		Experiment: experiment.ToProto(),
	}, nil
}

func (ts TrackingService) SetExperimentTag(
	ctx context.Context, input *protos.SetExperimentTag,
) (*protos.SetExperimentTag_Response, *contract.Error) {
	if err := ts.Store.SetExperimentTag(ctx, input.GetExperimentId(), input.GetKey(), input.GetValue()); err != nil {
		return nil, err
	}

	return &protos.SetExperimentTag_Response{}, nil
}

func (ts TrackingService) SearchExperiments(
	ctx context.Context, input *protos.SearchExperiments,
) (*protos.SearchExperiments_Response, *contract.Error) {
	experiments, nextPageToken, err := ts.Store.SearchExperiments(
		ctx,
		input.GetViewType(),
		input.GetMaxResults(),
		input.GetFilter(),
		input.GetOrderBy(),
		input.GetPageToken(),
	)
	if err != nil {
		return nil, contract.NewError(protos.ErrorCode_INTERNAL_ERROR, fmt.Sprintf("error getting experiments: %v", err))
	}

	response := protos.SearchExperiments_Response{
		Experiments: make([]*protos.Experiment, len(experiments)),
	}
	if nextPageToken != "" {
		response.NextPageToken = &nextPageToken
	}

	for i, experiment := range experiments {
		response.Experiments[i] = experiment.ToProto()
	}

	return &response, nil
}
