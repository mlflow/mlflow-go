package generate

type ServiceGenerationInfo struct {
	FileNameWithoutExtension string
	ServiceName              string
	ImplementedEndpoints     []string
}

var ServiceInfoMap = map[string]ServiceGenerationInfo{
	"MlflowService": {
		FileNameWithoutExtension: "tracking",
		ServiceName:              "TrackingService",
		ImplementedEndpoints: []string{
			"getExperimentByName",
			"createExperiment",
			"searchExperiments",
			"getExperiment",
			"deleteExperiment",
			"restoreExperiment",
			"updateExperiment",
			"getRun",
			"createRun",
			"updateRun",
			"deleteRun",
			"restoreRun",
			"logMetric",
			"logParam",
			"setExperimentTag",
			"setTag",
			"setTraceTag",
			"deleteTraceTag",
			"deleteTag",
			"searchRuns",
			// "listArtifacts",
			// "getMetricHistory",
			// "getMetricHistoryBulkInterval",
			"logBatch",
			// "logModel",
			"logInputs",
			// "startTrace",
			"endTrace",
			// "getTraceInfo",
			// "searchTraces",
			// "deleteTraces",
		},
	},
	"ModelRegistryService": {
		FileNameWithoutExtension: "model_registry",
		ServiceName:              "ModelRegistryService",
		ImplementedEndpoints: []string{
			// "createRegisteredModel",
			// "renameRegisteredModel",
			// "updateRegisteredModel",
			// "deleteRegisteredModel",
			// "getRegisteredModel",
			// "searchRegisteredModels",
			"getLatestVersions",
			// "createModelVersion",
			// "updateModelVersion",
			// "transitionModelVersionStage",
			// "deleteModelVersion",
			// "getModelVersion",
			// "searchModelVersions",
			// "getModelVersionDownloadUri",
			// "setRegisteredModelTag",
			// "setModelVersionTag",
			// "deleteRegisteredModelTag",
			// "deleteModelVersionTag",
			// "setRegisteredModelAlias",
			// "deleteRegisteredModelAlias",
			// "getModelVersionByAlias",
		},
	},
	"MlflowArtifactsService": {
		FileNameWithoutExtension: "artifacts",
		ServiceName:              "ArtifactsService",
		ImplementedEndpoints:     []string{
			// "downloadArtifact",
			// "uploadArtifact",
			// "listArtifacts",
			// "deleteArtifact",
			// "createMultipartUpload",
			// "completeMultipartUpload",
			// "abortMultipartUpload",
		},
	},
}
