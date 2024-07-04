//go:build mage

//nolint:wrapcheck
package main

import (
	"os"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Test mg.Namespace

// Run mlflow Python tests against the Go backend.
func (Test) Python() error {
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
		"-k",
		"not [file",
	); err != nil {
		return err
	}

	return nil
}

// Run the Go unit tests.
func (Test) Unit() error {
	return sh.RunV("go", "test", "./pkg/...")
}

// Run all tests.
func (Test) All() {
	mg.Deps(Test.Unit, Test.Python)
}
