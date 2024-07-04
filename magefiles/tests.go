//go:build mage

//nolint:wrapcheck
package main

import (
	"os"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Run mlflow Python tests against the Go backend.
func Tests() error {
	mg.Deps(Generate)

	libpath, err := os.MkdirTemp("", "")
	if err != nil {
		return err
	}

	// Remove the Go binary
	defer os.RemoveAll(libpath)

	// Build the Go binary in a temporary directory
	if err := sh.RunV("python", "-m", "mlflow_go.lib", ".", libpath); err != nil {
		return nil
	}

	//  Run the tests (currently just the server ones)
	if err := sh.RunWithV(map[string]string{
		"MLFLOW_GO_LIBRARY_PATH": libpath,
	}, "pytest",
		"--confcutdir=.",
		".mlflow.repo/tests/tracking/test_rest_tracking.py",
		".mlflow.repo/tests/tracking/test_model_registry.py",
	); err != nil {
		return err
	}

	return nil
}
