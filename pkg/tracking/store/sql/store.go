package sql

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/mlflow/mlflow-go/pkg/config"
	"github.com/mlflow/mlflow-go/pkg/sql"
	"github.com/mlflow/mlflow-go/pkg/utils"
)

type TrackingSQLStore struct {
	config *config.Config
	db     *gorm.DB
	logger *logrus.Logger
}

func NewTrackingSQLStore(ctx context.Context, config *config.Config) (*TrackingSQLStore, error) {
	logger := utils.GetLoggerFromContext(ctx)

	database, err := sql.NewDatabase(ctx, config.TrackingStoreURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database %q: %w", config.TrackingStoreURI, err)
	}

	return &TrackingSQLStore{
		config: config,
		db:     database,
		logger: logger,
	}, nil
}
