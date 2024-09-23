package models

import (
	"database/sql"

	"github.com/mlflow/mlflow-go/pkg/entities"
	"github.com/mlflow/mlflow-go/pkg/utils"
)

// Run mapped from table <runs>.
type Run struct {
	ID             string         `db:"run_uuid"         gorm:"column:run_uuid;primaryKey"`
	Name           string         `db:"name"             gorm:"column:name"`
	SourceType     SourceType     `db:"source_type"      gorm:"column:source_type"`
	SourceName     string         `db:"source_name"      gorm:"column:source_name"`
	EntryPointName string         `db:"entry_point_name" gorm:"column:entry_point_name"`
	UserID         string         `db:"user_id"          gorm:"column:user_id"`
	Status         RunStatus      `db:"status"           gorm:"column:status"`
	StartTime      int64          `db:"start_time"       gorm:"column:start_time"`
	EndTime        sql.NullInt64  `db:"end_time"         gorm:"column:end_time"`
	SourceVersion  string         `db:"source_version"   gorm:"column:source_version"`
	LifecycleStage LifecycleStage `db:"lifecycle_stage"  gorm:"column:lifecycle_stage"`
	ArtifactURI    string         `db:"artifact_uri"     gorm:"column:artifact_uri"`
	ExperimentID   int32          `db:"experiment_id"    gorm:"column:experiment_id"`
	DeletedTime    sql.NullInt64  `db:"deleted_time"     gorm:"column:deleted_time"`
	Params         []Param
	Tags           []Tag
	Metrics        []Metric
	LatestMetrics  []LatestMetric
	Inputs         []Input `gorm:"foreignKey:DestinationID"`
}

type RunStatus string

func (s RunStatus) String() string {
	return string(s)
}

const (
	RunStatusRunning   RunStatus = "RUNNING"
	RunStatusScheduled RunStatus = "SCHEDULED"
	RunStatusFinished  RunStatus = "FINISHED"
	RunStatusFailed    RunStatus = "FAILED"
	RunStatusKilled    RunStatus = "KILLED"
)

type SourceType string

const (
	SourceTypeNotebook SourceType = "NOTEBOOK"
	SourceTypeJob      SourceType = "JOB"
	SourceTypeProject  SourceType = "PROJECT"
	SourceTypeLocal    SourceType = "LOCAL"
	SourceTypeUnknown  SourceType = "UNKNOWN"
	SourceTypeRecipe   SourceType = "RECIPE"
)

func (r Run) ToEntity() *entities.Run {
	metrics := make([]*entities.Metric, 0, len(r.LatestMetrics))
	for _, metric := range r.LatestMetrics {
		metrics = append(metrics, metric.ToEntity())
	}

	params := make([]*entities.Param, 0, len(r.Params))
	for _, param := range r.Params {
		params = append(params, param.ToEntity())
	}

	tags := make([]*entities.RunTag, 0, len(r.Tags))
	for _, tag := range r.Tags {
		tags = append(tags, tag.ToEntity())
	}

	datasetInputs := make([]*entities.DatasetInput, 0, len(r.Inputs))
	for _, input := range r.Inputs {
		datasetInputs = append(datasetInputs, input.ToEntity())
	}

	var endTime *int64
	if r.EndTime.Valid {
		endTime = utils.PtrTo(r.EndTime.Int64)
	}

	return &entities.Run{
		Info: &entities.RunInfo{
			RunID:          r.ID,
			RunUUID:        r.ID,
			RunName:        r.Name,
			ExperimentID:   r.ExperimentID,
			UserID:         r.UserID,
			Status:         r.Status.String(),
			StartTime:      r.StartTime,
			EndTime:        endTime,
			ArtifactURI:    r.ArtifactURI,
			LifecycleStage: r.LifecycleStage.String(),
		},
		Data: &entities.RunData{
			Tags:    tags,
			Params:  params,
			Metrics: metrics,
		},
		Inputs: &entities.RunInputs{
			DatasetInputs: datasetInputs,
		},
	}
}
