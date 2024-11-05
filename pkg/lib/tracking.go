package main

// #include <stdlib.h>
import "C"

import (
	"unsafe"

	"github.com/mlflow/mlflow-go/pkg/tracking/service"
)

var trackingServices = newInstanceMap[*service.TrackingService]()

//export CreateTrackingService
func CreateTrackingService(configData unsafe.Pointer, configSize C.int) int64 {
	//nolint:nlreturn
	return trackingServices.Create(
		service.NewTrackingService, C.GoBytes(configData, configSize),
	)
}

//export DestroyTrackingService
func DestroyTrackingService(id int64) {
	trackingServices.Destroy(id)
}

//export FreeResponse
func FreeResponse(pointer *int64) {
	C.free(unsafe.Pointer(pointer))
}
