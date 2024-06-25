package parser

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/tidwall/gjson"

	"github.com/mlflow/mlflow-go/mlflow_go/go/contract"
	"github.com/mlflow/mlflow-go/mlflow_go/go/protos"
	"github.com/mlflow/mlflow-go/mlflow_go/go/validation"
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

func (p *HTTPRequestParser) ParseBody(ctx *fiber.Ctx, input interface{}) *contract.Error {
	if err := ctx.BodyParser(input); err != nil {
		var unmarshalTypeError *json.UnmarshalTypeError
		if errors.As(err, &unmarshalTypeError) {
			result := gjson.GetBytes(ctx.Body(), unmarshalTypeError.Field)

			return contract.NewError(
				protos.ErrorCode_INVALID_PARAMETER_VALUE,
				fmt.Sprintf("Invalid value %s for parameter '%s' supplied", result.Raw, unmarshalTypeError.Field),
			)
		}

		return contract.NewError(protos.ErrorCode_BAD_REQUEST, err.Error())
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
