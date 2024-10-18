package entities

import "github.com/mlflow/mlflow-go/pkg/protos"

type DatasetInput struct {
	Tags    []*InputTag
	Dataset *Dataset
}

func (ds DatasetInput) ToProto() *protos.DatasetInput {
	tags := make([]*protos.InputTag, 0, len(ds.Tags))
	for _, tag := range ds.Tags {
		tags = append(tags, tag.ToProto())
	}

	return &protos.DatasetInput{
		Tags:    tags,
		Dataset: ds.Dataset.ToProto(),
	}
}

func NewDatasetInputFromProto(proto *protos.DatasetInput) *DatasetInput {
	tags := make([]*InputTag, 0, len(proto.GetTags()))
	for _, t := range proto.GetTags() {
		tags = append(tags, NewInputTagFromProto(t))
	}

	return &DatasetInput{
		Dataset: NewDatasetFromProto(proto.GetDataset()),
		Tags:    tags,
	}
}
