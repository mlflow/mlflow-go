package entities

import (
	"strings"

	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/utils"
)

func RunStatusToProto(status string) *protos.RunStatus {
	if status == "" {
		return nil
	}

	if protoStatus, ok := protos.RunStatus_value[strings.ToUpper(status)]; ok {
		return (*protos.RunStatus)(&protoStatus)
	}

	return nil
}

type Run struct {
	Info   *RunInfo
	Data   *RunData
	Inputs *RunInputs
}

func (r Run) ToProto() *protos.Run {
	metrics := make([]*protos.Metric, 0, len(r.Data.Metrics))
	for _, metric := range r.Data.Metrics {
		metrics = append(metrics, metric.ToProto())
	}

	params := make([]*protos.Param, 0, len(r.Data.Params))
	for _, param := range r.Data.Params {
		params = append(params, param.ToProto())
	}

	tags := make([]*protos.RunTag, 0, len(r.Data.Tags))
	for _, tag := range r.Data.Tags {
		tags = append(tags, tag.ToProto())
	}

	data := &protos.RunData{
		Metrics: metrics,
		Params:  params,
		Tags:    tags,
	}

	datasetInputs := make([]*protos.DatasetInput, 0, len(r.Inputs.DatasetInputs))
	for _, input := range r.Inputs.DatasetInputs {
		datasetInputs = append(datasetInputs, input.ToProto())
	}

	inputs := &protos.RunInputs{
		DatasetInputs: datasetInputs,
	}

	return &protos.Run{
		Info: &protos.RunInfo{
			RunId:          &r.Info.RunID,
			RunUuid:        &r.Info.RunID,
			RunName:        &r.Info.RunName,
			ExperimentId:   utils.ConvertInt32PointerToStringPointer(&r.Info.ExperimentID),
			UserId:         &r.Info.UserID,
			Status:         RunStatusToProto(r.Info.Status),
			StartTime:      &r.Info.StartTime,
			EndTime:        r.Info.EndTime,
			ArtifactUri:    &r.Info.ArtifactURI,
			LifecycleStage: utils.PtrTo(r.Info.LifecycleStage),
		},
		Data:   data,
		Inputs: inputs,
	}
}
