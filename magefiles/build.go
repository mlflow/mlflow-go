//go:build mage

//nolint:wrapcheck
package main

import (
	"github.com/magefile/mage/sh"
)

// Build a Python wheel.
func Build(goos, goarch string) error {
	if err := sh.RunWithV(map[string]string{
		"TARGET_GOOS":   goos,
		"TARGET_GOARCH": goarch,
	}, "uvx", "--from", "build[uv]", "pyproject-build", "--installer", "uv"); err != nil {
		return err
	}

	return nil
}
