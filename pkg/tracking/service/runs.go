package service

import (
	"context"
	"fmt"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/entities"
	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/tracking/store/sql/models"
)

func (ts TrackingService) SearchRuns(
	ctx context.Context, input *protos.SearchRuns,
) (*protos.SearchRuns_Response, *contract.Error) {
	var runViewType protos.ViewType
	if input.RunViewType == nil {
		runViewType = protos.ViewType_ALL
	} else {
		runViewType = input.GetRunViewType()
	}

	maxResults := int(input.GetMaxResults())

	runs, nextPageToken, err := ts.Store.SearchRuns(
		ctx,
		input.GetExperimentIds(),
		input.GetFilter(),
		runViewType,
		maxResults,
		input.GetOrderBy(),
		input.GetPageToken(),
	)
	if err != nil {
		return nil, contract.NewError(protos.ErrorCode_INTERNAL_ERROR, fmt.Sprintf("error getting runs: %v", err))
	}

	response := protos.SearchRuns_Response{
		Runs:          make([]*protos.Run, len(runs)),
		NextPageToken: &nextPageToken,
	}

	for i, run := range runs {
		response.Runs[i] = run.ToProto()
	}

	return &response, nil
}

func (ts TrackingService) LogBatch(
	ctx context.Context, input *protos.LogBatch,
) (*protos.LogBatch_Response, *contract.Error) {
	metrics := make([]*entities.Metric, len(input.GetMetrics()))
	for i, metric := range input.GetMetrics() {
		metrics[i] = entities.MetricFromProto(metric)
	}

	params := make([]*entities.Param, len(input.GetParams()))
	for i, param := range input.GetParams() {
		params[i] = entities.ParamFromProto(param)
	}

	tags := make([]*entities.RunTag, len(input.GetTags()))
	for i, tag := range input.GetTags() {
		tags[i] = entities.NewTagFromProto(tag)
	}

	err := ts.Store.LogBatch(ctx, input.GetRunId(), metrics, params, tags)
	if err != nil {
		return nil, err
	}

	return &protos.LogBatch_Response{}, nil
}

func (ts TrackingService) GetRun(
	ctx context.Context, input *protos.GetRun,
) (*protos.GetRun_Response, *contract.Error) {
	run, err := ts.Store.GetRun(ctx, input.GetRunId())
	if err != nil {
		return nil, err
	}

	return &protos.GetRun_Response{Run: run.ToProto()}, nil
}

func (ts TrackingService) CreateRun(
	ctx context.Context, input *protos.CreateRun,
) (*protos.CreateRun_Response, *contract.Error) {
	tags := make([]*entities.RunTag, 0, len(input.GetTags()))
	for _, tag := range input.GetTags() {
		tags = append(tags, entities.NewTagFromProto(tag))
	}

	run, err := ts.Store.CreateRun(
		ctx,
		input.GetExperimentId(),
		input.GetUserId(),
		input.GetStartTime(),
		tags,
		input.GetRunName(),
	)
	if err != nil {
		return nil, err
	}

	return &protos.CreateRun_Response{Run: run.ToProto()}, nil
}

func (ts TrackingService) UpdateRun(
	ctx context.Context, input *protos.UpdateRun,
) (*protos.UpdateRun_Response, *contract.Error) {
	run, err := ts.Store.GetRun(ctx, input.GetRunId())
	if err != nil {
		return nil, err
	}

	if run.Info.LifecycleStage != string(models.LifecycleStageActive) {
		return nil, contract.NewError(
			protos.ErrorCode_INVALID_STATE,
			fmt.Sprintf(
				"The run %s must be in the 'active' state. Current state is %s.",
				input.GetRunUuid(),
				run.Info.LifecycleStage,
			),
		)
	}

	if status := input.GetStatus(); status != 0 {
		run.Info.Status = status.String()
	}

	if runName := input.GetRunName(); runName != "" {
		run.Info.RunName = runName
	}

	if err := ts.Store.UpdateRun(
		ctx,
		run.Info.RunID,
		run.Info.Status,
		input.EndTime,
		run.Info.RunName,
	); err != nil {
		return nil, err
	}

	return &protos.UpdateRun_Response{RunInfo: run.Info.ToProto()}, nil
}

func (ts TrackingService) DeleteRun(
	ctx context.Context, input *protos.DeleteRun,
) (*protos.DeleteRun_Response, *contract.Error) {
	if err := ts.Store.DeleteRun(ctx, input.GetRunId()); err != nil {
		return nil, err
	}

	return &protos.DeleteRun_Response{}, nil
}

func (ts TrackingService) RestoreRun(
	ctx context.Context, input *protos.RestoreRun,
) (*protos.RestoreRun_Response, *contract.Error) {
	if err := ts.Store.RestoreRun(ctx, input.GetRunId()); err != nil {
		return nil, err
	}

	return &protos.RestoreRun_Response{}, nil
}
