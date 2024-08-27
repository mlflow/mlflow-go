package service //nolint:testpackage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/tracking/store"
	"github.com/mlflow/mlflow-go/pkg/utils"
)

type testRelativeArtifactLocationScenario struct {
	name  string
	input string
}

func TestRelativeArtifactLocation(t *testing.T) {
	t.Parallel()

	scenarios := []testRelativeArtifactLocationScenario{
		{name: "without scheme", input: "../yow"},
		{name: "with file scheme", input: "file:///../yow"},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()

			store := store.NewMockTrackingStore(t)
			store.EXPECT().CreateExperiment(context.Background(), mock.Anything).Return(mock.Anything, nil)

			service := TrackingService{
				Store: store,
			}

			input := protos.CreateExperiment{
				ArtifactLocation: utils.PtrTo(scenario.input),
			}

			response, err := service.CreateExperiment(context.Background(), &input)
			if err != nil {
				t.Error("expected create experiment to succeed")
			}

			if response == nil {
				t.Error("expected response to be non-nil")
			}

			if input.GetArtifactLocation() == scenario.input {
				t.Errorf("expected artifact location to be absolute, got %s", input.GetArtifactLocation())
			}
		})
	}
}
