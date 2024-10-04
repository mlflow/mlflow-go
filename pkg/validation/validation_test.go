package validation_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/utils"
	"github.com/mlflow/mlflow-go/pkg/validation"
)

type PositiveInteger struct {
	Value string `validate:"stringAsPositiveInteger"`
}

type validationScenario struct {
	name          string
	input         any
	shouldTrigger bool
}

func runscenarios(t *testing.T, scenarios []validationScenario) {
	t.Helper()

	validator, err := validation.NewValidator()
	require.NoError(t, err)

	for _, scenario := range scenarios {
		currentScenario := scenario
		t.Run(currentScenario.name, func(t *testing.T) {
			t.Parallel()

			errs := validator.Struct(currentScenario.input)

			if currentScenario.shouldTrigger && errs == nil {
				t.Errorf("Expected validation error, got nil")
			}

			if !currentScenario.shouldTrigger && errs != nil {
				t.Errorf("Expected no validation error, got %v", errs)
			}
		})
	}
}

func TestStringAsPositiveInteger(t *testing.T) {
	t.Parallel()

	scenarios := []validationScenario{
		{
			name:          "positive integer",
			input:         PositiveInteger{Value: "1"},
			shouldTrigger: false,
		},
		{
			name:          "zero",
			input:         PositiveInteger{Value: "0"},
			shouldTrigger: false,
		},
		{
			name:          "negative integer",
			input:         PositiveInteger{Value: "-1"},
			shouldTrigger: true,
		},
		{
			name:          "alphabet",
			input:         PositiveInteger{Value: "a"},
			shouldTrigger: true,
		},
	}

	runscenarios(t, scenarios)
}

type uriWithoutFragmentsOrParams struct {
	Value string `validate:"uriWithoutFragmentsOrParamsOrDotDotInQuery"`
}

func TestUriWithoutFragmentsOrParams(t *testing.T) {
	t.Parallel()

	scenarios := []validationScenario{
		{
			name:          "valid url",
			input:         uriWithoutFragmentsOrParams{Value: "http://example.com"},
			shouldTrigger: false,
		},
		{
			name:          "only trigger when url is not empty",
			input:         uriWithoutFragmentsOrParams{Value: ""},
			shouldTrigger: false,
		},
		{
			name:          "url with fragment",
			input:         uriWithoutFragmentsOrParams{Value: "http://example.com#fragment"},
			shouldTrigger: true,
		},
		{
			name:          "url with query parameters",
			input:         uriWithoutFragmentsOrParams{Value: "http://example.com?query=param"},
			shouldTrigger: true,
		},
		{
			name:          "unparsable url",
			input:         uriWithoutFragmentsOrParams{Value: ":invalid-url"},
			shouldTrigger: true,
		},
		{
			name:          ".. in query",
			input:         uriWithoutFragmentsOrParams{Value: "http://example.com?query=./.."},
			shouldTrigger: true,
		},
	}

	runscenarios(t, scenarios)
}

func TestUniqueParamsInLogBatch(t *testing.T) {
	t.Parallel()

	logBatchRequest := &protos.LogBatch{
		Params: []*protos.Param{
			{Key: utils.PtrTo("key1"), Value: utils.PtrTo("value1")},
			{Key: utils.PtrTo("key1"), Value: utils.PtrTo("value2")},
		},
	}

	validator, err := validation.NewValidator()
	require.NoError(t, err)

	err = validator.Struct(logBatchRequest)
	if err == nil {
		t.Error("Expected uniqueParams validation error, got none")
	}
}

func TestEmptyParamsInLogBatch(t *testing.T) {
	t.Parallel()

	logBatchRequest := &protos.LogBatch{
		RunId:  utils.PtrTo("odcppTsGTMkHeDcqfZOYDMZSf"),
		Params: make([]*protos.Param, 0),
	}

	validator, err := validation.NewValidator()
	require.NoError(t, err)

	err = validator.Struct(logBatchRequest)
	if err != nil {
		t.Errorf("Unexpected uniqueParams validation error, got %v", err)
	}
}

func TestMissingTimestampInNestedMetric(t *testing.T) {
	t.Parallel()

	serverValidator, err := validation.NewValidator()
	require.NoError(t, err)

	logBatch := protos.LogBatch{
		RunId: utils.PtrTo("odcppTsGTMkHeDcqfZOYDMZSf"),
		Metrics: []*protos.Metric{
			{
				Key:   utils.PtrTo("mae"),
				Value: utils.PtrTo(2.5),
			},
		},
	}

	err = serverValidator.Struct(&logBatch)
	if err == nil {
		t.Error("Expected dive validation error, got none")
	}

	msg := validation.NewErrorFromValidationError(err).Message
	if !strings.Contains(msg, "metrics[0].timestamp") {
		t.Errorf("Expected required validation error for nested property, got %v", msg)
	}
}

type avecTruncate struct {
	X *string `validate:"truncate=3"`
	Y string  `validate:"truncate=3"`
}

func TestTruncate(t *testing.T) {
	input := &avecTruncate{
		X: utils.PtrTo("123456"),
		Y: "654321",
	}

	t.Setenv("MLFLOW_TRUNCATE_LONG_VALUES", "true")

	validator, err := validation.NewValidator()
	require.NoError(t, err)

	err = validator.Struct(input)
	require.NoError(t, err)

	if len(*input.X) != 3 {
		t.Errorf("Expected the length of x to be 3, was %d", len(*input.X))
	}

	if len(input.Y) != 3 {
		t.Errorf("Expected the length of y to be 3, was %d", len(input.Y))
	}
}

// This unit test is a sanity test that confirms the `dive` validation
// enters a nested slice of pointer structs.
func TestNestedErrorsInSubCollection(t *testing.T) {
	t.Parallel()

	value := strings.Repeat("X", 6001) + "Y"

	logBatchRequest := &protos.LogBatch{
		RunId: utils.PtrTo("odcppTsGTMkHeDcqfZOYDMZSf"),
		Params: []*protos.Param{
			{Key: utils.PtrTo("key1"), Value: utils.PtrTo(value)},
			{Key: utils.PtrTo("key2"), Value: utils.PtrTo(value)},
		},
	}

	validator, err := validation.NewValidator()
	require.NoError(t, err)

	err = validator.Struct(logBatchRequest)
	if err != nil {
		msg := validation.NewErrorFromValidationError(err).Message
		// Assert the root struct name is not present in the error message
		if strings.Contains(msg, "logBatch") {
			t.Errorf("Validation message contained root struct name, got %s", msg)
		}

		// Assert the index is listed in the parameter path
		if !strings.Contains(msg, "params[0].value") ||
			!strings.Contains(msg, "params[1].value") ||
			!strings.Contains(msg, "length 6002 exceeded length limit of 6000") {
			t.Errorf("Unexpected validation error message, got %s", msg)
		}
	}
}
