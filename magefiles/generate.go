//go:build mage

//nolint:wrapcheck
package main

import (
	"path"
	"path/filepath"

	"github.com/gofiber/fiber/v2/log"
	"github.com/magefile/mage/mg"

	"github.com/mlflow/mlflow-go/magefiles/generate"
)

// Generate Go files based on proto files and other configuration.
func Generate() error {
	mg.Deps(Repo.Init)

	protoFolder, err := filepath.Abs(path.Join(MLFlowRepoFolderName, "mlflow", "protos"))
	if err != nil {
		return err
	}

	if err := generate.RunProtoc(protoFolder); err != nil {
		return err
	}

	pkgFolder, err := filepath.Abs("pkg")
	if err != nil {
		return err
	}

	if err := generate.AddQueryAnnotations(pkgFolder); err != nil {
		return err
	}

	if err := generate.SourceCode(pkgFolder); err != nil {
		return err
	}

	log.Info("Successfully added query annotations and generated services!")

	return nil
}
