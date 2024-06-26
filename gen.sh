#!/bin/sh

# Generate Go files from protos
protoc -I="mlflow_go/protos" \
    --go_out="." \
    --go_opt=module=github.com/mlflow/mlflow-go \
    --go_opt=Mdatabricks.proto=github.com/mlflow/mlflow-go/mlflow_go/go/protos \
    --go_opt=Mservice.proto=github.com/mlflow/mlflow-go/mlflow_go/go/protos \
    --go_opt=Mmodel_registry.proto=github.com/mlflow/mlflow-go/mlflow_go/go/protos \
    --go_opt=Mdatabricks_artifacts.proto=github.com/mlflow/mlflow-go/mlflow_go/go/protos \
    --go_opt=Mmlflow_artifacts.proto=github.com/mlflow/mlflow-go/mlflow_go/go/protos/artifacts \
    --go_opt=Minternal.proto=github.com/mlflow/mlflow-go/mlflow_go/go/protos \
    --go_opt=Mscalapb/scalapb.proto=github.com/mlflow/mlflow-go/mlflow_go/go/protos/scalapb \
    mlflow_go/protos/model_registry.proto \
    mlflow_go/protos/databricks_artifacts.proto \
    mlflow_go/protos/mlflow_artifacts.proto \
    mlflow_go/protos/internal.proto \
    mlflow_go/protos/service.proto \
    mlflow_go/protos/databricks.proto \
    mlflow_go/protos/scalapb/scalapb.proto

# Apply validation
go run ./mlflow_go/go/cmd/generate/ ./mlflow_go/go

# Assert code can be build
go build -o /dev/null ./mlflow_go/go/cmd/server
