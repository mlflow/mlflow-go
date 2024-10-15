//go:build mage

//nolint:wrapcheck
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/magefile/mage/sh"
)

const (
	MacOS   = "darwin"
	Linux   = "linux"
	Windows = "windows"
	Amd64   = "amd64"
	Arm64   = "arm64"
)

func IsWindows() bool {
	return runtime.GOOS == Windows
}

func IsNotMacOS() bool {
	return runtime.GOOS != MacOS
}

var errUnknownTarget = errors.New("could not determine zig target")

// Helper function to determine the Zig target triple based on OS and architecture.
func getTargetTriple(goos, goarch string) (string, error) {
	switch goos {
	case Linux:
		if goarch == Amd64 {
			return "x86_64-linux-gnu", nil
		} else if goarch == Arm64 {
			return "aarch64-linux-gnu", nil
		}
	case Windows:
		if goarch == Amd64 {
			return "x86_64-windows-gnu", nil
		} else if goarch == Arm64 {
			return "aarch64-windows-gnu", nil
		}
	}

	return "", fmt.Errorf("%w: %s/%s", errUnknownTarget, goos, goarch)
}

func getCC(pythonExecutable, goos, goarch string) (string, error) {
	target, err := getTargetTriple(goos, goarch)
	if err != nil {
		return "", err
	}

	return pythonExecutable + " -mziglang cc -target " + target, nil
}

var errUnsupportedDarwin = errors.New(`it is unsupported to build a Python wheel on Mac on a non-Mac platform`)

// Build a Python wheel.
func Build(goos, goarch string) error {
	tmp, err := os.MkdirTemp("", "")
	if err != nil {
		return err
	}

	defer os.RemoveAll(tmp)

	venv, err := filepath.Abs(filepath.Join(tmp, ".venv"))
	if err != nil {
		return err
	}

	if err := sh.RunV("uv", "venv", "--python", "3.8", venv); err != nil {
		return err
	}

	if err := sh.RunWithV(map[string]string{
		"UV_PROJECT_ENVIRONMENT": venv,
	}, "uv", "sync", "--all-extras"); err != nil {
		return err
	}

	binDir := "bin"
	if IsWindows() {
		binDir = "Scripts"
	}

	python := filepath.Join(venv, binDir, "python")

	environmentVariables := map[string]string{
		"UV_PROJECT_ENVIRONMENT": venv,
		"GOOS":                   goos,
		"GOARCH":                 goarch,
	}

	// Set Zig as the C compiler for cross-compilation
	// If we are on Mac and targeting Mac we don't need Zig.
	if goos == MacOS {
		if IsNotMacOS() {
			return errUnsupportedDarwin
		}
	} else {
		zigCC, err := getCC(python, goos, goarch)
		if err != nil {
			return err
		}

		environmentVariables["CC"] = zigCC
	}

	if err := sh.RunWithV(
		environmentVariables,
		"uv", "run", "python", "-mbuild"); err != nil {
		return err
	}

	return nil
}
