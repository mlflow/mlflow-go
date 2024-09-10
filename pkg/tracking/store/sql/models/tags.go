package models

import (
	"github.com/mlflow/mlflow-go/pkg/entities"
	"github.com/mlflow/mlflow-go/pkg/protos"
)

// Tag mapped from table <tags>.
type Tag struct {
	Key   string `db:"key"      gorm:"column:key;primaryKey"`
	Value string `db:"value"    gorm:"column:value"`
	RunID string `db:"run_uuid" gorm:"column:run_uuid;primaryKey"`
}

func (t Tag) ToEntity() *entities.RunTag {
	return &entities.RunTag{
		Key:   t.Key,
		Value: t.Value,
	}
}

func (t Tag) ToProto() *protos.RunTag {
	return &protos.RunTag{
		Key:   &t.Key,
		Value: &t.Value,
	}
}

func NewTagFromProto(runID string, proto *protos.RunTag) Tag {
	tag := Tag{
		Key:   *proto.Key,
		Value: *proto.Value,
	}
	if runID != "" {
		tag.RunID = runID
	}

	return tag
}

func NewTagFromEntity(runID string, entity *entities.RunTag) Tag {
	tag := Tag{
		Key:   entity.Key,
		Value: entity.Value,
	}
	if runID != "" {
		tag.RunID = runID
	}

	return tag
}
