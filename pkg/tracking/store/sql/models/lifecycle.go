package models

type LifecycleStage string

func (s LifecycleStage) String() string {
	return string(s)
}

const (
	LifecycleStageActive  LifecycleStage = "active"
	LifecycleStageDeleted LifecycleStage = "deleted"
)
