package sql

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/mlflow/mlflow-go/mlflow_go/go/config"
	"github.com/mlflow/mlflow-go/mlflow_go/go/sql"
)

type ModelRegistrySQLStore struct {
	config *config.Config
	db     *gorm.DB
}

func NewModelRegistrySQLStore(logger *logrus.Logger, config *config.Config) (*ModelRegistrySQLStore, error) {
	database, err := sql.NewDatabase(logger, config.ModelRegistryStoreURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database %q: %w", config.ModelRegistryStoreURI, err)
	}

	return &ModelRegistrySQLStore{config: config, db: database}, nil
}
