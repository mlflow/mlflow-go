package sql

import (
	"context"

	"gorm.io/gorm/clause"

	"github.com/mlflow/mlflow-go/pkg/tracking/store/sql/models"
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
