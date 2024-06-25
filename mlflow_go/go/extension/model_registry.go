package main

import "C"

import (
	"unsafe"

	"github.com/mlflow/mlflow-go/mlflow_go/go/model_registry/service"
)

var modelRegistryServices = newServiceMap[service.ModelRegistryService]()

//export CreateModelRegistryService
func CreateModelRegistryService(configData unsafe.Pointer, configSize C.int) int64 {
	//nolint:nlreturn
	return modelRegistryServices.Create(service.NewModelRegistryService, C.GoBytes(configData, configSize))
}

//export DestroyModelRegistryService
func DestroyModelRegistryService(id int64) {
	modelRegistryServices.Destroy(id)
}
