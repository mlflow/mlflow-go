package entities

import "github.com/mlflow/mlflow-go/pkg/protos"

type Dataset struct {
	Name       string
	Digest     string
	SourceType string
	Source     string
	Schema     string
	Profile    string
}

func (d *Dataset) ToProto() *protos.Dataset {
	return &protos.Dataset{
		Name:       &d.Name,
		Digest:     &d.Digest,
		SourceType: &d.SourceType,
		Source:     &d.Source,
		Schema:     &d.Schema,
		Profile:    &d.Profile,
	}
}
