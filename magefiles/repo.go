//go:build mage

//nolint:wrapcheck
package main

// rename to repo.go

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	MLFlowRepoFolderName = ".mlflow.repo"
)

type Repo mg.Namespace

func folderExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return info.IsDir()
}

func git(args ...string) error {
	return sh.RunV("git", args...)
}

func gitMlflowRepo(args ...string) error {
	allArgs := append([]string{"-C", MLFlowRepoFolderName}, args...)

	return sh.RunV("git", allArgs...)
}

func gitMlflowRepoOutput(args ...string) (string, error) {
	allArgs := append([]string{"-C", MLFlowRepoFolderName}, args...)

	return sh.Output("git", allArgs...)
}

type gitReference struct {
	remote    string
	reference string
}

const refFileName = ".mlflow.ref"

func readFile(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

var ErrInvalidGitRefFormat = errors.New("invalid format in .mlflow.ref file: expected 'remote#reference'")

func readGitReference() (gitReference, error) {
	refFilePath, err := filepath.Abs(refFileName)
	if err != nil {
		return gitReference{}, fmt.Errorf("failed to get .mlflow.ref: %w", err)
	}

	content, err := readFile(refFilePath)
	if err != nil {
		return gitReference{}, err
	}

	parts := strings.Split(content, "#")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return gitReference{}, ErrInvalidGitRefFormat
	}

	remote := strings.TrimSpace(parts[0])
	reference := strings.TrimSpace(parts[1])

	return gitReference{remote: remote, reference: reference}, nil
}

func freshCheckout(gitReference gitReference) error {
	if err := git("clone", "--no-checkout", gitReference.remote, MLFlowRepoFolderName); err != nil {
		return err
	}

	return gitMlflowRepo("checkout", gitReference.reference)
}

func checkRemote(gitReference gitReference) bool {
	// git -C .mlflow.repo remote get-url origin
	output, err := gitMlflowRepoOutput("remote", "get-url", "origin")
	if err != nil {
		return false
	}

	return strings.TrimSpace(output) == gitReference.remote
}

func checkBranch(gitReference gitReference) bool {
	// git -C .mlflow.repo rev-parse --abbrev-ref HEAD
	output, err := gitMlflowRepoOutput("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return false
	}

	return strings.TrimSpace(output) == gitReference.reference
}

func checkTag(gitReference gitReference) bool {
	// git -C .mlflow.repo  describe --tags HEAD
	output, err := gitMlflowRepoOutput("describe", "--tags", "HEAD")
	if err != nil {
		return false
	}

	return strings.TrimSpace(output) == gitReference.reference
}

func checkCommit(gitReference gitReference) bool {
	// git -C .mlflow.repo rev-parse HEAD
	output, err := gitMlflowRepoOutput("rev-parse", "HEAD")
	if err != nil {
		return false
	}

	return strings.TrimSpace(output) == gitReference.reference
}

func checkReference(gitReference gitReference) bool {
	switch {
	case checkBranch(gitReference):
		log.Printf("Already on branch %q", gitReference.reference)

		return true
	case checkTag(gitReference):
		log.Printf("Already on tag %q", gitReference.reference)

		return true
	case checkCommit(gitReference):
		log.Printf("Already on commit %q", gitReference.reference)

		return true
	}

	return false
}

func syncRepo(gitReference gitReference) error {
	log.Printf("syncing mlflow repo to %s#%s", gitReference.remote, gitReference.reference)

	if err := gitMlflowRepo("remote", "set-url", "origin", gitReference.remote); err != nil {
		return err
	}

	if err := gitMlflowRepo("fetch", "origin"); err != nil {
		return err
	}

	if err := gitMlflowRepo("checkout", gitReference.reference); err != nil {
		return err
	}

	if checkBranch(gitReference) {
		return gitMlflowRepo("pull")
	}

	return nil
}

// Clone or reset the .mlflow.repo fork.
func (Repo) Init() error {
	gitReference, err := readGitReference()
	if err != nil {
		return err
	}

	repoPath, err := filepath.Abs(MLFlowRepoFolderName)
	if err != nil {
		return err
	}

	if !folderExists(repoPath) {
		return freshCheckout(gitReference)
	}

	// Verify remote
	if !checkRemote(gitReference) {
		log.Printf("Remote %s no longer matches", gitReference.remote)

		return syncRepo(gitReference)
	}

	// Verify reference
	if !checkReference(gitReference) {
		log.Printf("The current reference %q no longer matches", gitReference.reference)

		return syncRepo(gitReference)
	}

	return nil
}

// Forcefully update the .mlflow.repo according to the .mlflow.ref.
func (Repo) Update() error {
	gitReference, err := readGitReference()
	if err != nil {
		return err
	}

	return syncRepo(gitReference)
}
