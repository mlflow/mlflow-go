package sql

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/entities"
	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/tracking/store/sql/models"
)

type digestKey struct {
	Name   string
	Digest string
}

type existingInput struct {
	SourceID      string
	DestinationID string
}

const batchSize = 100

func newGUID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

func getExperimentIDFromRun(transaction *gorm.DB, runID string) (int32, error) {
	var experimentID int32

	err := transaction.
		Model(&models.Run{}).
		Select("ExperimentID").
		Where("run_uuid = ?", runID).
		Pluck("ExperimentID", &experimentID).
		Error
	if err != nil {
		return -1, err
	}

	return experimentID, nil
}

func dedupDatasetInputs(datasets []*entities.DatasetInput) ([]string, []string, []*entities.DatasetInput) {
	nameDigestKeys := make(map[digestKey]*entities.DatasetInput)

	for _, dataset := range datasets {
		key := digestKey{Name: dataset.Dataset.Name, Digest: dataset.Dataset.Digest}
		if _, ok := nameDigestKeys[key]; !ok {
			nameDigestKeys[key] = dataset
		}
	}

	datasetNamesToCheck := make([]string, 0, len(nameDigestKeys))
	datasetDigestsToCheck := make([]string, 0, len(nameDigestKeys))
	datasetInputs := make([]*entities.DatasetInput, 0, len(nameDigestKeys))

	for _, dataset := range nameDigestKeys {
		datasetInputs = append(datasetInputs, dataset)
		datasetNamesToCheck = append(datasetNamesToCheck, dataset.Dataset.Name)
		datasetDigestsToCheck = append(datasetDigestsToCheck, dataset.Dataset.Digest)
	}

	return datasetNamesToCheck, datasetDigestsToCheck, datasetInputs
}

func findExistingDatasets(
	transaction *gorm.DB,
	datasetNamesToCheck []string,
	datasetDigestsToCheck []string,
) ([]*models.Dataset, *contract.Error) {
	existingDatasets := make([]*models.Dataset, 0)

	err := transaction.
		Where("name IN ?", datasetNamesToCheck).
		Where("digest IN ?", datasetDigestsToCheck).
		Find(&existingDatasets).
		Error
	if err != nil {
		return nil,
			contract.NewErrorWith(
				protos.ErrorCode_INTERNAL_ERROR,
				"failed to find existing datasets",
				err)
	}

	return existingDatasets, nil
}

func findExistingInputs(
	transaction *gorm.DB, datasetUuidsValues []string, runID string,
) (map[existingInput]any, *contract.Error) {
	existingInputs := make([]existingInput, 0)

	err := transaction.
		Model(&models.Input{}).
		Where("source_type = ?", models.SourceTypeDataset).
		Where("source_id IN ?", datasetUuidsValues).
		Where("destination_type = ?", models.DestinationTypeRun).
		Where("destination_id = ?", runID).
		Find(&existingInputs).Error
	if err != nil {
		return nil, contract.NewErrorWith(
			protos.ErrorCode_INTERNAL_ERROR,
			fmt.Sprintf("failed to find existing inputs for runID %q", runID),
			err)
	}

	existingInputsMap := make(map[existingInput]any, len(existingInputs))
	for _, existingInput := range existingInputs {
		existingInputsMap[existingInput] = struct{}{}
	}

	return existingInputsMap, nil
}

func collectDatasetUuids(existingDatasets []*models.Dataset) (map[digestKey]string, []string) {
	datasetUuids := make(map[digestKey]string, len(existingDatasets))
	datasetUuidsValues := make([]string, 0, len(existingDatasets))

	for _, dataset := range existingDatasets {
		datasetUuids[digestKey{Name: dataset.Name, Digest: dataset.Digest}] = dataset.ID
		datasetUuidsValues = append(datasetUuidsValues, dataset.ID)
	}

	return datasetUuids, datasetUuidsValues
}

func mkDataset(
	newDatasetUUID string, experimentID int32, dataset *entities.DatasetInput,
) *models.Dataset {
	return &models.Dataset{
		ID:           newDatasetUUID,
		ExperimentID: experimentID,
		Name:         dataset.Dataset.Name,
		Digest:       dataset.Dataset.Digest,
		SourceType:   dataset.Dataset.SourceType,
		Source:       dataset.Dataset.Source,
		Schema:       dataset.Dataset.Schema,
		Profile:      dataset.Dataset.Profile,
	}
}

//nolint:funlen,cyclop
func (s TrackingSQLStore) LogInputs(
	ctx context.Context, runID string, datasets []*entities.DatasetInput,
) *contract.Error {
	err := s.db.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		contractError := checkRunIsActive(transaction, runID)
		if contractError != nil {
			return contractError
		}

		// Do we want to combine this with the checkRunIsActive function?
		experimentID, err := getExperimentIDFromRun(transaction, runID)
		if err != nil {
			return contract.NewErrorWith(
				protos.ErrorCode_INTERNAL_ERROR,
				fmt.Sprintf("could not find experiment_id for %q", runID),
				err,
			)
		}

		// dedup dataset_inputs list if two dataset inputs have the same name and digest
		// keeping the first occurrence
		datasetNamesToCheck, datasetDigestsToCheck, datasetInputs := dedupDatasetInputs(datasets)

		existingDatasets, constractError := findExistingDatasets(transaction, datasetNamesToCheck, datasetDigestsToCheck)
		if constractError != nil {
			return constractError
		}

		datasetUuids, datasetUuidsValues := collectDatasetUuids(existingDatasets)

		existingInputsMap, contractError := findExistingInputs(transaction, datasetUuidsValues, runID)
		if contractError != nil {
			return contractError
		}

		datasetToInsert := make([]*models.Dataset, 0)
		inputToInsert := make([]*models.Input, 0)
		inputTagsToInsert := make([]*models.InputTag, 0)

		for _, dataset := range datasetInputs {
			key := digestKey{Name: dataset.Dataset.Name, Digest: dataset.Dataset.Digest}
			if _, ok := datasetUuids[key]; !ok {
				newDatasetUUID := uuid.New().String()
				datasetUuids[key] = newDatasetUUID
				datasetToInsert = append(datasetToInsert, mkDataset(newDatasetUUID, experimentID, dataset))
			}

			inputKey := existingInput{SourceID: datasetUuids[key], DestinationID: runID}
			if _, ok := existingInputsMap[inputKey]; !ok {
				existingInputsMap[inputKey] = struct{}{}
				newInputUUID := newGUID()
				inputToInsert = append(inputToInsert, models.NewInputFromEntity(newInputUUID, inputKey.SourceID, runID))

				for _, tag := range dataset.Tags {
					inputTagsToInsert = append(inputTagsToInsert, models.NewInputTagFromEntity(newInputUUID, tag))
				}
			}
		}

		err = transaction.CreateInBatches(&datasetToInsert, batchSize).Error
		if err != nil {
			return err
		}

		err = transaction.CreateInBatches(&inputToInsert, batchSize).Error
		if err != nil {
			return err
		}

		err = transaction.CreateInBatches(&inputTagsToInsert, batchSize).Error
		if err != nil {
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
			fmt.Sprintf("log inputs transaction failed for %q", runID),
			err,
		)
	}

	return nil
}
