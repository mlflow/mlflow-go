package generate

var validations = map[string]string{
	"GetExperiment_ExperimentId":         "required,stringAsPositiveInteger",
	"CreateExperiment_Name":              "required,max=500",
	"CreateExperiment_ArtifactLocation":  "omitempty,uriWithoutFragmentsOrParamsOrDotDotInQuery",
	"SearchRuns_RunViewType":             "omitempty",
	"SearchRuns_MaxResults":              "gt=0,max=50000",
	"DeleteExperiment_ExperimentId":      "required,stringAsPositiveInteger",
	"LogParam_Key":                       "required,max=250,validMetricParamOrTagName,pathIsUnique",
	"LogParam_Value":                     "omitempty,truncate=6000",
	"LogBatch_RunId":                     "required,runId",
	"LogBatch_Params":                    "omitempty,uniqueParams,max=100,dive",
	"LogBatch_Metrics":                   "max=1000,dive",
	"LogBatch_Tags":                      "max=100",
	"RunTag_Key":                         "required,max=250,validMetricParamOrTagName,pathIsUnique",
	"RunTag_Value":                       "omitempty,max=5000",
	"Param_Key":                          "required,max=250,validMetricParamOrTagName,pathIsUnique",
	"Param_Value":                        "omitempty,truncate=6000",
	"Metric_Key":                         "required,max=250,validMetricParamOrTagName,pathIsUnique",
	"Metric_Timestamp":                   "required",
	"Metric_Value":                       "required",
	"CreateRun_ExperimentId":             "required,stringAsPositiveInteger",
	"GetExperimentByName_ExperimentName": "required",
	"GetLatestVersions_Name":             "required",
	"LogMetric_RunId":                    "required",
	"LogMetric_Key":                      "required",
	"LogMetric_Value":                    "required",
	"LogMetric_Timestamp":                "required",
	"DeleteTag_RunId":                    "required",
	"DeleteTag_Key":                      "required",
	"SetExperimentTag_ExperimentId":      "required",
	"SetExperimentTag_Key":               "required,max=250,validMetricParamOrTagName",
	"SetExperimentTag_Value":             "max=5000",
	"SearchExperiments_MaxResults":       "positiveNonZeroInteger,max=50000",
}
