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
)

const paramsBatchSize = 100

func verifyBatchParamsInserts(
	transaction *gorm.DB, runID string, deduplicatedParamsMap map[string]*string,
) *contract.Error {
	keys := make([]string, 0, len(deduplicatedParamsMap))
	for key := range deduplicatedParamsMap {
		keys = append(keys, key)
	}

	var existingParams []models.Param

	err := transaction.
		Model(&models.Param{}).
		Select("key, value").
		Where("run_uuid = ?", runID).
		Where("key IN ?", keys).
		Find(&existingParams).Error
	if err != nil {
		return contract.NewErrorWith(
			protos.ErrorCode_INTERNAL_ERROR,
			fmt.Sprintf(
				"failed to get existing params to check if duplicates for run_id %q",
				runID,
			),
			err)
	}

	for _, existingParam := range existingParams {
		if currentValue, ok := deduplicatedParamsMap[existingParam.Key]; ok &&
			currentValue != nil && *currentValue != existingParam.Value.String {
			return contract.NewError(
				protos.ErrorCode_INVALID_PARAMETER_VALUE,
				fmt.Sprintf(
					"Changing param values is not allowed. "+
						"Params with key=%q was already logged "+
						"with value=%q for run ID=%q. "+
						"Attempted logging new value %q",
					existingParam.Key,
					existingParam.Value.String,
					runID,
					*currentValue,
				),
			)
		}
	}

	return nil
}

func (s TrackingSQLStore) logParamsWithTransaction(
	transaction *gorm.DB, runID string, params []*entities.Param,
) *contract.Error {
	deduplicatedParamsMap := make(map[string]*string, len(params))
	deduplicatedParams := make([]models.Param, 0, len(deduplicatedParamsMap))

	for _, param := range params {
		oldValue, paramIsPresent := deduplicatedParamsMap[param.Key]
		if paramIsPresent && param.Value != nil && *param.Value != *oldValue {
			return contract.NewError(
				protos.ErrorCode_INVALID_PARAMETER_VALUE,
				fmt.Sprintf(
					"Changing param values is not allowed. "+
						"Params with key=%q was already logged "+
						"with value=%q for run ID=%q. "+
						"Attempted logging new value %q",
					param.Key,
					*oldValue,
					runID,
					*param.Value,
				),
			)
		}

		if !paramIsPresent {
			deduplicatedParamsMap[param.Key] = param.Value
			deduplicatedParams = append(deduplicatedParams, models.NewParamFromEntity(runID, param))
		}
	}

	// Try and create all params.
	// Params are unique by (run_uuid, key) so any potentially conflicts will not be inserted.
	err := transaction.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "run_uuid"}, {Name: "key"}},
			DoNothing: true,
		}).
		CreateInBatches(deduplicatedParams, paramsBatchSize).Error
	if err != nil {
		// The strange thing that error has `protos.ErrorCode_BAD_REQUEST` code.
		// This is exactly what MLFlow tests expect to see.
		return contract.NewErrorWith(
			protos.ErrorCode_BAD_REQUEST,
			fmt.Sprintf("error creating params in batch for run_uuid %q: %v", runID, err),
			err,
		)
	}

	// if there were ignored conflicts, we assert that the values are the same.
	if transaction.RowsAffected != int64(len(params)) {
		contractError := verifyBatchParamsInserts(transaction, runID, deduplicatedParamsMap)
		if contractError != nil {
			return contractError
		}
	}

	return nil
}

func (s TrackingSQLStore) LogParam(
	ctx context.Context, runID string, param *entities.Param,
) *contract.Error {
	err := s.db.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		if err := checkRunIsActive(transaction, runID); err != nil {
			return err
		}

		if err := s.logParamsWithTransaction(transaction, runID, []*entities.Param{param}); err != nil {
			return err
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
			fmt.Sprintf("log param failed for %q", runID),
			err,
		)
	}

	return nil
}
