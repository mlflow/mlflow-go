package entities

type RunInfo struct {
	RunID          string
	RunUUID        string
	RunName        string
	ExperimentID   int32
	UserID         string
	Status         string
	StartTime      int64
	EndTime        int64
	ArtifactURI    string
	LifecycleStage string
}
