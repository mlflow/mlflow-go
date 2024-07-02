//go:build mage

// mlflow-go build scripts
package main

import (
	"github.com/magefile/mage/mg"
)

func Clean() error {
	println("imagine cleaning up some folder")
	return nil
}

// Build the Go source code. This does not produce any binary.
func Build() error {
	mg.Deps(Clean)

	println("building")
	return nil
}
