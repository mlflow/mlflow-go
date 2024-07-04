//go:build mage

//nolint:wrapcheck
package main

import (
	"os"
	"path"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

func pipInstall(args ...string) error {
	allArgs := append([]string{"install"}, args...)

	return sh.RunV("pip", allArgs...)
}

func tar(args ...string) error {
	return sh.RunV("tar", args...)
}

func Temp() error {
	mg.Deps(Init)

	// Install our Python package and its dependencies
	if err := pipInstall("-e", "."); err != nil {
		return err
	}

	// Install the dreaded psycho
	if err := pipInstall("psycopg2-binary"); err != nil {
		return err
	}

	// Archive the MLFlow pre-built UI
	if err := tar(
		"-C", "/usr/local/python/current/lib/python3.8/site-packages/mlflow",
		"-czvf",
		"./ui.tgz",
		"./server/js/build",
	); err != nil {
		return err
	}

	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	mlflowRepoPath := path.Join(pwd, MLFlowRepoFolderName)

	// Add the UI back to it
	if err := tar(
		"-C", mlflowRepoPath,
		"-xzvf", "./ui.tgz",
	); err != nil {
		return err
	}

	// Remove tar file
	defer os.RemoveAll(path.Join(pwd, "ui.tgz"))

	// Install it in editable mode
	if err := pipInstall("-e", mlflowRepoPath); err != nil {
		return err
	}

	return nil
}
