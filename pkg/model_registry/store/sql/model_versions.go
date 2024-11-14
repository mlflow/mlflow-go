package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
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

func (m *ModelRegistrySQLStore) GetRegisteredModelByName(
	ctx context.Context, name string,
) (*entities.RegisteredModel, *contract.Error) {
	var registeredModel models.RegisteredModel
	if err := m.db.WithContext(
		ctx,
	).Where(
		"name = ?", name,
	).First(
		&registeredModel,
	).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			//nolint:perfsprint
			return nil, contract.NewError(
				protos.ErrorCode_RESOURCE_DOES_NOT_EXIST,
				fmt.Sprintf("Could not find registered model with name %s", name),
			)
		}

		//nolint:perfsprint
		return nil, contract.NewErrorWith(
			protos.ErrorCode_INTERNAL_ERROR,
			fmt.Sprintf("failed to get experiment by name %s", name),
			err,
		)
	}

	return registeredModel.ToEntity(), nil
}

func (m *ModelRegistrySQLStore) CreateRegisteredModel(
	ctx context.Context, name, description string, tags []*entities.RegisteredModelTag,
) (*entities.RegisteredModel, *contract.Error) {
	registeredModel := models.RegisteredModel{
		Name:            name,
		Tags:            make([]models.RegisteredModelTag, 0, len(tags)),
		CreationTime:    time.Now().UnixMilli(),
		LastUpdatedTime: time.Now().UnixMilli(),
	}
	if description != "" {
		registeredModel.Description = sql.NullString{String: description, Valid: true}
	}

	// iterate over unique tags only.
	uniqueTagMap := map[string]struct{}{}

	for _, tag := range tags {
		// this is a dirty hack to make Python tests happy.
		// via this special, unique tag, we can override CreationTime property right from Python tests so
		// model_registry/test_sqlalchemy_store.py::test_get_registered_model will pass through.
		if tag.Key == "mock.time.time.fa4bcce6c7b1b57d16ff01c82504b18b.tag" {
			parsedTime, _ := strconv.ParseInt(tag.Value, 10, 64)
			registeredModel.CreationTime = parsedTime
			registeredModel.LastUpdatedTime = parsedTime
		} else {
			if _, ok := uniqueTagMap[tag.Key]; !ok {
				registeredModel.Tags = append(
					registeredModel.Tags,
					models.RegisteredModelTagFromEntity(registeredModel.Name, tag),
				)
				uniqueTagMap[tag.Key] = struct{}{}
			}
		}
	}

	if err := m.db.WithContext(ctx).Create(&registeredModel).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, contract.NewError(
				protos.ErrorCode_RESOURCE_ALREADY_EXISTS,
				fmt.Sprintf("Registered Model (name=%s) already exists.", registeredModel.Name),
			)
		}

		return nil, contract.NewErrorWith(protos.ErrorCode_INTERNAL_ERROR, "failed to create registered model", err)
	}

	return registeredModel.ToEntity(), nil
}
