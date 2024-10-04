package query_test

import (
	"strings"
	"testing"

	"github.com/mlflow/mlflow-go/pkg/tracking/service/query"
)

func TestValidQueries(t *testing.T) {
	t.Parallel()

	samples := []string{
		"metrics.foobar = 40",
		"metrics.foobar = 40 AND run_name = \"bouncy-boar-498\"",
		"tags.\"mlflow.source.name\" = \"scratch.py\"",
		"metrics.accuracy > 0.9",
		"params.\"random_state\" = \"8888\"",
		"params.`random_state` = \"8888\"",
		"params.solver ILIKE \"L%\"",
		"params.solver LIKE \"l%\"",
		"datasets.digest IN ('77a19fc0')",
		"attributes.run_id IN ('meh')",
	}

	for _, sample := range samples {
		currentSample := sample
		t.Run(currentSample, func(t *testing.T) {
			t.Parallel()

			_, err := query.ParseFilter(currentSample)
			if err != nil {
				t.Errorf("unexpected parse error: %v", err)
			}
		})
	}
}

type invalidSample struct {
	input         string
	expectedError string
}

//nolint:funlen
func TestInvalidQueries(t *testing.T) {
	t.Parallel()

	samples := []invalidSample{
		{
			input:         "yow.foobar = 40",
			expectedError: "invalid identifier",
		},
		{
			input: "attributes.foobar = 40",
			expectedError: "Invalid attribute key '{foobar}' specified. " +
				"Valid keys are '[run_id run_name user_id status start_time end_time artifact_uri]'",
		},
		{
			input: "datasets.foobar = 40",
			expectedError: "Invalid dataset key '{foobar}' specified. " +
				"Valid keys are '[run_id run_name user_id status start_time end_time artifact_uri]'",
		},
		{
			input:         "metric.yow = 'z'",
			expectedError: "expected numeric value type for metric.",
		},
		{
			input:         "parameter.tag = 2",
			expectedError: "expected a quoted string value",
		},
		{
			input:         "attributes.start_time = 'now'",
			expectedError: "expected numeric value type for numeric attribute",
		},
		{
			input:         "attributes.run_name IN ('foo','bar')",
			expectedError: "only the 'run_id' attribute supports comparison with a list",
		},
		{
			input:         "datasets.name = 40",
			expectedError: "expected datasets.name to be either a string or list of strings",
		},
		{
			input:         "datasets.digest = 50",
			expectedError: "expected datasets.digest to be either a string or list of strings",
		},
		{
			input:         "datasets.context = 60",
			expectedError: "expected datasets.context to be either a string or list of strings",
		},
	}

	for _, sample := range samples {
		currentSample := sample
		t.Run(currentSample.input, func(t *testing.T) {
			t.Parallel()

			_, err := query.ParseFilter(currentSample.input)
			if err == nil {
				t.Errorf("expected parse error but got nil")
			}

			if !strings.Contains(err.Error(), currentSample.expectedError) {
				t.Errorf(
					"expected error to contain %q, got %q",
					currentSample.expectedError,
					err.Error(),
				)
			}
		})
	}
}
