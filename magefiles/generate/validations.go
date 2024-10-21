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
	"SetTraceTag_Key":                    "required,max=250,validMetricParamOrTagName,pathIsUnique",
	"SetTraceTag_Value":                  "omitempty,truncate=8000",
	"DeleteTag_RunId":                    "required",
	"DeleteTag_Key":                      "required",
	"SetExperimentTag_ExperimentId":      "required",
	"SetExperimentTag_Key":               "required,max=250,validMetricParamOrTagName",
	"SetExperimentTag_Value":             "max=5000",
	"SearchExperiments_MaxResults":       "positiveNonZeroInteger,max=50000",
	"SetTag_Key":                         "required,max=1000,validMetricParamOrTagName,pathIsUnique",
	"SetTag_Value":                       "omitempty,truncate=8000",
	"LogInputs_RunId":                    "required,runId",
	"LogInputs_Datasets":                 "required",
	"DatasetInput_Dataset":               "required",
	"Dataset_Name":                       "required,max=500",
	"Dataset_Digest":                     "required,max=36",
	"Dataset_SourceType":                 "required",
	"Dataset_Source":                     "required,max=65535",
	"Dataset_Profile":                    "max:16777215",
	"Dataset_Schema":                     "max:1048575",
	"InputTag_Key":                       "required,max=255",
	"InputTag_Value":                     "required,max=500",
}
