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
	amd64 = "amd64"
)

var errUnknownTarget = errors.New("could not determine zig target")

// Helper function to determine the Zig target triple based on OS and architecture.
func getTargetTriple(goos, goarch string) (string, error) {
	switch goos {
	case "linux":
		if goarch == amd64 {
			return "x86_64-linux-gnu", nil
		} else if goarch == "arm64" {
			return "aarch64-linux-gnu", nil
		}
	case "windows":
		if goarch == amd64 {
			return "x86_64-windows-gnu", nil
		} else if goarch == "arm64" {
			return "aarch64-windows-gnu", nil
		}
	}

	return "", fmt.Errorf("%w: %s/%s", errUnknownTarget, goos, goarch)
}

// TODO: someday add mussle

// Build a Python wheel.
func Build(goos, goarch string) error {
	// env := os.mkTemp
	// defer os.RemoveAll(tmp)

	env, err := filepath.Abs("build/.venv")
	if err != nil {
		return err
	}

	if err := sh.RunV("python3", "-mvenv", env); err != nil {
		return err
	}

	defer os.RemoveAll(env)

	pip := filepath.Join(env, "bin", "pip")
	python := filepath.Join(env, "bin", "python")

	if err := sh.RunV(pip, "install", "build", "ziglang"); err != nil {
		return err
	}

	environmentVariables := map[string]string{
		"GOOS":   goos,   // runtime.GOOS,
		"GOARCH": goarch, // runtime.GOARCH,
	}

	// Set Zig as the C compiler for cross-compilation
	// If we are on Mac and targeting Mac we don't need Zig.
	if goos == "darwin" {
		if runtime.GOOS != "darwin" {
			return errors.New("unsupported")
		}
	} else {
		target, err := getTargetTriple(goos, goarch)
		if err != nil {
			return err
		}

		zigCC := python + " -mziglang cc -target " + target
		environmentVariables["CC"] = zigCC
	}
	// runtime.GOOS == "darwin" &&

	// strings.HasSuffix(target, "-macos")) {

	if err := sh.RunWithV(
		environmentVariables,
		python, "-mbuild"); err != nil {
		return err
	}

	return nil
}
