package sql

import (
	"context"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/entities"
)

func (store TrackingSQLStore) LogInputs(
	ctx context.Context, runID string, datasets []*entities.DatasetInput,
) *contract.Error {
	return nil
}
