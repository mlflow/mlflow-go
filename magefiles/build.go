//go:build mage

//nolint:wrapcheck
package main

import (
	"errors"
	"fmt"
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
	case "darwin":
		if goarch == amd64 {
			return "x86_64-macos", nil
		} else if goarch == "arm64" {
			return "aarch64-macos", nil
		}
	case "windows":
		if goarch == amd64 {
			return "x86_64-windows-gnu", nil
		}
	}

	return "", fmt.Errorf("%w: %s/%s", errUnknownTarget, goos, goarch)
}

// Build a Python wheel.
func Build() error {
	if err := sh.RunV("python3", "-mvenv", "env"); err != nil {
		return err
	}

	if err := sh.RunV("./env/bin/pip", "install", "build", "ziglang"); err != nil {
		return err
	}

	// Set Zig as the C compiler for cross-compilation
	targetTriple, err := getTargetTriple(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
	}

	absPython, err := filepath.Abs("./env/bin/python")
	if err != nil {
		return err
	}

	zigCC := absPython + " -mziglang cc -target " + targetTriple

	if err := sh.RunWithV(
		map[string]string{
			"CC":     zigCC,
			"GOOS":   runtime.GOOS,
			"GOARCH": runtime.GOARCH,
		},
		"./env/bin/python", "-mbuild"); err != nil {
		return err
	}

	return nil
}
