package validation

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/mlflow/mlflow-go/mlflow_go/go/contract"
	"github.com/mlflow/mlflow-go/mlflow_go/go/protos"
)

const (
	QuoteLength         = 2
	MaxEntitiesPerBatch = 1000
)

// regex for valid param and metric names: may only contain slashes, alphanumerics,
// underscores, periods, dashes, and spaces.
var paramAndMetricNameRegex = regexp.MustCompile(`^[/\w.\- ]*$`)

// regex for valid run IDs: must be an alphanumeric string of length 1 to 256.
var runIDRegex = regexp.MustCompile(`^[a-zA-Z0-9][\w\-]{0,255}$`)

func getValue(x reflect.Value) reflect.Value {
	if x.Kind() == reflect.Pointer {
		return x.Elem()
	}

	return x
}

func validateNested(validate *validator.Validate, current reflect.Value) bool {
	val := getValue(current)

	//nolint:exhaustive
	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		for i := range val.Len() {
			if !validateNested(validate, val.Index(i)) {
				return false
			}
		}

	case reflect.Struct:
		if err := validate.Struct(val.Interface()); err != nil {
			return false
		}
	default:
		if err := validate.Var(val.Interface(), ""); err != nil {
			return false
		}
	}

	return true
}

func stringAsPositiveIntegerValidation(fl validator.FieldLevel) bool {
	valueStr := fl.Field().String()

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return false
	}

	return value > -1
}

func uriWithoutFragmentsOrParamsOrDotDotInQueryValidation(fl validator.FieldLevel) bool {
	valueStr := fl.Field().String()
	if valueStr == "" {
		return true
	}

	u, err := url.Parse(valueStr)
	if err != nil {
		return false
	}

	return u.Fragment == "" && u.RawQuery == "" && !strings.Contains(u.RawQuery, "..")
}

func uniqueParamsValidation(fl validator.FieldLevel) bool {
	value := fl.Field()

	params, areParams := value.Interface().([]*protos.Param)
	if !areParams || len(params) == 0 {
		return true
	}

	hasDuplicates := false
	keys := make(map[string]bool, len(params))

	for _, param := range params {
		if _, ok := keys[param.GetKey()]; ok {
			hasDuplicates = true

			break
		}

		keys[param.GetKey()] = true
	}

	return !hasDuplicates
}

func pathIsClean(fl validator.FieldLevel) bool {
	valueStr := fl.Field().String()
	norm := filepath.Clean(valueStr)

	return !(norm != valueStr || norm == "." || strings.HasPrefix(norm, "..") || strings.HasPrefix(norm, "/"))
}

func regexValidation(regex *regexp.Regexp) validator.Func {
	return func(fl validator.FieldLevel) bool {
		valueStr := fl.Field().String()

		return regex.MatchString(valueStr)
	}
}

// see _validate_batch_log_limits in validation.py.
func validateLogBatchLimits(structLevel validator.StructLevel) {
	logBatch, isLogBatch := structLevel.Current().Interface().(*protos.LogBatch)

	if isLogBatch {
		total := len(logBatch.GetParams()) + len(logBatch.GetMetrics()) + len(logBatch.GetTags())
		if total > MaxEntitiesPerBatch {
			structLevel.ReportError(&logBatch, "metrics, params, and tags", "", "", "")
		}
	}
}

func NewValidator() (*validator.Validate, error) {
	validate := validator.New()

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", QuoteLength)[0]
		// skip if tag key says it should be ignored
		if name == "-" {
			return ""
		}

		return name
	})

	// Validate nested content of a struct field while reporting a problem on the current level.
	if err := validate.RegisterValidation(
		"dip",
		func(fl validator.FieldLevel) bool {
			val := fl.Field()

			return validateNested(validate, val)
		},
	); err != nil {
		return nil, fmt.Errorf("validation registration for 'dip' failed: %w", err)
	}

	// Verify that the input string is a positive integer.
	if err := validate.RegisterValidation(
		"stringAsPositiveInteger", stringAsPositiveIntegerValidation,
	); err != nil {
		return nil, fmt.Errorf("validation registration for 'stringAsPositiveInteger' failed: %w", err)
	}

	// Verify that the input string, if present, is a Url without fragment or query parameters
	if err := validate.RegisterValidation(
		"uriWithoutFragmentsOrParamsOrDotDotInQuery", uriWithoutFragmentsOrParamsOrDotDotInQueryValidation); err != nil {
		return nil, fmt.Errorf("validation registration for 'uriWithoutFragmentsOrParamsOrDotDotInQuery' failed: %w", err)
	}

	if err := validate.RegisterValidation(
		"validMetricParamOrTagName", regexValidation(paramAndMetricNameRegex),
	); err != nil {
		return nil, fmt.Errorf("validation registration for 'validMetricParamOrTagName' failed: %w", err)
	}

	if err := validate.RegisterValidation("pathIsUnique", pathIsClean); err != nil {
		return nil, fmt.Errorf("validation registration for 'validMetricParamOrTagValue' failed: %w", err)
	}

	// unique params in LogBatch
	if err := validate.RegisterValidation("uniqueParams", uniqueParamsValidation); err != nil {
		return nil, fmt.Errorf("validation registration for 'uniqueParams' failed: %w", err)
	}

	if err := validate.RegisterValidation("runId", regexValidation(runIDRegex)); err != nil {
		return nil, fmt.Errorf("validation registration for 'runId' failed: %w", err)
	}

	validate.RegisterStructValidation(validateLogBatchLimits, &protos.LogBatch{})

	return validate, nil
}

func dereference(value interface{}) interface{} {
	valueOf := reflect.ValueOf(value)
	if valueOf.Kind() == reflect.Ptr {
		if valueOf.IsNil() {
			return ""
		}

		return valueOf.Elem().Interface()
	}

	return value
}

func NewErrorFromValidationError(err error) *contract.Error {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		validationErrors := make([]string, 0)

		for _, err := range ve {
			field := err.Field()
			tag := err.Tag()
			value := dereference(err.Value())

			switch tag {
			case "required":
				validationErrors = append(
					validationErrors,
					fmt.Sprintf("Missing value for required parameter '%s'", field),
				)
			default:
				formattedValue, err := json.Marshal(value)
				if err != nil {
					formattedValue = []byte(fmt.Sprintf("%v", value))
				}

				validationErrors = append(
					validationErrors,
					fmt.Sprintf("Invalid value %s for parameter '%s' supplied", formattedValue, field),
				)
			}
		}

		return contract.NewError(protos.ErrorCode_INVALID_PARAMETER_VALUE, strings.Join(validationErrors, ", "))
	}

	return contract.NewError(protos.ErrorCode_INTERNAL_ERROR, err.Error())
}
