package sql

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/mlflow/mlflow-go/pkg/config"
	"github.com/mlflow/mlflow-go/pkg/sql"
)

type TrackingSQLStore struct {
	config *config.Config
	db     *gorm.DB
}

func NewTrackingSQLStore(ctx context.Context, config *config.Config) (*TrackingSQLStore, error) {
	database, err := sql.NewDatabase(ctx, config.TrackingStoreURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database %q: %w", config.TrackingStoreURI, err)
	}

	return &TrackingSQLStore{
		config: config,
		db:     database,
	}, nil
}

func (ts TrackingSQLStore) Close() error {
	return sql.CloseDatabase(ts.db)
}
