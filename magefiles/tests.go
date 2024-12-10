//go:build mage

//nolint:wrapcheck
package main

import (
	"os"
	"runtime"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Test mg.Namespace

func runPythonTests(pytestArgs []string) error {
	libpath, err := os.MkdirTemp("", "")
	if err != nil {
		return err
	}

	// Remove the Go binary
	defer os.RemoveAll(libpath)

	// Build the Go binary in a temporary directory
	if err := sh.RunV("uv", "run", "-m", "mlflow_go.lib", "--", ".", libpath); err != nil {
		return nil
	}

	executable := "uv"
	args := []string{"run", "pytest"}

	// For some reason uv run on Mac in GitHub Actions can return exit code 1,
	// even when all the tests are passing. That is why we want to run pytest from the virtual directory.
	if runtime.GOOS == "darwin" {
		executable = ".venv/bin/pytest"
		args = []string{}
	}

	fixedPytestArgs := []string{
		// "-s",
		"--log-cli-level=DEBUG",
		"--confcutdir=.",
		"-k", "not [file",
		"-p", "no:warnings",
	}

	args = append(args, fixedPytestArgs...)
	args = append(args, pytestArgs...)

	environmentVariables := map[string]string{
		"MLFLOW_GO_LIBRARY_PATH": libpath,
		"PYTHONLOGGING":          "DEBUG",
	}

	if runtime.GOOS == "windows" {
		environmentVariables["MLFLOW_SQLALCHEMYSTORE_POOLCLASS"] = "NullPool"
	}

	//  Run the tests (currently just the server ones)
	if err := sh.RunWithV(environmentVariables,
		executable, args...,
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
	})
}

// Run specific Python test against the Go backend.
func (Test) PythonSpecific(testName string) error {
	return runPythonTests([]string{
		testName,
		"-vv",
	})
}

// Run the Go unit tests.
func (Test) Unit() error {
	return sh.RunV("go", "test", "./pkg/...")
}

// Run all tests.
func (Test) All() {
	mg.Deps(Test.Unit, Test.Python)
}
