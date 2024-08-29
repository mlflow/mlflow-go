package sql

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/mlflow/mlflow-go/pkg/contract"
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
