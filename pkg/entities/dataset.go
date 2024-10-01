package entities

import (
	"github.com/mlflow/mlflow-go/pkg/protos"
)

type Dataset struct {
	Name       string
	Digest     string
	SourceType string
	Source     string
	Schema     string
	Profile    string
}

func (d *Dataset) ToProto() *protos.Dataset {
	var schema *string
	if d.Schema != "" {
		schema = &d.Schema
	}

	var profile *string
	if d.Profile != "" {
		profile = &d.Profile
	}

	return &protos.Dataset{
		Name:       &d.Name,
		Digest:     &d.Digest,
		SourceType: &d.SourceType,
		Source:     &d.Source,
		Schema:     schema,
		Profile:    profile,
	}
}
