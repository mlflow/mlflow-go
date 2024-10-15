//go:build mage

//nolint:wrapcheck
package main

import (
	"io"
	"os"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// copyFile copies a file from src to dst, preserving file permissions.
func copyFile(srcFile, dstFile string) error {
	// Open the source file
	src, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer src.Close()

	// Create the destination file (create directory if needed)
	dst, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	// Set the same file permissions
	info, err := os.Stat(srcFile)
	if err != nil {
		return err
	}

	return os.Chmod(dstFile, info.Mode())
}

// CopyDirectory recursively copies a source directory to a destination directory.
func copyDirectory(srcDir, destDir string) error {
	// Walk through the source directory
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create the destination path by replacing the source base path with the destination path
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(destDir, relPath)

		// Check if it's a directory
		if info.IsDir() {
			// Create the directory in the destination if it doesn't exist
			if err := os.MkdirAll(destPath, info.Mode()); err != nil {
				return err
			}

			return nil
		}

		// Otherwise, copy the file
		return copyFile(path, destPath)
	})
}

// Configure development environment.
func Configure() error {
	mg.Deps(Repo.Init)

	// create virtual environment using uv
	if err := sh.Run("uv", "venv"); err != nil {
		return err
	}

	// Install our Python package and its dependencies
	if err := sh.RunV("uv", "pip", "install", "-e", "."); err != nil {
		return err
	}

	mlflowDir, err := filepath.Abs(filepath.Join(".venv", "lib", "python3.8", "site-packages", "mlflow"))
	if err != nil {
		return err
	}

	jsBuild := filepath.Join(mlflowDir, "server", "js", "build")

	mlflowRepoPath, err := filepath.Abs(MLFlowRepoFolderName)
	if err != nil {
		return err
	}

	outputJsBuild := filepath.Join(mlflowRepoPath, "server", "js", "build")

	// Copy the MLFlow pre-built UI into .mlflow.repo
	if err := copyDirectory(jsBuild, outputJsBuild); err != nil {
		return err
	}

	// Install it in editable mode
	if err := sh.Run("uv", "pip", "install", "-e", mlflowRepoPath); err != nil {
		return err
	}

	return nil
}
