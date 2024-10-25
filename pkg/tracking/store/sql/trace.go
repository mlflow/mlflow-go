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

func (s TrackingSQLStore) SetTrace(
	ctx context.Context,
	experimentID string,
	timestampMS int64,
	metadata []*entities.TraceRequestMetadata,
	tags []*entities.TraceTag,
) (*entities.TraceInfo, error) {
	traceInfo := &models.TraceInfo{
		RequestID:            utils.NewUUID(),
		ExperimentID:         experimentID,
		TimestampMS:          timestampMS,
		Status:               models.TraceInfoStatusInProgress,
		Tags:                 make([]models.TraceTag, 0, len(tags)),
		TraceRequestMetadata: make([]models.TraceRequestMetadata, 0, len(metadata)),
	}

	experiment, err := s.GetExperiment(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	for _, tag := range tags {
		traceInfo.Tags = append(traceInfo.Tags, models.NewTraceTagFromEntity(traceInfo.RequestID, tag))
	}

	traceArtifactLocationTag, artifactLocationTagErr := GetTraceArtifactLocationTag(experiment, traceInfo.RequestID)
	if artifactLocationTagErr != nil {
		return nil, contract.NewErrorWith(
			protos.ErrorCode_INTERNAL_ERROR,
			fmt.Sprintf("failed to create trace for experiment_id %q", experimentID),
			err,
		)
	}

	traceInfo.Tags = append(traceInfo.Tags, traceArtifactLocationTag)

	for _, m := range metadata {
		traceInfo.TraceRequestMetadata = append(
			traceInfo.TraceRequestMetadata, models.NewTraceRequestMetadataFromEntity(traceInfo.RequestID, m),
		)
	}

	if err := s.db.WithContext(ctx).Create(&traceInfo).Error; err != nil {
		return nil, contract.NewErrorWith(
			protos.ErrorCode_INTERNAL_ERROR,
			fmt.Sprintf("failed to create trace for experiment_id %q", experimentID),
			err,
		)
	}

	return traceInfo.ToEntity(), nil
}

func (s TrackingSQLStore) SetTraceTag(
	ctx context.Context, requestID, key, value string,
) error {
	if err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}, {Name: "request_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(models.TraceTag{
		Key:       key,
		Value:     value,
		RequestID: requestID,
	}).Error; err != nil {
		return err
	}

	return nil
}

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
