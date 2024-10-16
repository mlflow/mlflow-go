package sql

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/entities"
	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/tracking/store/sql/models"
	"github.com/mlflow/mlflow-go/pkg/utils"
)

const tagsBatchSize = 100

func (s TrackingSQLStore) GetRunTag(
	ctx context.Context, runID, tagKey string,
) (*entities.RunTag, *contract.Error) {
	var runTag models.Tag
	if err := s.db.WithContext(
		ctx,
	).Where(
		"run_uuid = ?", runID,
	).Where(
		"key = ?", tagKey,
	).First(&runTag).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, contract.NewErrorWith(
			protos.ErrorCode_INTERNAL_ERROR,
			fmt.Sprintf("failed to get run tag for run id %q", runID),
			err,
		)
	}

	return runTag.ToEntity(), nil
}

func (s TrackingSQLStore) setTagsWithTransaction(
	transaction *gorm.DB, runID string, tags []*entities.RunTag,
) error {
	runColumns := make(map[string]interface{})

	for _, tag := range tags {
		switch tag.Key {
		case utils.TagUser:
			runColumns["user_id"] = tag.Value
		case utils.TagRunName:
			runColumns["name"] = tag.Value
		}
	}

	if len(runColumns) != 0 {
		err := transaction.
			Model(&models.Run{}).
			Where("run_uuid = ?", runID).
			UpdateColumns(runColumns).Error
		if err != nil {
			return fmt.Errorf("failed to update run columns: %w", err)
		}
	}

	runTags := make([]models.Tag, 0, len(tags))

	for _, tag := range tags {
		runTags = append(runTags, models.NewTagFromEntity(runID, tag))
	}

	if err := transaction.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(runTags, tagsBatchSize).Error; err != nil {
		return fmt.Errorf("failed to create tags for run %q: %w", runID, err)
	}

	return nil
}

const badDataMessage = "Bad data in database - tags for a specific run must have\n" +
	"a single unique value.\n" +
	"See https://mlflow.org/docs/latest/tracking.html#adding-tags-to-runs"

func (s TrackingSQLStore) DeleteTag(
	ctx context.Context, runID, key string,
) *contract.Error {
	err := s.db.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		contractError := checkRunIsActive(transaction, runID)
		if contractError != nil {
			return contractError
		}

		var tags []models.Tag

		transaction.Model(models.Tag{}).Where("run_uuid = ?", runID).Where("key = ?", key).Find(&tags)

		if transaction.Error != nil {
			return contract.NewErrorWith(
				protos.ErrorCode_INTERNAL_ERROR,
				fmt.Sprintf("Failed to query tags for run_id %q and key %q", runID, key),
				transaction.Error,
			)
		}

		switch len(tags) {
		case 0:
			return contract.NewError(
				protos.ErrorCode_RESOURCE_DOES_NOT_EXIST,
				fmt.Sprintf("No tag with name: %s in run with id %s", key, runID),
			)
		case 1:
			transaction.Delete(tags[0])

			if transaction.Error != nil {
				return contract.NewErrorWith(
					protos.ErrorCode_INTERNAL_ERROR,
					fmt.Sprintf("Failed to query tags for run_id %q and key %q", runID, key),
					transaction.Error,
				)
			}

			return nil
		default:
			return contract.NewError(protos.ErrorCode_INVALID_STATE, badDataMessage)
		}
	})
	if err != nil {
		var contractError *contract.Error
		if errors.As(err, &contractError) {
			return contractError
		}

		return contract.NewErrorWith(
			protos.ErrorCode_INTERNAL_ERROR,
			fmt.Sprintf("delete tag transaction failed for %q", runID),
			err,
		)
	}

	return nil
}

func (s TrackingSQLStore) SetTag(
	ctx context.Context, runID, key, value string,
) *contract.Error {
	var run models.Run

	err := s.db.Where("run_uuid = ?", runID).First(&run).Error
	if err != nil {
		return contract.NewErrorWith(
			protos.ErrorCode_INTERNAL_ERROR,
			fmt.Sprintf("Failed to query run for run_id %q", runID),
			err,
		)
	}

	err = s.db.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		contractError := checkRunIsActive(transaction, runID)
		if contractError != nil {
			return contractError
		}

		return s.handleTagUpsert(transaction, runID, key, value)
	})
	if err != nil {
		var contractError *contract.Error
		if errors.As(err, &contractError) {
			return contractError
		}

		return contract.NewErrorWith(
			protos.ErrorCode_INTERNAL_ERROR,
			fmt.Sprintf("Set tag transaction failed for run_id %q", runID),
			err,
		)
	}

	if key == utils.TagRunName {
		return s.handleRunNameUpdate(ctx, run, runID, value)
	}

	return nil
}

// Handle tag creation and update.
func (s TrackingSQLStore) handleTagUpsert(
	transaction *gorm.DB, runID, key, value string,
) error {
	var tag models.Tag

	result := transaction.Where("run_uuid = ? AND key = ?", runID, key).First(&tag)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return contract.NewErrorWith(
			protos.ErrorCode_INTERNAL_ERROR,
			fmt.Sprintf("Failed to query tag for run_id %q and key %q", runID, key),
			result.Error,
		)
	}

	if result.RowsAffected == 1 {
		tag.Value = value
		if err := transaction.Save(&tag).Error; err != nil {
			return contract.NewErrorWith(
				protos.ErrorCode_INTERNAL_ERROR,
				fmt.Sprintf("Failed to update tag for run_id %q and key %q", runID, key),
				err,
			)
		}
	} else {
		newTag := models.Tag{
			RunID: runID,
			Key:   key,
			Value: value,
		}
		if err := transaction.Create(&newTag).Error; err != nil {
			return contract.NewErrorWith(
				protos.ErrorCode_INTERNAL_ERROR,
				fmt.Sprintf("Failed to create tag for run_id %q and key %q", runID, key),
				err,
			)
		}
	}

	return nil
}

// Handles updating the run name when setting tag.
func (s TrackingSQLStore) handleRunNameUpdate(
	ctx context.Context, run models.Run, runID, value string,
) *contract.Error {
	runStatus := run.Status.String()

	var endTimePtr *int64
	if run.EndTime.Valid {
		endTimePtr = &run.EndTime.Int64
	}

	if err := s.UpdateRun(ctx, runID, runStatus, endTimePtr, value); err != nil {
		return contract.NewErrorWith(
			protos.ErrorCode_INTERNAL_ERROR,
			fmt.Sprintf("Failed to update run info for run_id %q", runID),
			err,
		)
	}

	return nil
}
