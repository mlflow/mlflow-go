package generate

import (
	"fmt"

	"github.com/mlflow/mlflow-go/pkg/protos"
)

var validations = map[string]string{
	"GetExperiment_ExperimentId":         "required,stringAsPositiveInteger",
	"CreateExperiment_Name":              "required",
	"CreateExperiment_ArtifactLocation":  "omitempty,uriWithoutFragmentsOrParamsOrDotDotInQuery",
	"SearchRuns_RunViewType":             "omitempty",
	"SearchRuns_MaxResults":              fmt.Sprintf("lte=%d", protos.Default_SearchRuns_MaxResults),
	"DeleteExperiment_ExperimentId":      "required,stringAsPositiveInteger",
	"LogBatch_RunId":                     "required,runId",
	"LogBatch_Params":                    "omitempty,uniqueParams,max=100",
	"LogBatch_Metrics":                   "max=1000,dip",
	"LogBatch_Tags":                      "max=100",
	"RunTag_Key":                         "required,max=250,validMetricParamOrTagName,pathIsUnique",
	"RunTag_Value":                       "omitempty,max=5000",
	"Param_Key":                          "required,max=250,validMetricParamOrTagName,pathIsUnique",
	"Param_Value":                        "omitempty,max=6000",
	"Metric_Key":                         "required,max=250,validMetricParamOrTagName,pathIsUnique",
	"Metric_Timestamp":                   "required",
	"Metric_Value":                       "required",
	"CreateRun_ExperimentId":             "required,stringAsPositiveInteger",
	"GetExperimentByName_ExperimentName": "required",
	"GetLatestVersions_Name":             "required",
}
