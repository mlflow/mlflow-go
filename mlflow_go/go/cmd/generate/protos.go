//nolint:wrapcheck,err113
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
)

const MLFlowCommit = "3effa7380c86946f4557f03aa81119a097d8b433"

var protoFiles = map[string]string{
	"databricks.proto":           "github.com/mlflow/mlflow-go/mlflow_go/go/protos",
	"service.proto":              "github.com/mlflow/mlflow-go/mlflow_go/go/protos",
	"model_registry.proto":       "github.com/mlflow/mlflow-go/mlflow_go/go/protos",
	"databricks_artifacts.proto": "github.com/mlflow/mlflow-go/mlflow_go/go/protos",
	"mlflow_artifacts.proto":     "github.com/mlflow/mlflow-go/mlflow_go/go/protos/artifacts",
	"internal.proto":             "github.com/mlflow/mlflow-go/mlflow_go/go/protos",
	"scalapb/scalapb.proto":      "github.com/mlflow/mlflow-go/mlflow_go/go/protos/scalapb",
}

func downloadProtosFile(fileName, tempDir string) error {
	filePath := path.Join(tempDir, fileName)

	parentDir := path.Dir(filePath)
	if err := os.MkdirAll(parentDir, os.ModePerm); err != nil {
		return err
	}

	out, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	url := fmt.Sprintf("https://raw.githubusercontent.com/mlflow/mlflow/%s/mlflow/protos/%s", MLFlowCommit, fileName)

	//nolint:gosec,noctx
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("file download of %q is not OK", url)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

const fixedArguments = 3

func runProtoc(tempDir string) error {
	arguments := make([]string, 0, len(protoFiles)*2+fixedArguments)

	arguments = append(
		arguments,
		"-I="+tempDir,
		`--go_out=.`,
		`--go_opt=module=github.com/mlflow/mlflow-go`,
	)

	for fileName, goPackage := range protoFiles {
		arguments = append(
			arguments,
			fmt.Sprintf("--go_opt=M%s=%s", fileName, goPackage),
		)
	}

	for fileName := range protoFiles {
		arguments = append(arguments, path.Join(tempDir, fileName))
	}

	cmd := exec.Command("protoc", arguments...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(
			"could not run protoc %s process, err: %s: %w",
			strings.Join(arguments, " "),
			output,
			err,
		)
	}

	return nil
}

func SyncProtos() error {
	tempDir, err := os.MkdirTemp("", "protos")
	if err != nil {
		return err
	}

	defer os.RemoveAll(tempDir)

	for fileName := range protoFiles {
		if err := downloadProtosFile(fileName, tempDir); err != nil {
			return err
		}
	}

	err = runProtoc(tempDir)
	if err != nil {
		return fmt.Errorf("could not run protoc: %w", err)
	}

	return nil
}
