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

const Windows = "windows"

func IsWindows() bool {
	return runtime.GOOS == Windows
}

func IsNotMac() bool {
	return runtime.GOOS != "darwin"
}

var errUnknownTarget = errors.New("could not determine zig target")

// Helper function to determine the Zig target triple based on OS and architecture.
func getTargetTriple(goos, goarch string) (string, error) {
	switch goos {
	case "linux":
		if goarch == "amd64" {
			return "x86_64-linux-gnu", nil
		} else if goarch == "arm64" {
			return "aarch64-linux-gnu", nil
		}
	case Windows:
		if goarch == "amd64" {
			return "x86_64-windows-gnu", nil
		} else if goarch == "arm64" {
			return "aarch64-windows-gnu", nil
		}
	}

	return "", fmt.Errorf("%w: %s/%s", errUnknownTarget, goos, goarch)
}

func getCC(venv, pythonExecutable, goos, goarch string) (string, error) {
	if err := sh.RunWithV(map[string]string{
		"VIRTUAL_ENV": venv,
	}, "uv", "pip", "install", "build", "ziglang"); err != nil {
		return "", err
	}

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

	if err := sh.RunV("uv", "venv", venv); err != nil {
		return err
	}

	binDir := "bin"
	if IsWindows() {
		binDir = "Scripts"
	}

	python := filepath.Join(venv, binDir, "python")

	environmentVariables := map[string]string{
		"GOOS":   goos,
		"GOARCH": goarch,
	}

	// Set Zig as the C compiler for cross-compilation
	// If we are on Mac and targeting Mac we don't need Zig.
	if goos == "darwin" {
		if IsNotMac() {
			return errUnsupportedDarwin
		}
	} else {
		zigCC, err := getCC(venv, python, goos, goarch)
		if err != nil {
			return err
		}

		environmentVariables["CC"] = zigCC
	}

	if err := sh.RunWithV(
		environmentVariables,
		python, "-mbuild"); err != nil {
		return err
	}

	return nil
}
