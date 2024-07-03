//go:build mage

package main

import (
	"errors"
	"fmt"
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

const (
	MLFlowRepoFolderName = ".mlflow"
	reposityURL          = "https://github.com/jgiannuzzi/mlflow.git"
	branch               = "server-signals"
)

// remote (url)
// reference (branch, tag, sha)

// content of mlfow.ref file

func updateRepo(repoPath string) error {
	// if sha didn't change, skip

	// set remote
	// git fetch
	return git("--work-tree", repoPath, "reset", "--hard")
}

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

	return git("-C", MLFlowRepoFolderName, "checkout", gitReference.reference)
}

func checkRemote(gitReference gitReference) bool {
	// git -C .mlflow remote get-url origin
	output, err := sh.Output(
		"git", "-C", MLFlowRepoFolderName,
		"remote", "get-url", "origin",
	)
	if err != nil {
		return false
	}

	return strings.TrimSpace(output) == gitReference.remote
}

func checkBranch(gitReference gitReference) bool {
	// git -C .mlflow rev-parse --abbrev-ref HEAD
	output, err := sh.Output(
		"git", "-C", MLFlowRepoFolderName,
		"rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return false
	}

	return strings.TrimSpace(output) == gitReference.reference
}

func checkTag(gitReference gitReference) bool {
	// git -C .mlflow  describe --exact-match --tags HEAD
	output, err := sh.Output(
		"git", "-C", MLFlowRepoFolderName,
		"describe", "--tags", "HEAD",
	)
	if err != nil {
		return false
	}

	return strings.TrimSpace(output) == gitReference.reference
}

func checkCommit(gitReference gitReference) bool {
	// git -C .mlflow rev-parse --short HEAD
	output, err := sh.Output(
		"git", "-C", MLFlowRepoFolderName,
		"rev-parse", "--short", "HEAD",
	)
	if err != nil {
		return false
	}

	return strings.TrimSpace(output) == gitReference.reference
}

func syncRepo(gitReference gitReference) error {
	print("Syncing mlflow repo\n")

	if err := git("remote", "set-url", "origin", gitReference.remote); err != nil {
		return err
	}

	if err := git("-C", MLFlowRepoFolderName, "fetch", "origin"); err != nil {
		return err
	}

	return git("-C", MLFlowRepoFolderName, "checkout", gitReference.reference)
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
		return syncRepo(gitReference)
	}

	// Verify reference
	switch {
	case checkBranch(gitReference):
		return nil
	case checkTag(gitReference):
		return nil
	case checkCommit(gitReference):
		return nil
	}

	return syncRepo(gitReference)
}

// SHA => config file SHA

// set remote
// git fetch
// compare with existing SHA

func Update() error {
	// Always do fetch
	return nil
}
