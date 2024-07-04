package sql

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/mlflow/mlflow-go/pkg/config"
	"github.com/mlflow/mlflow-go/pkg/sql"
)

type TrackingSQLStore struct {
	logger *logrus.Logger
	config *config.Config
	db     *gorm.DB
}

func NewTrackingSQLStore(logger *logrus.Logger, config *config.Config) (*TrackingSQLStore, error) {
	database, err := sql.NewDatabase(logger, config.TrackingStoreURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database %q: %w", config.TrackingStoreURI, err)
	}

	return &TrackingSQLStore{logger: logger, config: config, db: database}, nil
}
