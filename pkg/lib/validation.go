package main

import (
	"sync"

	"github.com/go-playground/validator/v10"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/validation"
)

var getValidator = sync.OnceValues(func() (*validator.Validate, *contract.Error) {
	validate, err := validation.NewValidator()
	if err != nil {
		return nil, contract.NewError(
			protos.ErrorCode_INTERNAL_ERROR,
			err.Error(),
		)
	}

	return validate, nil
})
