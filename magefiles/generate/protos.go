package generate

import (
	"fmt"
	"os/exec"
	"path"
	"strings"
)

const MLFlowCommit = "3effa7380c86946f4557f03aa81119a097d8b433"

var protoFiles = map[string]string{
	"databricks.proto":           "github.com/mlflow/mlflow-go/pkg/protos",
	"service.proto":              "github.com/mlflow/mlflow-go/pkg/protos",
	"model_registry.proto":       "github.com/mlflow/mlflow-go/pkg/protos",
	"databricks_artifacts.proto": "github.com/mlflow/mlflow-go/pkg/protos",
	"mlflow_artifacts.proto":     "github.com/mlflow/mlflow-go/pkg/protos/artifacts",
	"internal.proto":             "github.com/mlflow/mlflow-go/pkg/protos",
	"scalapb/scalapb.proto":      "github.com/mlflow/mlflow-go/pkg/protos/scalapb",
}

const fixedArguments = 3

func RunProtoc(protoDir string) error {
	arguments := make([]string, 0, len(protoFiles)*2+fixedArguments)

	arguments = append(
		arguments,
		"-I="+protoDir,
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
		arguments = append(arguments, path.Join(protoDir, fileName))
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
