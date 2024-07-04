//go:build mage

//nolint:wrapcheck
package main

import (
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
	if err := pipInstall("psycopg2-binary"); err != nil {
		return err
	}

	if err := pipInstall("-e", "."); err != nil {
		return err
	}

	if err := tar(
		"-C", "/usr/local/python/current/lib/python3.8/site-packages/mlflow",
		"-czvf",
		"./ui.tgz",
		"./server/js/build",
	); err != nil {
		return err
	}

	if err := pipInstall("git+https://github.com/jgiannuzzi/mlflow.git@server-signals"); err != nil {
		return err
	}

	if err := tar(
		"-C",
		"/usr/local/python/current/lib/python3.8/site-packages/mlflow",
		"-xzvf",
		"./ui.tgz",
	); err != nil {
		return err
	}

	return nil
}

// Start the mlflow-go dev server connecting to postgres.
func Dev() error {
	mg.Deps(Generate)

	return sh.RunV(
		"mlflow-go",
		"server",
		"--backend-store-uri",
		"postgresql://postgres:postgres@localhost:5432/postgres",
		"--go-opts",
		"log_level=debug,shutdown_timeout=5s",
	)
}
