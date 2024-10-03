//go:build mage

//nolint:wrapcheck
package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/magefile/mage/sh"
)

const (
	amd64 = "amd64"
)

// Build a Python wheel.
func Build(target string) error {
	env := "build/.venv"
	if err := sh.RunV("python3", "-mvenv", env); err != nil {
		return err
	}

	defer os.RemoveAll(env)

	pip := env + "/bin/pip"
	python := env + "/bin/python"

	if err := sh.RunV(pip, "install", "build", "ziglang"); err != nil {
		return err
	}

	absPython, err := filepath.Abs(python)
	if err != nil {
		return err
	}

	environmentVariables := map[string]string{
		"GOOS":   runtime.GOOS,
		"GOARCH": runtime.GOARCH,
	}

	// Set Zig as the C compiler for cross-compilation
	// If we are on Mac and targeting Mac we don't need Zig.
	if !(runtime.GOOS == "darwin" && strings.HasSuffix(target, "-macos")) {
		zigCC := absPython + " -mziglang cc -target " + target
		environmentVariables["CC"] = zigCC
	}

	if err := sh.RunWithV(
		environmentVariables,
		python, "-mbuild"); err != nil {
		return err
	}

	return nil
}
