//go:build mage

//nolint:wrapcheck
package main

import (
	"os"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Test mg.Namespace

func cleanUpMemoryFile() error {
	// Clean up :memory: file
	filename := ":memory:"
	_, err := os.Stat(filename)

	if err == nil {
		// File exists, delete it
		err = os.Remove(filename)
		if err != nil {
			return err
		}
	}

	return nil
}

func runPythonTests(testFiles []string, testName string) error {
	libpath, err := os.MkdirTemp("", "")
	if err != nil {
		return err
	}

	// Remove the Go binary
	defer os.RemoveAll(libpath)
	//nolint:errcheck
	defer cleanUpMemoryFile()

	// Build the Go binary in a temporary directory
	if err := sh.RunV("python", "-m", "mlflow_go.lib", ".", libpath); err != nil {
		return nil
	}

	args := []string{
		"--confcutdir=.",
	}
	args = append(args, testFiles...)

	// Add testName filter if provided
	if testName != "" {
		args = append(args, "-k", testName)
	} else {
		args = append(args, "-k", "not [file")
	}

	//  Run the tests (currently just the server ones)
	if err := sh.RunWithV(map[string]string{
		"MLFLOW_GO_LIBRARY_PATH": libpath,
	}, "pytest", args...,
		// "-vv",
	); err != nil {
		return err
	}

	return nil
}

// Run mlflow Python tests against the Go backend.
func (Test) Python() error {
	return runPythonTests([]string{
		".mlflow.repo/tests/tracking/test_rest_tracking.py",
		".mlflow.repo/tests/tracking/test_model_registry.py",
		".mlflow.repo/tests/store/tracking/test_sqlalchemy_store.py",
		".mlflow.repo/tests/store/model_registry/test_sqlalchemy_store.py",
	}, "")
}

// Run specific Python test against the Go backend.
func (Test) PythonSpecific(testName string) error {
	return runPythonTests([]string{
		".mlflow.repo/tests/tracking/test_rest_tracking.py",
	}, testName)
}

// Run the Go unit tests.
func (Test) Unit() error {
	return sh.RunV("go", "test", "./pkg/...")
}

// Run all tests.
func (Test) All() {
	mg.Deps(Test.Unit, Test.Python)
}
