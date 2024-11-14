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
		// Very often Python tests mock generation of `request_id` of the flight.
		// It easily works with Python, but it doesn't work with GO,
		// so that's why we need to pass `request_id`
		// from Pythong to Go and override traceInfo.RequestID with value from Python.
		if tag.Key == "request_id" {
			traceInfo.RequestID = tag.Value
		} else {
			traceInfo.Tags = append(traceInfo.Tags, models.NewTraceTagFromEntity(traceInfo.RequestID, tag))
		}
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

const (
	BatchSize = 100
)

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
	if err := s.db.WithContext(
		ctx,
	).Where(
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

func (s TrackingSQLStore) GetTraceInfo(ctx context.Context, reqeustID string) (*entities.TraceInfo, *contract.Error) {
	var traceInfo models.TraceInfo
	if err := s.db.WithContext(
		ctx,
	).Where(
		"request_id = ?", reqeustID,
	).Preload(
		"Tags",
	).Preload(
		"TraceRequestMetadata",
	).First(
		&traceInfo,
	).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, contract.NewError(
				protos.ErrorCode_RESOURCE_DOES_NOT_EXIST,
				fmt.Sprintf(
					"Trace with request_id '%s' not found.",
					reqeustID,
				),
			)
		}

		return nil, contract.NewError(
			protos.ErrorCode_INTERNAL_ERROR, fmt.Sprintf("error getting trace info: %v", err),
		)
	}

	return traceInfo.ToEntity(), nil
}

func (s TrackingSQLStore) EndTrace(
	ctx context.Context,
	reqeustID string,
	timestampMS int64,
	status string,
	metadata []*entities.TraceRequestMetadata,
	tags []*entities.TraceTag,
) (*entities.TraceInfo, error) {
	traceInfo, err := s.GetTraceInfo(ctx, reqeustID)
	if err != nil {
		return nil, err
	}

	if err := s.db.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		if err := transaction.Model(
			&models.TraceInfo{},
		).Where(
			"request_id = ?", traceInfo.RequestID,
		).UpdateColumns(map[string]interface{}{
			"status":            status,
			"execution_time_ms": timestampMS - traceInfo.TimestampMS,
		}).Error; err != nil {
			return contract.NewErrorWith(
				protos.ErrorCode_INTERNAL_ERROR,
				fmt.Sprintf("failed to update trace with request_id '%s'", reqeustID),
				err,
			)
		}

		if err := s.createTraceTags(transaction, reqeustID, tags); err != nil {
			return err
		}

		if err := s.createTraceMetadata(transaction, reqeustID, metadata); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err //nolint
	}

	traceInfo, err = s.GetTraceInfo(ctx, reqeustID)
	if err != nil {
		return nil, err
	}

	return traceInfo, nil
}

func (s TrackingSQLStore) createTraceTags(transaction *gorm.DB, requestID string, tags []*entities.TraceTag) error {
	traceTags := make([]models.TraceTag, 0, len(tags))
	for _, tag := range tags {
		traceTags = append(traceTags, models.NewTraceTagFromEntity(requestID, tag))
	}

	if err := transaction.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(traceTags, batchSize).Error; err != nil {
		return contract.NewErrorWith(
			protos.ErrorCode_INTERNAL_ERROR,
			fmt.Sprintf("failed to update trace tags %v", err),
			err,
		)
	}

	return nil
}

func (s TrackingSQLStore) createTraceMetadata(
	transaction *gorm.DB, requestID string, metadata []*entities.TraceRequestMetadata,
) error {
	traceMetadata := make([]models.TraceRequestMetadata, 0, len(metadata))
	for _, m := range metadata {
		traceMetadata = append(traceMetadata, models.NewTraceRequestMetadataFromEntity(requestID, m))
	}

	if err := transaction.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(traceMetadata, batchSize).Error; err != nil {
		return contract.NewErrorWith(
			protos.ErrorCode_INTERNAL_ERROR,
			fmt.Sprintf("failed to update trace metadata %v", err),
			err,
		)
	}

	return nil
}

func (s TrackingSQLStore) DeleteTraces(
	ctx context.Context,
	experimentID string,
	maxTimestampMillis int64,
	maxTraces int32,
	requestIDs []string,
) (int32, *contract.Error) {
	query := s.db.WithContext(
		ctx,
	).Where(
		"experiment_id = ?", experimentID,
	)

	if maxTimestampMillis != 0 {
		query = query.Where("timestamp_ms <= ?", maxTimestampMillis)
	}

	if len(requestIDs) > 0 {
		query = query.Where("request_id IN (?)", requestIDs)
	}

	if maxTraces != 0 {
		query = query.Where(
			"request_id IN (?)",
			s.db.Select(
				"request_id",
			).Model(
				&models.TraceInfo{},
			).Order(
				"timestamp_ms ASC",
			).Limit(
				int(maxTraces),
			),
		)
	}

	var traces []models.TraceInfo
	if err := query.Debug().Clauses(
		clause.Returning{
			Columns: []clause.Column{
				{Name: "request_id"},
			},
		},
	).Delete(
		&traces,
	).Error; err != nil {
		return 0, contract.NewErrorWith(
			protos.ErrorCode_INTERNAL_ERROR,
			fmt.Sprintf("failed to delete traces %v", err),
			err,
		)
	}

	//nolint:gosec
	return int32(len(traces)), nil
}
