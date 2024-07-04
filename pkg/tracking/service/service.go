package service

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/mlflow/mlflow-go/pkg/config"
	"github.com/mlflow/mlflow-go/pkg/tracking/store"
	"github.com/mlflow/mlflow-go/pkg/tracking/store/sql"
)

type TrackingService struct {
	config *config.Config
	Store  store.TrackingStore
}

func NewTrackingService(logger *logrus.Logger, config *config.Config) (*TrackingService, error) {
	store, err := sql.NewTrackingSQLStore(logger, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create new sql store: %w", err)
	}

	return &TrackingService{
		config: config,
		Store:  store,
	}, nil
}
