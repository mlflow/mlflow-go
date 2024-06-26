package sql

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/mlflow/mlflow-go/mlflow_go/go/config"
	"github.com/mlflow/mlflow-go/mlflow_go/go/sql"

	_ "github.com/ncruces/go-sqlite3/embed" // embed sqlite3 driver
)

type TrackingSQLStore struct {
	config *config.Config
	db     *gorm.DB
}

func NewTrackingSQLStore(logger *logrus.Logger, config *config.Config) (*TrackingSQLStore, error) {
	database, err := sql.NewDatabase(logger, config.TrackingStoreURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database %q: %w", config.TrackingStoreURI, err)
	}

	return &TrackingSQLStore{config: config, db: database}, nil
}
