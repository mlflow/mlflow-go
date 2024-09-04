package service

import (
	"context"
	"fmt"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/tracking/store/sql/models"
	"github.com/mlflow/mlflow-go/pkg/utils"
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

	page, err := ts.Store.SearchRuns(
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
		Runs:          page.Items,
		NextPageToken: page.NextPageToken,
	}

	return &response, nil
}

func (ts TrackingService) LogBatch(
	ctx context.Context, input *protos.LogBatch,
) (*protos.LogBatch_Response, *contract.Error) {
	err := ts.Store.LogBatch(ctx, input.GetRunId(), input.GetMetrics(), input.GetParams(), input.GetTags())
	if err != nil {
		return nil, err
	}

	return &protos.LogBatch_Response{}, nil
}

func (ts TrackingService) CreateRun(
	ctx context.Context, input *protos.CreateRun,
) (*protos.CreateRun_Response, *contract.Error) {
	run, err := ts.Store.CreateRun(ctx, input)
	if err != nil {
		return nil, err
	}

	return &protos.CreateRun_Response{Run: run}, nil
}

func (ts TrackingService) UpdateRun(
	ctx context.Context, input *protos.UpdateRun,
) (*protos.UpdateRun_Response, *contract.Error) {
	run, err := ts.Store.GetRun(ctx, input.GetRunId())
	if err != nil {
		return nil, err
	}

	if *run.Info.LifecycleStage != string(models.LifecycleStageActive) {
		return nil, contract.NewError(
			protos.ErrorCode_INVALID_STATE,
			fmt.Sprintf(
				"The run %s must be in the 'active' state. Current state is %s.",
				input.GetRunUuid(),
				*run.Info.LifecycleStage,
			),
		)
	}

	if status := input.GetStatus(); status != 0 {
		run.Info.Status = utils.PtrTo(status)
	}

	if endTime := input.GetEndTime(); endTime != 0 {
		run.Info.EndTime = utils.PtrTo(endTime)
	}

	if runName := input.GetRunName(); runName != "" {
		run.Info.RunName = utils.PtrTo(runName)

		runTag, err := ts.Store.GetRunTag(ctx, input.GetRunId(), utils.TagRunName)
		if err != nil {
			return nil, err
		}

		if runTag == nil {
			run.Data.Tags = append(run.Data.Tags, &protos.RunTag{
				Key:   utils.PtrTo(utils.TagRunName),
				Value: &runName,
			})
		}
	}

	if err := ts.Store.UpdateRun(ctx, run); err != nil {
		return nil, err
	}

	return &protos.UpdateRun_Response{RunInfo: run.Info}, nil
}
