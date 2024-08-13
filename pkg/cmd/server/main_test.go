package main

import (
	"context"
	"testing"

	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/sql"
	"github.com/mlflow/mlflow-go/pkg/utils"
)

type ExperimentModelWithPointers struct {
	ID               *int32  `gorm:"column:experiment_id;primaryKey;autoIncrement:true"`
	Name             *string `gorm:"column:name;not null"`
	ArtifactLocation *string `gorm:"column:artifact_location"`
}

type ExperimentModelWithoutPointers struct {
	ID               int32  `gorm:"column:experiment_id;primaryKey;autoIncrement:true"`
	Name             string `gorm:"column:name;not null"`
	ArtifactLocation string `gorm:"column:artifact_location"`
}

func BenchmarkPointerImplementation(b *testing.B) {
	req := protos.CreateExperiment{
		Name:             utils.PtrTo("name"),
		ArtifactLocation: utils.PtrTo("location"),
	}
	database, _ := sql.NewDatabase(context.Background(), "sqlite:///value.db")
	database.AutoMigrate(&ExperimentModelWithPointers{})

	for i := 0; i < b.N; i++ {
		model := ExperimentModelWithPointers{
			ID:               utils.PtrTo(int32(i)),
			Name:             req.Name,
			ArtifactLocation: req.ArtifactLocation,
		}
		database.Save(&model)
		database.Where("experiment_id = ?", model.ID).First(&ExperimentModelWithPointers{})
	}
}

func BenchmarkNonPointerImplementation(b *testing.B) {
	req := protos.CreateExperiment{
		Name:             utils.PtrTo("name"),
		ArtifactLocation: utils.PtrTo("location"),
	}
	database, _ := sql.NewDatabase(context.Background(), "sqlite:///pointer.db")
	database.AutoMigrate(&ExperimentModelWithoutPointers{})

	for i := 0; i < b.N; i++ {
		model := ExperimentModelWithoutPointers{
			ID:               int32(i),
			Name:             req.GetName(),
			ArtifactLocation: req.GetArtifactLocation(),
		}
		database.Save(&model)
		database.Where("experiment_id = ?", model.ID).First(&ExperimentModelWithoutPointers{})
	}
}
