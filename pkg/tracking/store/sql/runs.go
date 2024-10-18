package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/entities"
	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/tracking/store/sql/models"
	"github.com/mlflow/mlflow-go/pkg/utils"
)

func checkRunIsActive(transaction *gorm.DB, runID string) *contract.Error {
	var run models.Run

	err := transaction.
		Model(&models.Run{}).
		Where("run_uuid = ?", runID).
		Select("lifecycle_stage").
		First(&run).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return contract.NewError(
				protos.ErrorCode_RESOURCE_DOES_NOT_EXIST,
				fmt.Sprintf("Run with id=%s not found", runID),
			)
		}

		return contract.NewErrorWith(
			protos.ErrorCode_INTERNAL_ERROR,
			fmt.Sprintf(
				"failed to get lifecycle stage for run %q",
				runID,
			),
			err,
		)
	}

	if run.LifecycleStage != models.LifecycleStageActive {
		return contract.NewError(
			protos.ErrorCode_INVALID_PARAMETER_VALUE,
			fmt.Sprintf(
				"The run %s must be in the 'active' state.\n"+
					"Current state is %v.",
				runID,
				run.LifecycleStage,
			),
		)
	}

	return nil
}

func (s TrackingSQLStore) SearchRuns(
	ctx context.Context,
	experimentIDs []string, filter string,
	runViewType protos.ViewType, maxResults int, orderBy []string, pageToken string,
) ([]*entities.Run, string, *contract.Error) {
	// ViewType
	transaction := s.db.WithContext(ctx).Where(
		"runs.experiment_id IN ?", experimentIDs,
	).Where(
		"runs.lifecycle_stage IN ?", applyLifecycleStagesFilter(runViewType),
	)

	// MaxResults
	transaction.Limit(maxResults)

	// PageToken
	offset, contractError := getOffset(pageToken)
	if contractError != nil {
		return nil, "", contractError
	}

	transaction.Offset(offset)

	// Filter
	contractError = applyFilter(ctx, s.db, transaction, filter)
	if contractError != nil {
		return nil, "", contractError
	}

	// OrderBy
	contractError = applyOrderBy(ctx, s.db, transaction, orderBy)
	if contractError != nil {
		return nil, "", contractError
	}

	// Actual query
	var runs []models.Run

	transaction.Preload("LatestMetrics").Preload("Params").Preload("Tags").
		Preload("Inputs", "inputs.destination_type = 'RUN'").
		Preload("Inputs.Dataset").Preload("Inputs.Tags").Find(&runs)

	if transaction.Error != nil {
		return nil, "", contract.NewErrorWith(
			protos.ErrorCode_INTERNAL_ERROR,
			"Failed to query search runs",
			transaction.Error,
		)
	}

	entityRuns := make([]*entities.Run, len(runs))
	for i, run := range runs {
		entityRuns[i] = run.ToEntity()
	}

	nextPageToken, contractError := mkNextPageToken(len(runs), maxResults, offset)
	if contractError != nil {
		return nil, "", contractError
	}

	return entityRuns, nextPageToken, nil
}

const RunIDMaxLength = 32

const (
	ArtifactFolderName  = "artifacts"
	RunNameIntegerScale = 3
	RunNameMaxLength    = 20
)

func getRunNameFromTags(tags []models.Tag) string {
	for _, tag := range tags {
		if tag.Key == utils.TagRunName {
			return tag.Value
		}
	}

	return ""
}

func ensureRunName(runModel *models.Run) *contract.Error {
	runNameFromTags := getRunNameFromTags(runModel.Tags)

	switch {
	// run_name and name in tags differ
	case utils.IsNotNilOrEmptyString(&runModel.Name) && runNameFromTags != "" && runModel.Name != runNameFromTags:
		return contract.NewError(
			protos.ErrorCode_INVALID_PARAMETER_VALUE,
			fmt.Sprintf(
				"Both 'run_name' argument and 'mlflow.runName' tag are specified, but with "+
					"different values (run_name='%s', mlflow.runName='%s').",
				runModel.Name,
				runNameFromTags,
			),
		)
	// no name was provided, generate a random name
	case utils.IsNilOrEmptyString(&runModel.Name) && runNameFromTags == "":
		randomName, err := utils.GenerateRandomName()
		if err != nil {
			return contract.NewErrorWith(
				protos.ErrorCode_INTERNAL_ERROR,
				"failed to generate random run name",
				err,
			)
		}

		runModel.Name = randomName
	// use name from tags
	case utils.IsNilOrEmptyString(&runModel.Name) && runNameFromTags != "":
		runModel.Name = runNameFromTags
	}

	if runNameFromTags == "" {
		runModel.Tags = append(runModel.Tags, models.Tag{
			Key:   utils.TagRunName,
			Value: runModel.Name,
		})
	}

	return nil
}

func (s TrackingSQLStore) GetRun(ctx context.Context, runID string) (*entities.Run, *contract.Error) {
	var run models.Run
	if err := s.db.WithContext(ctx).Where(
		"run_uuid = ?", runID,
	).Preload(
		"Tags",
	).Preload(
		"Params",
	).Preload(
		"Inputs.Tags",
	).Preload(
		"LatestMetrics",
	).Preload(
		"Inputs.Dataset",
	).First(&run).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, contract.NewError(
				protos.ErrorCode_RESOURCE_DOES_NOT_EXIST,
				fmt.Sprintf("Run with id=%s not found", runID),
			)
		}

		return nil, contract.NewErrorWith(protos.ErrorCode_INTERNAL_ERROR, "failed to get run", err)
	}

	return run.ToEntity(), nil
}

//nolint:funlen
func (s TrackingSQLStore) CreateRun(
	ctx context.Context,
	experimentID, userID string,
	startTime int64,
	tags []*entities.RunTag,
	runName string,
) (*entities.Run, *contract.Error) {
	experiment, err := s.GetExperiment(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	if err := checkExperimentIsActive(experiment); err != nil {
		return nil, err
	}

	runModel := &models.Run{
		ID:             utils.NewUUID(),
		Name:           runName,
		ExperimentID:   utils.ConvertStringPointerToInt32Pointer(&experimentID),
		StartTime:      startTime,
		UserID:         userID,
		Tags:           make([]models.Tag, 0, len(tags)),
		LifecycleStage: models.LifecycleStageActive,
		Status:         models.RunStatusRunning,
		SourceType:     models.SourceTypeUnknown,
	}

	for _, tag := range tags {
		runModel.Tags = append(runModel.Tags, models.NewTagFromEntity(runModel.ID, tag))
	}

	artifactLocation, appendErr := utils.AppendToURIPath(
		experiment.ArtifactLocation,
		runModel.ID,
		ArtifactFolderName,
	)
	if appendErr != nil {
		return nil, contract.NewError(
			protos.ErrorCode_INTERNAL_ERROR,
			"failed to append run ID to experiment artifact location",
		)
	}

	runModel.ArtifactURI = artifactLocation

	errRunName := ensureRunName(runModel)
	if errRunName != nil {
		return nil, errRunName
	}

	if err := s.db.Create(&runModel).Error; err != nil {
		return nil, contract.NewErrorWith(
			protos.ErrorCode_INTERNAL_ERROR,
			fmt.Sprintf(
				"failed to create run for experiment_id %q",
				experiment.ExperimentID,
			),
			err,
		)
	}

	return runModel.ToEntity(), nil
}

func (s TrackingSQLStore) UpdateRun(
	ctx context.Context,
	runID string,
	runStatus string,
	endTime *int64,
	runName string,
) *contract.Error {
	runTag, err := s.GetRunTag(ctx, runID, utils.TagRunName)
	if err != nil {
		return err
	}

	tags := make([]models.Tag, 0, 1)
	if runTag == nil {
		tags = append(tags, models.Tag{
			RunID: runID,
			Key:   utils.TagRunName,
			Value: runName,
		})
	}

	var endTimeValue sql.NullInt64
	if endTime == nil {
		endTimeValue = sql.NullInt64{}
	} else {
		endTimeValue = sql.NullInt64{Int64: *endTime, Valid: true}
	}

	if err := s.db.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		if err := transaction.Model(&models.Run{}).
			Where("run_uuid = ?", runID).
			Updates(&models.Run{
				Name:    runName,
				Status:  models.RunStatus(runStatus),
				EndTime: endTimeValue,
			}).Error; err != nil {
			return err
		}

		if len(tags) > 0 {
			if err := transaction.Clauses(clause.OnConflict{
				UpdateAll: true,
			}).CreateInBatches(tags, tagsBatchSize).Error; err != nil {
				return fmt.Errorf("failed to create tags for run %q: %w", runID, err)
			}
		}

		return nil
	}); err != nil {
		return contract.NewErrorWith(protos.ErrorCode_INTERNAL_ERROR, "failed to update run", err)
	}

	return nil
}

func (s TrackingSQLStore) DeleteRun(ctx context.Context, runID string) *contract.Error {
	run, err := s.GetRun(ctx, runID)
	if err != nil {
		return err
	}

	if err := s.db.WithContext(ctx).Model(&models.Run{}).
		Where("run_uuid = ?", run.Info.RunID).
		Updates(&models.Run{
			DeletedTime:    sql.NullInt64{Valid: true, Int64: time.Now().UnixMilli()},
			LifecycleStage: models.LifecycleStageDeleted,
		}).Error; err != nil {
		return contract.NewErrorWith(protos.ErrorCode_INTERNAL_ERROR, "failed to delete run", err)
	}

	return nil
}

func (s TrackingSQLStore) RestoreRun(ctx context.Context, runID string) *contract.Error {
	run, err := s.GetRun(ctx, runID)
	if err != nil {
		return err
	}

	if err := s.db.WithContext(ctx).Model(&models.Run{}).
		Where("run_uuid = ?", run.Info.RunID).
		// Force GORM to update fields with zero values by selecting them.
		Select("DeletedTime", "LifecycleStage").
		Updates(&models.Run{
			DeletedTime:    sql.NullInt64{},
			LifecycleStage: models.LifecycleStageActive,
		}).Error; err != nil {
		return contract.NewErrorWith(protos.ErrorCode_INTERNAL_ERROR, "failed to restore run", err)
	}

	return nil
}

func (s TrackingSQLStore) LogBatch(
	ctx context.Context, runID string, metrics []*entities.Metric, params []*entities.Param, tags []*entities.RunTag,
) *contract.Error {
	err := s.db.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		contractError := checkRunIsActive(transaction, runID)
		if contractError != nil {
			return contractError
		}

		err := s.setTagsWithTransaction(transaction, runID, tags)
		if err != nil {
			return fmt.Errorf("error setting tags for run_id %q: %w", runID, err)
		}

		contractError = s.logParamsWithTransaction(transaction, runID, params)
		if contractError != nil {
			return contractError
		}

		contractError = s.logMetricsWithTransaction(transaction, runID, metrics)
		if contractError != nil {
			return contractError
		}

		return nil
	})
	if err != nil {
		var contractError *contract.Error
		if errors.As(err, &contractError) {
			return contractError
		}

		return contract.NewErrorWith(
			protos.ErrorCode_INTERNAL_ERROR,
			fmt.Sprintf("log batch transaction failed for %q", runID),
			err,
		)
	}

	return nil
}
