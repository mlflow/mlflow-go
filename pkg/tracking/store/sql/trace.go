package sql

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/entities"
	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/tracking/store/sql/models"
)

func (s TrackingSQLStore) GetTraceTag(
	ctx context.Context, requestID, key string,
) (*entities.TraceTag, *contract.Error) {
	var tag models.TraceTag
	if err := s.db.WithContext(ctx).Where(
		"request_id = ?", requestID,
	).Where(
		"key = ?", key,
	).First(
		&tag,
	).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, contract.NewError(
				protos.ErrorCode_RESOURCE_DOES_NOT_EXIST,
				fmt.Sprintf(
					"No trace tag with key '%s' for trace with request_id '%s'",
					key,
					requestID,
				),
			)
		}

		return nil, contract.NewError(protos.ErrorCode_INTERNAL_ERROR, fmt.Sprintf("error getting trace tag: %v", err))
	}

	return tag.ToEntity(), nil
}

func (s TrackingSQLStore) DeleteTraceTag(
	ctx context.Context, tag *entities.TraceTag,
) *contract.Error {
	if err := s.db.WithContext(ctx).Where(
		"request_id = ?", tag.RequestID,
	).Where(
		"key = ?", tag.Key,
	).Delete(
		entities.TraceTag{},
	).Error; err != nil {
		return contract.NewError(protos.ErrorCode_INTERNAL_ERROR, fmt.Sprintf("error deleting trace tag: %v", err))
	}

	return nil
}
