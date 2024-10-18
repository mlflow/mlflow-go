// Code generated by mlflow/go/cmd/generate/main.go. DO NOT EDIT.

package main

import "C"
import (
	"unsafe"
	"github.com/mlflow/mlflow-go/pkg/protos"
)
//export TrackingServiceGetExperimentByName
func TrackingServiceGetExperimentByName(serviceID int64, requestData unsafe.Pointer, requestSize C.int, responseSize *C.int) unsafe.Pointer {
	service, err := trackingServices.Get(serviceID)
	if err != nil {
		return makePointerFromError(err, responseSize)
	}
	return invokeServiceMethod(service.GetExperimentByName, new(protos.GetExperimentByName), requestData, requestSize, responseSize)
}
//export TrackingServiceCreateExperiment
func TrackingServiceCreateExperiment(serviceID int64, requestData unsafe.Pointer, requestSize C.int, responseSize *C.int) unsafe.Pointer {
	service, err := trackingServices.Get(serviceID)
	if err != nil {
		return makePointerFromError(err, responseSize)
	}
	return invokeServiceMethod(service.CreateExperiment, new(protos.CreateExperiment), requestData, requestSize, responseSize)
}
//export TrackingServiceSearchExperiments
func TrackingServiceSearchExperiments(serviceID int64, requestData unsafe.Pointer, requestSize C.int, responseSize *C.int) unsafe.Pointer {
	service, err := trackingServices.Get(serviceID)
	if err != nil {
		return makePointerFromError(err, responseSize)
	}
	return invokeServiceMethod(service.SearchExperiments, new(protos.SearchExperiments), requestData, requestSize, responseSize)
}
//export TrackingServiceGetExperiment
func TrackingServiceGetExperiment(serviceID int64, requestData unsafe.Pointer, requestSize C.int, responseSize *C.int) unsafe.Pointer {
	service, err := trackingServices.Get(serviceID)
	if err != nil {
		return makePointerFromError(err, responseSize)
	}
	return invokeServiceMethod(service.GetExperiment, new(protos.GetExperiment), requestData, requestSize, responseSize)
}
//export TrackingServiceDeleteExperiment
func TrackingServiceDeleteExperiment(serviceID int64, requestData unsafe.Pointer, requestSize C.int, responseSize *C.int) unsafe.Pointer {
	service, err := trackingServices.Get(serviceID)
	if err != nil {
		return makePointerFromError(err, responseSize)
	}
	return invokeServiceMethod(service.DeleteExperiment, new(protos.DeleteExperiment), requestData, requestSize, responseSize)
}
//export TrackingServiceRestoreExperiment
func TrackingServiceRestoreExperiment(serviceID int64, requestData unsafe.Pointer, requestSize C.int, responseSize *C.int) unsafe.Pointer {
	service, err := trackingServices.Get(serviceID)
	if err != nil {
		return makePointerFromError(err, responseSize)
	}
	return invokeServiceMethod(service.RestoreExperiment, new(protos.RestoreExperiment), requestData, requestSize, responseSize)
}
//export TrackingServiceUpdateExperiment
func TrackingServiceUpdateExperiment(serviceID int64, requestData unsafe.Pointer, requestSize C.int, responseSize *C.int) unsafe.Pointer {
	service, err := trackingServices.Get(serviceID)
	if err != nil {
		return makePointerFromError(err, responseSize)
	}
	return invokeServiceMethod(service.UpdateExperiment, new(protos.UpdateExperiment), requestData, requestSize, responseSize)
}
//export TrackingServiceCreateRun
func TrackingServiceCreateRun(serviceID int64, requestData unsafe.Pointer, requestSize C.int, responseSize *C.int) unsafe.Pointer {
	service, err := trackingServices.Get(serviceID)
	if err != nil {
		return makePointerFromError(err, responseSize)
	}
	return invokeServiceMethod(service.CreateRun, new(protos.CreateRun), requestData, requestSize, responseSize)
}
//export TrackingServiceUpdateRun
func TrackingServiceUpdateRun(serviceID int64, requestData unsafe.Pointer, requestSize C.int, responseSize *C.int) unsafe.Pointer {
	service, err := trackingServices.Get(serviceID)
	if err != nil {
		return makePointerFromError(err, responseSize)
	}
	return invokeServiceMethod(service.UpdateRun, new(protos.UpdateRun), requestData, requestSize, responseSize)
}
//export TrackingServiceDeleteRun
func TrackingServiceDeleteRun(serviceID int64, requestData unsafe.Pointer, requestSize C.int, responseSize *C.int) unsafe.Pointer {
	service, err := trackingServices.Get(serviceID)
	if err != nil {
		return makePointerFromError(err, responseSize)
	}
	return invokeServiceMethod(service.DeleteRun, new(protos.DeleteRun), requestData, requestSize, responseSize)
}
//export TrackingServiceRestoreRun
func TrackingServiceRestoreRun(serviceID int64, requestData unsafe.Pointer, requestSize C.int, responseSize *C.int) unsafe.Pointer {
	service, err := trackingServices.Get(serviceID)
	if err != nil {
		return makePointerFromError(err, responseSize)
	}
	return invokeServiceMethod(service.RestoreRun, new(protos.RestoreRun), requestData, requestSize, responseSize)
}
//export TrackingServiceLogMetric
func TrackingServiceLogMetric(serviceID int64, requestData unsafe.Pointer, requestSize C.int, responseSize *C.int) unsafe.Pointer {
	service, err := trackingServices.Get(serviceID)
	if err != nil {
		return makePointerFromError(err, responseSize)
	}
	return invokeServiceMethod(service.LogMetric, new(protos.LogMetric), requestData, requestSize, responseSize)
}
//export TrackingServiceLogParam
func TrackingServiceLogParam(serviceID int64, requestData unsafe.Pointer, requestSize C.int, responseSize *C.int) unsafe.Pointer {
	service, err := trackingServices.Get(serviceID)
	if err != nil {
		return makePointerFromError(err, responseSize)
	}
	return invokeServiceMethod(service.LogParam, new(protos.LogParam), requestData, requestSize, responseSize)
}
//export TrackingServiceSetExperimentTag
func TrackingServiceSetExperimentTag(serviceID int64, requestData unsafe.Pointer, requestSize C.int, responseSize *C.int) unsafe.Pointer {
	service, err := trackingServices.Get(serviceID)
	if err != nil {
		return makePointerFromError(err, responseSize)
	}
	return invokeServiceMethod(service.SetExperimentTag, new(protos.SetExperimentTag), requestData, requestSize, responseSize)
}
//export TrackingServiceSetTraceTag
func TrackingServiceSetTraceTag(serviceID int64, requestData unsafe.Pointer, requestSize C.int, responseSize *C.int) unsafe.Pointer {
	service, err := trackingServices.Get(serviceID)
	if err != nil {
		return makePointerFromError(err, responseSize)
	}
	return invokeServiceMethod(service.SetTraceTag, new(protos.SetTraceTag), requestData, requestSize, responseSize)
}
//export TrackingServiceDeleteTraceTag
func TrackingServiceDeleteTraceTag(serviceID int64, requestData unsafe.Pointer, requestSize C.int, responseSize *C.int) unsafe.Pointer {
	service, err := trackingServices.Get(serviceID)
	if err != nil {
		return makePointerFromError(err, responseSize)
	}
	return invokeServiceMethod(service.DeleteTraceTag, new(protos.DeleteTraceTag), requestData, requestSize, responseSize)
}
//export TrackingServiceDeleteTag
func TrackingServiceDeleteTag(serviceID int64, requestData unsafe.Pointer, requestSize C.int, responseSize *C.int) unsafe.Pointer {
	service, err := trackingServices.Get(serviceID)
	if err != nil {
		return makePointerFromError(err, responseSize)
	}
	return invokeServiceMethod(service.DeleteTag, new(protos.DeleteTag), requestData, requestSize, responseSize)
}
//export TrackingServiceGetRun
func TrackingServiceGetRun(serviceID int64, requestData unsafe.Pointer, requestSize C.int, responseSize *C.int) unsafe.Pointer {
	service, err := trackingServices.Get(serviceID)
	if err != nil {
		return makePointerFromError(err, responseSize)
	}
	return invokeServiceMethod(service.GetRun, new(protos.GetRun), requestData, requestSize, responseSize)
}
//export TrackingServiceSearchRuns
func TrackingServiceSearchRuns(serviceID int64, requestData unsafe.Pointer, requestSize C.int, responseSize *C.int) unsafe.Pointer {
	service, err := trackingServices.Get(serviceID)
	if err != nil {
		return makePointerFromError(err, responseSize)
	}
	return invokeServiceMethod(service.SearchRuns, new(protos.SearchRuns), requestData, requestSize, responseSize)
}
//export TrackingServiceLogBatch
func TrackingServiceLogBatch(serviceID int64, requestData unsafe.Pointer, requestSize C.int, responseSize *C.int) unsafe.Pointer {
	service, err := trackingServices.Get(serviceID)
	if err != nil {
		return makePointerFromError(err, responseSize)
	}
	return invokeServiceMethod(service.LogBatch, new(protos.LogBatch), requestData, requestSize, responseSize)
}
//export TrackingServiceLogInputs
func TrackingServiceLogInputs(serviceID int64, requestData unsafe.Pointer, requestSize C.int, responseSize *C.int) unsafe.Pointer {
	service, err := trackingServices.Get(serviceID)
	if err != nil {
		return makePointerFromError(err, responseSize)
	}
	return invokeServiceMethod(service.LogInputs, new(protos.LogInputs), requestData, requestSize, responseSize)
}
