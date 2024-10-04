package parser

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/protos"
)

/*

This is the equivalent of type-checking the untyped tree.
Not every parsed tree is a valid one.

Grammar rule: identifier.key operator value

The rules are:

For identifiers:

identifier.key

Or if only key is passed, the identifier is "attribute"

Identifiers can have aliases.

if the identifier is dataset, the allowed keys are: name, digest and context.

*/

type ValidIdentifier int

const (
	Metric ValidIdentifier = iota
	Parameter
	Tag
	Attribute
	Dataset
)

func (v ValidIdentifier) String() string {
	switch v {
	case Metric:
		return "metric"
	case Parameter:
		return "parameter"
	case Tag:
		return "tag"
	case Attribute:
		return "attribute"
	case Dataset:
		return "dataset"
	default:
		return "unknown"
	}
}

type ValidCompareExpr struct {
	Identifier ValidIdentifier
	Key        string
	Operator   OperatorKind
	Value      interface{}
}

func (v ValidCompareExpr) String() string {
	return fmt.Sprintf("%s.%s %s %v", v.Identifier, v.Key, v.Operator, v.Value)
}

type ValidationError struct {
	message string
}

func (e *ValidationError) Error() string {
	return e.message
}

func NewValidationError(format string, a ...interface{}) *ValidationError {
	return &ValidationError{message: fmt.Sprintf(format, a...)}
}

const (
	metricIdentifier    = "metric"
	parameterIdentifier = "parameter"
	tagIdentifier       = "tag"
	attributeIdentifier = "attribute"
	datasetIdentifier   = "dataset"
)

var identifiers = []string{
	metricIdentifier,
	parameterIdentifier,
	tagIdentifier,
	attributeIdentifier,
	datasetIdentifier,
}

func parseValidIdentifier(identifier string) (ValidIdentifier, error) {
	switch identifier {
	case metricIdentifier, "metrics":
		return Metric, nil
	case parameterIdentifier, "parameters", "param", "params":
		return Parameter, nil
	case tagIdentifier, "tags":
		return Tag, nil
	case "", attributeIdentifier, "attr", "attributes", "run":
		return Attribute, nil
	case datasetIdentifier, "datasets":
		return Dataset, nil
	default:
		return -1, NewValidationError("invalid identifier %q", identifier)
	}
}

const (
	RunID     = "run_id"
	RunName   = "run_name"
	Created   = "created"
	StartTime = "start_time"
)

// This should be configurable and only applies to the runs table.
var searchableRunAttributes = []string{
	RunID,
	RunName,
	"user_id",
	"status",
	StartTime,
	"end_time",
	"artifact_uri",
}

var datasetAttributes = []string{"name", "digest", "context"}

func parseAttributeKey(key string) (string, error) {
	switch key {
	case "run_id":
		// We return run_uuid before that is the SQL column name.
		return "run_uuid", nil
	case
		"user_id",
		"status",
		StartTime,
		"end_time",
		"artifact_uri":
		return key, nil
	case Created, "Created":
		return StartTime, nil
	case RunName, "run name", "Run name", "Run Name":
		return RunName, nil
	default:
		return "", contract.NewError(protos.ErrorCode_BAD_REQUEST,
			fmt.Sprintf(
				"Invalid attribute key '{%s}' specified. Valid keys are '%v'",
				key,
				searchableRunAttributes,
			),
		)
	}
}

func parseKey(identifier ValidIdentifier, key string) (string, error) {
	if key == "" {
		return attributeIdentifier, nil
	}

	//nolint:exhaustive
	switch identifier {
	case Attribute:
		return parseAttributeKey(key)
	case Dataset:
		switch key {
		case "name", "digest", "context":
			return key, nil
		default:
			return "", contract.NewError(protos.ErrorCode_BAD_REQUEST,
				fmt.Sprintf(
					"Invalid dataset key '{%s}' specified. Valid keys are '%v'",
					key,
					searchableRunAttributes,
				),
			)
		}
	default:
		return key, nil
	}
}

// Returns a standardized LongIdentifierExpr.
func validatedIdentifier(identifier *Identifier) (ValidIdentifier, string, error) {
	validIdentifier, err := parseValidIdentifier(identifier.Identifier)
	if err != nil {
		return -1, "", err
	}

	validKey, err := parseKey(validIdentifier, identifier.Key)
	if err != nil {
		return -1, "", err
	}

	identifier.Key = validKey

	return validIdentifier, validKey, nil
}

/*

The value part is determined by the identifier

"metric" takes numbers
"parameter" and "tag" takes strings

"attribute" could be either string or number,
number when StartTime, "end_time" or "created", "Created"
otherwise string

"dataset" takes strings for "name", "digest" and "context"

*/

func validateDatasetValue(key string, value Value) (interface{}, error) {
	switch key {
	case "name", "digest", "context":
		if _, ok := value.(NumberExpr); ok {
			return nil, NewValidationError(
				"expected datasets.%s to be either a string or list of strings. Found %s",
				key,
				value,
			)
		}

		return value.value(), nil
	default:
		return nil, NewValidationError(
			"expected dataset attribute key to be one of %s. Found %s",
			strings.Join(datasetAttributes, ", "),
			key,
		)
	}
}

// Port of _get_value in search_utils.py.
func validateValue(identifier ValidIdentifier, key string, value Value) (interface{}, error) {
	switch identifier {
	case Metric:
		if _, ok := value.(NumberExpr); !ok {
			return nil, NewValidationError(
				"expected numeric value type for metric. Found %s",
				value,
			)
		}

		return value.value(), nil
	case Parameter, Tag:
		if _, ok := value.(StringExpr); !ok {
			return nil, NewValidationError(
				"expected a quoted string value for %s. Found %s",
				identifier, value,
			)
		}

		return value.value(), nil
	case Attribute:
		value, err := validateAttributeValue(key, value)

		return value, err
	case Dataset:
		return validateDatasetValue(key, value)
	default:
		return nil, NewValidationError(
			"Invalid identifier type %s. Expected one of %s",
			identifier,
			strings.Join(identifiers, ", "),
		)
	}
}

func validateAttributeValue(key string, value Value) (interface{}, error) {
	switch key {
	case StartTime, "end_time", Created:
		if _, ok := value.(NumberExpr); !ok {
			return nil, NewValidationError(
				"expected numeric value type for numeric attribute: %s. Found %s",
				key,
				value,
			)
		}

		return value.value(), nil
	default:
		// run_id was earlier converted to run_uuid
		if _, ok := value.(StringListExpr); key != "run_uuid" && ok {
			return nil, NewValidationError(
				"only the 'run_id' attribute supports comparison with a list of quoted string values",
			)
		}

		return value.value(), nil
	}
}

// Validate an expression according to the mlflow domain.
// This represent is a simple type-checker for the expression.
// Not every identifier is valid according to the mlflow domain.
// The same for the value part.
func ValidateExpression(expression *CompareExpr) (*ValidCompareExpr, error) {
	validIdentifier, validKey, err := validatedIdentifier(&expression.Left)
	if err != nil {
		var contractError *contract.Error
		if errors.As(err, &contractError) {
			return nil, contractError
		}

		return nil, fmt.Errorf("Error on parsing filter expression: %w", err)
	}

	value, err := validateValue(validIdentifier, validKey, expression.Right)
	if err != nil {
		return nil, fmt.Errorf("Error on parsing filter expression: %w", err)
	}

	return &ValidCompareExpr{
		Identifier: validIdentifier,
		Key:        validKey,
		Operator:   expression.Operator,
		Value:      value,
	}, nil
}
