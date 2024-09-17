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
