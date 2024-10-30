//go:build mage

//nolint:wrapcheck
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/codeclysm/extract"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Release struct {
	URL         string `json:"url"`
	PackageType string `json:"packagetype"`
}

type PyPIPackage struct {
	Releases map[string][]Release `json:"releases"`
}

var errNoReleasesFound = errors.New("no mlflow sdist releases found")

func getMlflowArtifact(ctx context.Context, version string) (string, error) {
	url := "https://pypi.org/pypi/mlflow/json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to perform request: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var pypiData PyPIPackage

	err = json.Unmarshal(body, &pypiData)
	if err != nil {
		return "", err
	}

	if releases, found := pypiData.Releases[version]; found && len(releases) > 0 {
		for _, r := range releases {
			if r.PackageType == "sdist" {
				return r.URL, nil
			}
		}
	}

	return "", fmt.Errorf("version %s: %w", version, errNoReleasesFound)
}

func downloadTarFile(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	tempDir := os.TempDir()

	tempFile, err := os.CreateTemp(tempDir, "*.tar.gz")
	if err != nil {
		return "", err
	}

	defer tempFile.Close()

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return "", err
	}

	return tempFile.Name(), nil
}

func extractedTar(ctx context.Context, tarball string) (string, error) {
	data, err := os.ReadFile(tarball)
	if err != nil {
		return "", err
	}

	buffer := bytes.NewBuffer(data)
	destination := strings.TrimSuffix(tarball, ".tar.gz")

	err = extract.Gz(ctx, buffer, destination, nil)
	if err != nil {
		return "", err
	}

	return destination, err
}

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

var mlflowVersion = "2.17.0"

// Configure development environment.
func Configure(ctx context.Context) error {
	mg.Deps(Repo.Init)

	// create virtual environment using uv
	if err := sh.Run("uv", "venv", "--python", "3.9"); err != nil {
		return err
	}

	// Install our Python package and its dependencies
	if err := sh.RunV("uv", "sync", "--all-extras"); err != nil {
		return err
	}

	// Download the current mlflow source artifact from PyPi,
	// we do this to grab the compiled JavaScript assets.
	mlflowArtifactURL, err := getMlflowArtifact(ctx, mlflowVersion)
	if err != nil {
		return err
	}

	tarball, err := downloadTarFile(ctx, mlflowArtifactURL)
	if err != nil {
		return err
	}

	defer os.RemoveAll(tarball)

	extractedDir, err := extractedTar(ctx, tarball)
	if err != nil {
		return err
	}

	defer os.RemoveAll(extractedDir)

	jsBuild := filepath.Join(extractedDir, "mlflow-"+mlflowVersion, "mlflow", "server", "js", "build")

	mlflowRepoPath, err := filepath.Abs(MLFlowRepoFolderName)
	if err != nil {
		return err
	}

	// Copy the MLFlow pre-built UI into .mlflow.repo
	outputJsBuild := filepath.Join(mlflowRepoPath, "mlflow", "server", "js", "build")
	if err := copyDirectory(jsBuild, outputJsBuild); err != nil {
		return err
	}

	return nil
}
