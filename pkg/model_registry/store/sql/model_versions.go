package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/entities"
	"github.com/mlflow/mlflow-go/pkg/model_registry/store/sql/models"
	"github.com/mlflow/mlflow-go/pkg/protos"
)

// Validate whether there is a registered model with the given name.
func assertModelExists(db *gorm.DB, name string) *contract.Error {
	if err := db.Select("name").Where("name = ?", name).First(&models.RegisteredModel{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return contract.NewError(
				protos.ErrorCode_RESOURCE_DOES_NOT_EXIST,
				fmt.Sprintf("registered model with name=%q not found", name),
			)
		}

		return contract.NewErrorWith(
			protos.ErrorCode_INTERNAL_ERROR,
			fmt.Sprintf("failed to query registered model with name=%q", name),
			err,
		)
	}

	return nil
}

func (m *ModelRegistrySQLStore) GetLatestVersions(
	ctx context.Context, name string, stages []string,
) ([]*protos.ModelVersion, *contract.Error) {
	if err := assertModelExists(m.db.WithContext(ctx), name); err != nil {
		return nil, err
	}

	var modelVersions []*models.ModelVersion

	subQuery := m.db.
		WithContext(ctx).
		Model(&models.ModelVersion{}).
		Select("name, MAX(version) AS max_version").
		Where("name = ?", name).
		Where("current_stage <> ?", models.StageDeletedInternal).
		Group("name, current_stage")

	if len(stages) > 0 {
		for idx, stage := range stages {
			stages[idx] = strings.ToLower(stage)
			if canonicalStage, ok := models.CanonicalMapping[stages[idx]]; ok {
				stages[idx] = canonicalStage

				continue
			}

			return nil, contract.NewError(
				protos.ErrorCode_BAD_REQUEST,
				fmt.Sprintf(
					"Invalid Model Version stage: %s. Value must be one of %s.",
					stage,
					models.AllModelVersionStages(),
				),
			)
		}

		subQuery = subQuery.Where("current_stage IN (?)", stages)
	}

	err := m.db.
		WithContext(ctx).
		Model(&models.ModelVersion{}).
		Joins("JOIN (?) AS sub ON model_versions.name = sub.name AND model_versions.version = sub.max_version", subQuery).
		Find(&modelVersions).Error
	if err != nil {
		return nil, contract.NewErrorWith(
			protos.ErrorCode_INTERNAL_ERROR,
			fmt.Sprintf("failed to query latest model version for %q", name),
			err,
		)
	}

	results := make([]*protos.ModelVersion, 0, len(modelVersions))
	for _, modelVersion := range modelVersions {
		results = append(results, modelVersion.ToProto())
	}

	return results, nil
}

func (m *ModelRegistrySQLStore) GetRegisteredModel(
	ctx context.Context, name string,
) (*entities.RegisteredModel, *contract.Error) {
	var registeredModel models.RegisteredModel
	if err := m.db.WithContext(
		ctx,
	).Where(
		"name = ?", name,
	).Preload(
		"Tags",
	).Preload(
		"Aliases",
	).Preload(
		"Versions",
	).First(
		&registeredModel,
	).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, contract.NewError(
				protos.ErrorCode_RESOURCE_DOES_NOT_EXIST,
				fmt.Sprintf("Registered Model with name=%s not found", name),
			)
		}

		//nolint:perfsprint
		return nil, contract.NewErrorWith(
			protos.ErrorCode_INTERNAL_ERROR,
			fmt.Sprintf("failed to get Registered Model by name %s", name),
			err,
		)
	}

	return registeredModel.ToEntity(), nil
}

func (m *ModelRegistrySQLStore) UpdateRegisteredModel(
	ctx context.Context, name, description string,
) (*entities.RegisteredModel, *contract.Error) {
	registeredModel, err := m.GetRegisteredModel(ctx, name)
	if err != nil {
		return nil, err
	}

	if err := m.db.WithContext(ctx).Model(
		&models.RegisteredModel{},
	).Where(
		"name = ?", registeredModel.Name,
	).Updates(&models.RegisteredModel{
		Name:            name,
		Description:     sql.NullString{String: description, Valid: true},
		LastUpdatedTime: time.Now().UnixMilli(),
	}).Error; err != nil {
		return nil, contract.NewErrorWith(protos.ErrorCode_INTERNAL_ERROR, "failed to update registered model", err)
	}

	return registeredModel, nil
}

func (m *ModelRegistrySQLStore) RenameRegisteredModel(
	ctx context.Context, name, newName string,
) (*entities.RegisteredModel, *contract.Error) {
	registeredModel, err := m.GetRegisteredModel(ctx, name)
	if err != nil {
		return nil, err
	}

	if err := m.db.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		if err := transaction.Model(
			&models.ModelVersion{},
		).Where(
			"name = ?", registeredModel.Name,
		).Updates(&models.ModelVersion{
			Name:            newName,
			LastUpdatedTime: time.Now().UnixMilli(),
		}).Error; err != nil {
			return err
		}

		if err := transaction.Model(
			&models.RegisteredModel{},
		).Where(
			"name = ?", registeredModel.Name,
		).Updates(&models.RegisteredModel{
			Name:            newName,
			LastUpdatedTime: time.Now().UnixMilli(),
		}).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, contract.NewErrorWith(
				protos.ErrorCode_RESOURCE_ALREADY_EXISTS,
				fmt.Sprintf("Registered Model (name=%s) already exists", newName),
				err,
			)
		}

		return nil, contract.NewErrorWith(protos.ErrorCode_INTERNAL_ERROR, "failed to rename registered model", err)
	}

	registeredModel, err = m.GetRegisteredModel(ctx, newName)
	if err != nil {
		return nil, err
	}

	return registeredModel, nil
}

func (m *ModelRegistrySQLStore) DeleteRegisteredModel(ctx context.Context, name string) *contract.Error {
	registeredModel, err := m.GetRegisteredModel(ctx, name)
	if err != nil {
		return err
	}

	if err := m.db.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		if err := transaction.Where(
			"name = ?", registeredModel.Name,
		).Delete(
			models.ModelVersionTag{},
		).Error; err != nil {
			return err
		}

		if err := transaction.Where(
			"name = ?", registeredModel.Name,
		).Delete(
			models.ModelVersion{},
		).Error; err != nil {
			return err
		}

		if err := transaction.Where(
			"name = ?", registeredModel.Name,
		).Delete(
			models.RegisteredModelTag{},
		).Error; err != nil {
			return err
		}

		if err := transaction.Where(
			"name = ?", registeredModel.Name,
		).Delete(
			models.RegisteredModelAlias{},
		).Error; err != nil {
			return err
		}

		if err := transaction.Where(
			"name = ?", registeredModel.Name,
		).Delete(
			models.RegisteredModel{},
		).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return contract.NewError(
			protos.ErrorCode_INTERNAL_ERROR, fmt.Sprintf("error deleting registered model: %v", err),
		)
	}

	return nil
}

func (m *ModelRegistrySQLStore) GetModelVersion(
	ctx context.Context, name, version string,
) (*entities.ModelVersion, *contract.Error) {
	var modelVersion models.ModelVersion
	if err := m.db.WithContext(
		ctx,
	).Where(
		"name = ?", name,
	).Where(
		"version = ?", version,
	).Where(
		"current_stage != ?", models.StageDeletedInternal,
	).First(
		&modelVersion,
	).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, contract.NewError(
				protos.ErrorCode_RESOURCE_DOES_NOT_EXIST,
				fmt.Sprintf("Model Version (name=%s, version=%s) not found", name, version),
			)
		}

		return nil, contract.NewErrorWith(
			protos.ErrorCode_INTERNAL_ERROR,
			fmt.Sprintf("failed to get Model Version by name %s and version %s", name, version),
			err,
		)
	}

	return modelVersion.ToEntity(), nil
}

func (m *ModelRegistrySQLStore) DeleteModelVersion(ctx context.Context, name, version string) *contract.Error {
	registeredModel, err := m.GetRegisteredModel(ctx, name)
	if err != nil {
		return err
	}

	modelVersion, err := m.GetModelVersion(ctx, name, version)
	if err != nil {
		return err
	}

	if err := m.db.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		if err := transaction.Model(
			&models.RegisteredModel{},
		).Where(
			"name = ?", registeredModel.Name,
		).Updates(&models.RegisteredModel{
			LastUpdatedTime: time.Now().UnixMilli(),
		}).Error; err != nil {
			return err
		}

		if err := transaction.Where(
			"name = ?", registeredModel.Name,
		).Where(
			"version = ?", version,
		).Delete(
			&models.RegisteredModelAlias{},
		).Error; err != nil {
			return err
		}

		if err := transaction.Model(
			&models.ModelVersion{},
		).Where(
			"name = ?", modelVersion.Name,
		).Where(
			"version = ?", modelVersion.Version,
		).Updates(&models.ModelVersion{
			RunID:           "REDACTED-RUN-ID",
			UserID:          sql.NullString{Valid: true},
			Source:          "REDACTED-SOURCE-PATH",
			RunLink:         "REDACTED-RUN-LINK",
			CurrentStage:    models.StageDeletedInternal,
			Description:     sql.NullString{Valid: true},
			StatusMessage:   sql.NullString{Valid: true},
			LastUpdatedTime: time.Now().UnixMilli(),
		}).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return contract.NewErrorWith(protos.ErrorCode_INTERNAL_ERROR, "error deleting model version", err)
	}

	return nil
}

func (m *ModelRegistrySQLStore) UpdateModelVersion(
	ctx context.Context, name, version, description string,
) (*entities.ModelVersion, *contract.Error) {
	modelVersion, err := m.GetModelVersion(ctx, name, version)
	if err != nil {
		return nil, err
	}

	if err := m.db.WithContext(ctx).Model(
		&models.ModelVersion{},
	).Where(
		"name = ?", modelVersion.Name,
	).Where(
		"version = ?", modelVersion.Version,
	).Updates(&models.ModelVersion{
		Description:     sql.NullString{String: description, Valid: description != ""},
		LastUpdatedTime: time.Now().UnixMilli(),
	}).Error; err != nil {
		return nil, contract.NewErrorWith(protos.ErrorCode_INTERNAL_ERROR, "error updating model version", err)
	}

	return modelVersion, nil
}
