package validation

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/protos"
)

const (
	QuoteLength              = 2
	MaxEntitiesPerBatch      = 1000
	MaxValidationInputLength = 100
)

// regex for valid param and metric names: may only contain slashes, alphanumerics,
// underscores, periods, dashes, and spaces.
var paramAndMetricNameRegex = regexp.MustCompile(`^[/\w.\- ]*$`)

// regex for valid run IDs: must be an alphanumeric string of length 1 to 256.
var runIDRegex = regexp.MustCompile(`^[a-zA-Z0-9][\w\-]{0,255}$`)

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

func truncateFn(fieldLevel validator.FieldLevel) bool {
	param := fieldLevel.Param() // Get the parameter from the tag

	maxLength, err := strconv.Atoi(param)
	if err != nil {
		return false // If the parameter isn't a valid integer, fail the validation.
	}

	truncateLongValues, shouldTruncate := os.LookupEnv("MLFLOW_TRUNCATE_LONG_VALUES")
	shouldTruncate = shouldTruncate && truncateLongValues == "true"

	field := fieldLevel.Field()

	if field.Kind() == reflect.String {
		strValue := field.String()
		if len(strValue) <= maxLength {
			return true
		}

		if shouldTruncate {
			field.SetString(strValue[:maxLength])

			return true
		}

		return false
	}

	return true
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

	if err := validate.RegisterValidation("truncate", truncateFn); err != nil {
		return nil, fmt.Errorf("validation registration for 'truncateFn' failed: %w", err)
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

func getErrorPath(err validator.FieldError) string {
	path := err.Field()

	if err.Namespace() != "" {
		// Strip first item in struct namespace
		idx := strings.Index(err.Namespace(), ".")
		if idx != -1 {
			path = err.Namespace()[(idx + 1):]
		}
	}

	return path
}

func constructValidationError(field string, value any, suffix string) string {
	formattedValue, err := json.Marshal(value)
	if err != nil {
		formattedValue = []byte(fmt.Sprintf("%v", value))
	}

	return fmt.Sprintf("Invalid value %s for parameter '%s' supplied%s", formattedValue, field, suffix)
}

func mkTruncateValidationError(field string, value interface{}, err validator.FieldError) string {
	strValue, ok := value.(string)
	if ok {
		expected := len(strValue)

		if expected > MaxValidationInputLength {
			strValue = strValue[:MaxValidationInputLength] + "..."
		}

		return constructValidationError(
			field,
			strValue,
			fmt.Sprintf(": length %d exceeded length limit of %s", expected, err.Param()),
		)
	}

	return constructValidationError(field, value, "")
}

func mkMaxValidationError(field string, value interface{}, err validator.FieldError) string {
	if _, ok := value.(string); ok {
		return fmt.Sprintf(
			"'%s' exceeds the maximum length of %s characters",
			field,
			err.Param(),
		)
	}

	return constructValidationError(field, value, "")
}

func NewErrorFromValidationError(err error) *contract.Error {
	var validatorValidationErrors validator.ValidationErrors
	if errors.As(err, &validatorValidationErrors) {
		validationErrors := make([]string, 0)

		for _, err := range validatorValidationErrors {
			field := getErrorPath(err)
			tag := err.Tag()
			value := dereference(err.Value())

			switch tag {
			case "required":
				validationErrors = append(
					validationErrors,
					fmt.Sprintf("Missing value for required parameter '%s'", field),
				)
			case "truncate":
				validationErrors = append(validationErrors, mkTruncateValidationError(field, value, err))
			case "uniqueParams":
				validationErrors = append(
					validationErrors,
					"Duplicate parameter keys have been submitted",
				)
			case "max":
				validationErrors = append(validationErrors, mkMaxValidationError(field, value, err))
			default:
				validationErrors = append(
					validationErrors,
					constructValidationError(field, value, ""),
				)
			}
		}

		return contract.NewError(protos.ErrorCode_INVALID_PARAMETER_VALUE, strings.Join(validationErrors, ", "))
	}

	return contract.NewError(protos.ErrorCode_INTERNAL_ERROR, err.Error())
}
