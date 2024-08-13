package main

import "C"

import (
	"context"
	"encoding/json"
	"unsafe"

	"google.golang.org/protobuf/proto"

	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/validation"
)

func unmarshalAndValidateProto(
	data []byte,
	msg proto.Message,
) *contract.Error {
	if err := proto.Unmarshal(data, msg); err != nil {
		return contract.NewError(
			protos.ErrorCode_BAD_REQUEST,
			err.Error(),
		)
	}

	validate, cErr := getValidator()
	if cErr != nil {
		return cErr
	}

	if err := validate.Struct(msg); err != nil {
		return validation.NewErrorFromValidationError(err)
	}

	return nil
}

func marshalProto(msg proto.Message) ([]byte, *contract.Error) {
	res, err := proto.Marshal(msg)
	if err != nil {
		return nil, contract.NewError(
			protos.ErrorCode_INTERNAL_ERROR,
			err.Error(),
		)
	}

	return res, nil
}

func makePointerFromBytes(data []byte, size *C.int) unsafe.Pointer {
	*size = C.int(len(data))

	return C.CBytes(data) //nolint:nlreturn
}

func makePointerFromError(err *contract.Error, size *C.int) unsafe.Pointer {
	data, _ := json.Marshal(err) //nolint:errchkjson

	return makePointerFromBytes(data, size)
}

// invokeServiceMethod is a helper function that invokes a service method and handles
// marshalling/unmarshalling of request/response data through the FFI boundary.
func invokeServiceMethod[I, O proto.Message](
	serviceMethod func(context.Context, I) (O, *contract.Error),
	request I,
	requestData unsafe.Pointer,
	requestSize C.int,
	responseSize *C.int,
) unsafe.Pointer {
	requestBytes := C.GoBytes(requestData, requestSize) //nolint:nlreturn

	err := unmarshalAndValidateProto(requestBytes, request)
	if err != nil {
		return makePointerFromError(err, responseSize)
	}

	response, err := serviceMethod(context.Background(), request)
	if err != nil {
		return makePointerFromError(err, responseSize)
	}

	responseBytes, err := marshalProto(response)
	if err != nil {
		return makePointerFromError(err, responseSize)
	}

	return makePointerFromBytes(responseBytes, responseSize)
}
