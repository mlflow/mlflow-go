package service //nolint:testpackage

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/valyala/fasthttp"

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
			store.EXPECT().CreateExperiment(mock.Anything).Return(mock.Anything, nil)

			service := TrackingService{
				Store: store,
			}

			input := protos.CreateExperiment{
				ArtifactLocation: utils.PtrTo(scenario.input),
			}

			ctx := fiber.New().AcquireCtx(&fasthttp.RequestCtx{})

			response, err := service.CreateExperiment(ctx, &input)
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
