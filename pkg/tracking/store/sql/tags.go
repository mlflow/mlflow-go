package sql

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/tracking/store/sql/models"
	"github.com/mlflow/mlflow-go/pkg/utils"
)

const tagsBatchSize = 100

func (s TrackingSQLStore) GetRunTag(
	ctx context.Context, runID, tagKey string,
) (*protos.RunTag, *contract.Error) {
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

	return runTag.ToProto(), nil
}

func (s TrackingSQLStore) setTagsWithTransaction(
	transaction *gorm.DB, runID string, tags []*protos.RunTag,
) error {
	runColumns := make(map[string]interface{})

	for _, tag := range tags {
		switch tag.GetKey() {
		case utils.TagUser:
			runColumns["user_id"] = tag.GetValue()
		case utils.TagRunName:
			runColumns["name"] = tag.GetValue()
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
		runTags = append(runTags, models.NewTagFromProto(runID, tag))
	}

	if err := transaction.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(runTags, tagsBatchSize).Error; err != nil {
		return fmt.Errorf("failed to create tags for run %q: %w", runID, err)
	}

	return nil
}
