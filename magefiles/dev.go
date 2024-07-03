//go:build mage

package main

import (
	"github.com/magefile/mage/sh"
)

func pip_install(args ...string) error {
	allArgs := append([]string{"install"}, args...)

	return sh.RunV("pip", allArgs...)
}

func tar(args ...string) error {
	return sh.RunV("tar", args...)
}

func Temp() error {
	if err := pip_install("psycopg2-binary"); err != nil {
		return err
	}

	if err := pip_install("-e", "."); err != nil {
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

	if err := pip_install("git+https://github.com/jgiannuzzi/mlflow.git@server-signals"); err != nil {
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
	return sh.RunV(
		"mlflow-go",
		"server",
		"--backend-store-uri",
		"postgresql://postgres:postgres@localhost:5432/postgres",
		"--go-opts",
		"log_level=debug,shutdown_timeout=5s",
	)
}
