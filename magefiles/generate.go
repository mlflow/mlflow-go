//go:build mage

package main

import (
	"fmt"
	"os"
	"path"

	"github.com/mlflow/mlflow-go/magefiles/generate"
)

// Generate Go files based on proto files and other configuration.
func Generate() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	pkgFolder := path.Join(pwd, "pkg")

	if err := generate.RunProtoc(""); err != nil {
		return err
	}

	if err := generate.AddQueryAnnotations(pkgFolder); err != nil {
		return err
	}

	if err := generate.SourceCode(pkgFolder); err != nil {
		return err
	}

	fmt.Println("Successfully added query annotations and generated services!")

	return nil
}
