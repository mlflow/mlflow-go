package service

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/protos"
)

func (ts TrackingService) SearchRuns(
	_ *fiber.Ctx, input *protos.SearchRuns,
) (*protos.SearchRuns_Response, *contract.Error) {
	var runViewType protos.ViewType
	if input.RunViewType == nil {
		runViewType = protos.ViewType_ALL
	} else {
		runViewType = input.GetRunViewType()
	}

	maxResults := int(input.GetMaxResults())

	page, err := ts.Store.SearchRuns(
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
	_ *fiber.Ctx, input *protos.LogBatch,
) (*protos.LogBatch_Response, *contract.Error) {
	err := ts.Store.LogBatch(input.GetRunId(), input.GetMetrics(), input.GetParams(), input.GetTags())
	if err != nil {
		return nil, err
	}

	return &protos.LogBatch_Response{}, nil
}

func (ts TrackingService) CreateRun(
	_ *fiber.Ctx, input *protos.CreateRun,
) (*protos.CreateRun_Response, *contract.Error) {
	run, err := ts.Store.CreateRun(input)
	if err != nil {
		return nil, err
	}

	return &protos.CreateRun_Response{Run: run}, nil
}
