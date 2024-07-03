//go:build mage

package main

import (
	"os"
	"path"

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
	mlflowRepoFolderName = ".mlflow"
	reposityUrl          = "https://github.com/jgiannuzzi/mlflow.git"
	branch               = "server-signals"
)

func updateRepo(repoPath string) error {
	return git("--work-tree", repoPath, "reset", "--hard")
}

func freshCheckout() error {
	return git(
		"clone",
		"--branch", branch,
		"--single-branch", reposityUrl,
		mlflowRepoFolderName,
	)
}

// Clone or reset the .mlflow fork.
func Init() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	repoPath := path.Join(pwd, mlflowRepoFolderName)

	if folderExists(repoPath) {
		return updateRepo(repoPath)
	} else {
		return freshCheckout()
	}

	return nil
}
