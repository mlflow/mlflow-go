//go:build mage

//nolint:wrapcheck
package main

import "github.com/magefile/mage/sh"

// Build a Python wheel.
func Build() error {
	if err := sh.RunV("python3", "-mvenv", "foo"); err != nil {
		return err
	}

	if err := sh.RunV("sh", "-c", "'source ./env/bin/activate'"); err != nil {
		return err
	}

	if err := sh.RunV("pip", "install", "build", "ziglang"); err != nil {
		return err
	}

	if err := sh.RunV("build"); err != nil {
		return err
	}

	return nil
}
