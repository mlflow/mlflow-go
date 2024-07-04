//go:build mage

//nolint:wrapcheck
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/magefile/mage/sh"
)

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

const (
	MLFlowRepoFolderName = ".mlflow.repo"
	reposityURL          = "https://github.com/jgiannuzzi/mlflow.git"
	branch               = "server-signals"
)

type gitReference struct {
	remote    string
	reference string
	pwd       string
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
	pwd, err := os.Getwd()
	if err != nil {
		return gitReference{}, fmt.Errorf("failed to get working directory: %w", err)
	}

	content, err := readFile(path.Join(pwd, refFileName))
	if err != nil {
		return gitReference{}, err
	}

	parts := strings.Split(content, "#")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return gitReference{}, ErrInvalidGitRefFormat
	}

	remote := strings.TrimSpace(parts[0])
	reference := strings.TrimSpace(parts[1])

	return gitReference{remote: remote, reference: reference, pwd: pwd}, nil
}

func freshCheckout(gitReference gitReference) error {
	if err := git("clone", "--no-checkout", gitReference.remote, MLFlowRepoFolderName); err != nil {
		return err
	}

	return gitMlflowRepo("checkout", gitReference.reference)
}

func checkRemote(gitReference gitReference) bool {
	// git -C .mlflow remote get-url origin
	output, err := gitMlflowRepoOutput("remote", "get-url", "origin")
	if err != nil {
		return false
	}

	return strings.TrimSpace(output) == gitReference.remote
}

func checkBranch(gitReference gitReference) bool {
	// git -C .mlflow rev-parse --abbrev-ref HEAD
	output, err := gitMlflowRepoOutput("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return false
	}

	return strings.TrimSpace(output) == gitReference.reference
}

func checkTag(gitReference gitReference) bool {
	// git -C .mlflow  describe --exact-match --tags HEAD
	output, err := gitMlflowRepoOutput("describe", "--tags", "HEAD")
	if err != nil {
		return false
	}

	return strings.TrimSpace(output) == gitReference.reference
}

func checkCommit(gitReference gitReference) bool {
	// git -C .mlflow rev-parse --short HEAD
	output, err := gitMlflowRepoOutput("rev-parse", "HEAD")
	if err != nil {
		return false
	}

	return strings.TrimSpace(output) == gitReference.reference
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

// Clone or reset the .mlflow fork.
func Init() error {
	gitReference, err := readGitReference()
	if err != nil {
		return err
	}

	repoPath := path.Join(gitReference.pwd, MLFlowRepoFolderName)
	if !folderExists(repoPath) {
		return freshCheckout(gitReference)
	}

	// Verify remote
	if !checkRemote(gitReference) {
		log.Printf("Remote %s no longer matches", gitReference.remote)

		return syncRepo(gitReference)
	}

	// Verify reference
	switch {
	case checkBranch(gitReference):
		log.Printf("Already on branch %q", gitReference.reference)

		return nil
	case checkTag(gitReference):
		log.Printf("Already on tag %q", gitReference.reference)

		return nil
	case checkCommit(gitReference):
		log.Printf("Already on commit %q", gitReference.reference)

		return nil
	}

	log.Printf("The current reference %q no longer matches", gitReference.reference)

	return syncRepo(gitReference)
}

// Forcefully update the .mlflow.repo according to the .mlflow.ref.
func Update() error {
	gitReference, err := readGitReference()
	if err != nil {
		return err
	}

	return syncRepo(gitReference)
}
