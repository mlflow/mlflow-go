package service

import (
	"context"
	"fmt"

	"github.com/mlflow/mlflow-go/pkg/config"
	"github.com/mlflow/mlflow-go/pkg/tracking/store"
	"github.com/mlflow/mlflow-go/pkg/tracking/store/sql"
)

type TrackingService struct {
	config *config.Config
	Store  store.TrackingStore
}

func NewTrackingService(ctx context.Context, config *config.Config) (*TrackingService, error) {
	store, err := sql.NewTrackingSQLStore(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create new sql store: %w", err)
	}

	return &TrackingService{
		config: config,
		Store:  store,
	}, nil
}

func (ts TrackingService) Destroy() error {
	if err := ts.Store.Destroy(); err != nil {
		return fmt.Errorf("failed to close store: %w", err)
	}

	return nil
}
