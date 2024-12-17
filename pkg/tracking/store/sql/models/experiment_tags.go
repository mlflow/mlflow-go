package models

const TableNameExperimentTag = "experiment_tags"

// ExperimentTag mapped from table <experiment_tags>.
type ExperimentTag struct {
	Key          string `gorm:"column:key;primaryKey"`
	Value        string `gorm:"column:value"`
	ExperimentID int32  `gorm:"column:experiment_id;primaryKey"`
}
