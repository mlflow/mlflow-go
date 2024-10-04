package parser

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/validation"
)

type HTTPRequestParser struct {
	validator *validator.Validate
}

func NewHTTPRequestParser() (*HTTPRequestParser, error) {
	validator, err := validation.NewValidator()
	if err != nil {
		return nil, fmt.Errorf("failed to create validator: %w", err)
	}

	return &HTTPRequestParser{
		validator: validator,
	}, nil
}

func (p *HTTPRequestParser) ParseBody(ctx *fiber.Ctx, input proto.Message) *contract.Error {
	if protojsonErr := protojson.Unmarshal(ctx.Body(), input); protojsonErr != nil {
		// falling back to JSON, because `protojson` doesn't provide any information
		// about `field` name for which ut fails. MLFlow tests expect to know the exact
		// `field` name where validation failed. This approach has no effect on MLFlow
		// tests, so let's keep it for now.
		if jsonErr := json.Unmarshal(ctx.Body(), input); jsonErr != nil {
			var unmarshalTypeError *json.UnmarshalTypeError
			if errors.As(jsonErr, &unmarshalTypeError) {
				result := gjson.GetBytes(ctx.Body(), unmarshalTypeError.Field)

				return contract.NewError(
					protos.ErrorCode_INVALID_PARAMETER_VALUE,
					fmt.Sprintf("Invalid value %s for parameter '%s' supplied", result.Raw, unmarshalTypeError.Field),
				)
			}
		}

		return contract.NewError(protos.ErrorCode_BAD_REQUEST, protojsonErr.Error())
	}

	if err := p.validator.Struct(input); err != nil {
		return validation.NewErrorFromValidationError(err)
	}

	return nil
}

func (p *HTTPRequestParser) ParseQuery(ctx *fiber.Ctx, input interface{}) *contract.Error {
	if err := ctx.QueryParser(input); err != nil {
		return contract.NewError(protos.ErrorCode_BAD_REQUEST, err.Error())
	}

	if err := p.validator.Struct(input); err != nil {
		return validation.NewErrorFromValidationError(err)
	}

	return nil
}
