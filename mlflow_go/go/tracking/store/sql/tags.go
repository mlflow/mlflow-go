package sql

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/mlflow/mlflow-go/mlflow_go/go/protos"
	"github.com/mlflow/mlflow-go/mlflow_go/go/tracking/store/sql/models"
	"github.com/mlflow/mlflow-go/mlflow_go/go/utils"
)

const tagsBatchSize = 100

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
		runTags = append(runTags, models.NewTagFromProto(&runID, tag))
	}

	if err := transaction.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(runTags, tagsBatchSize).Error; err != nil {
		return fmt.Errorf("failed to create tags for run %q: %w", runID, err)
	}

	return nil
}
