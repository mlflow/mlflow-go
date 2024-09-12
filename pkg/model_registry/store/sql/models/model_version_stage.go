package models

import "strings"

type ModelVersionStage string

func (s ModelVersionStage) String() string {
	return string(s)
}

const (
	ModelVersionStageNone       = "None"
	ModelVersionStageStaging    = "Staging"
	ModelVersionStageProduction = "Production"
	ModelVersionStageArchived   = "Archived"
)

var CanonicalMapping = map[string]string{
	strings.ToLower(ModelVersionStageNone):       ModelVersionStageNone,
	strings.ToLower(ModelVersionStageStaging):    ModelVersionStageStaging,
	strings.ToLower(ModelVersionStageProduction): ModelVersionStageProduction,
	strings.ToLower(ModelVersionStageArchived):   ModelVersionStageArchived,
}

func AllModelVersionStages() string {
	pairs := make([]string, 0, len(CanonicalMapping))

	for _, v := range CanonicalMapping {
		pairs = append(pairs, v)
	}

	return strings.Join(pairs, ",")
}
