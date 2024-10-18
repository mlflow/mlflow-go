//go:build mage

//nolint:wrapcheck
package main

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Test mg.Namespace

// Run mlflow Python tests against the Go backend.
func (Test) Python() error {
	libpath, err := os.MkdirTemp("", "")
	if err != nil {
		return err
	}

	// Remove the Go binary
	defer os.RemoveAll(libpath)

	venv, err := filepath.Abs(".venv")
	if err != nil {
		return err
	}

	python := filepath.Join(venv, "bin", "python")
	if IsWindows() {
		python = filepath.Join(venv, "Scripts", "python")
	}

	buildEnv := make(map[string]string)

	if IsNotMac() {
		cc, err := getCC(python, runtime.GOOS, runtime.GOARCH)
		if err != nil {
			return err
		}

		buildEnv["CC"] = cc
	}

	// Build the Go binary in a temporary directory
	if err := sh.RunWithV(buildEnv, python, "-m", "mlflow_go.lib", ".", libpath); err != nil {
		return nil
	}

	pytest := filepath.Join(venv, "bin", "pytest")
	if IsWindows() {
		pytest = filepath.Join(venv, "Scripts", "pytest")
	}

	//  Run the tests (currently just the server ones)
	if err := sh.RunWithV(map[string]string{
		"MLFLOW_GO_LIBRARY_PATH": libpath,
	}, pytest,
		"--confcutdir=.",
		".mlflow.repo/tests/tracking/test_rest_tracking.py",
		".mlflow.repo/tests/tracking/test_model_registry.py",
		".mlflow.repo/tests/store/tracking/test_sqlalchemy_store.py",
		".mlflow.repo/tests/store/model_registry/test_sqlalchemy_store.py",
		"-k",
		"not [file",
		// "-vv",
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
