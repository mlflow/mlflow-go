package main

import "C"

import (
	"unsafe"

	"github.com/mlflow/mlflow-go/pkg/artifacts/service"
)

var artifactsServices = newInstanceMap[*service.ArtifactsService]()

//export CreateArtifactsService
func CreateArtifactsService(configData unsafe.Pointer, configSize C.int) int64 {
	//nolint:nlreturn
	return artifactsServices.Create(service.NewArtifactsService, C.GoBytes(configData, configSize))
}

//export DestroyArtifactsService
func DestroyArtifactsService(id int64) {
	artifactsServices.Destroy(id)
}
