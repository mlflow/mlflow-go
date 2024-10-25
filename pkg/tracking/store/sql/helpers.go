package sql

import (
	"fmt"

	"github.com/mlflow/mlflow-go/pkg/entities"
	"github.com/mlflow/mlflow-go/pkg/tracking/store/sql/models"
	"github.com/mlflow/mlflow-go/pkg/utils"
)

const (
	TraceFolderName     = "traces"
	ArtifactsFolderName = "artifacts"
)

const (
	MlflowArtifactLocation = "mlflow.artifactLocation"
)

func GetTraceArtifactLocationTag(
	experiment *entities.Experiment, requestID string,
) (models.TraceTag, error) {
	traceArtifactLocationTag, err := utils.AppendToURIPath(
		experiment.ArtifactLocation,
		TraceFolderName,
		requestID,
		ArtifactsFolderName,
	)
	if err != nil {
		return models.TraceTag{}, fmt.Errorf("failed to create trace artifact location tag: %w", err)
	}

	return models.TraceTag{
		Key:       MlflowArtifactLocation,
		Value:     traceArtifactLocationTag,
		RequestID: requestID,
	}, nil
}
