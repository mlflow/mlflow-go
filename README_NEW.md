# Go backend for MLflow

In order to increase the performance of the tracking server and the various stores, we propose to rewrite the server and store implementation in Go.


# Topics
* [Installation](#installation)
* [Mage](#mage)
* [Contributing](./docs/CONTRIBUTING.md)
* [Common information](#common-information)
  * [MLflow source code](#mlflow-source-code)
  * [Protos](#protos)
  * [Request validation](#request-validation)
  * [Data access](#data-access)
  * [Linting](#linting)
* [Usage](#usage)
  * [Go Server](#go-server)
  * [Building the Go binary](#building-the-go-binary)
  * [Client-side Go implementation](#client-side-go-implementation)
  * [Go store in Python](#go-store-in-python)
* [Misc](#misc)
  * [Debug Failing Tests](#debug-failing-tests)
  * [Supported endpoints](#supported-endpoints)
  * [Targeting Local Postgres in Python Tests](#targeting-local-postgres-in-python-tests)

## Installation

This package is not yet available on PyPI and currently requires the [Go SDK](https://go.dev) to be installed.

You can then install the package via pip:
```bash
pip install git+https://github.com/jgiannuzzi/mlflow-go.git
```

## Mage

This repository uses [mage](https://magefile.org/) to streamline some utility functions.

```bash
# Install mage (already done in the dev container)
go install github.com/magefile/mage@v1.15.0
```

```bash
# See all targets
mage
```

```bash
# Execute single target
mage generate
```

The beauty of Mage is that we can use regular Go code for our scripting. That being said, we are not married to this tool.

## Common information

### MLflow source code

To integrate with MLflow, you need to include the source code. The [mlflow/mlflow](https://github.com/mlflow/mlflow/) repository contains proto files that define the tracking API. It also includes Python tests that we use to verify our Go implementation produces identical behaviour.

We use a `.mlflow.ref` file to specify the exact location from which to pull our sources. The format should be `remote#reference`, where `remote` is a git remote and `reference` is a branch, tag, or commit SHA.

If the `.mlflow.ref` file is modified and becomes out of sync with the current source files, the mage target will automatically detect this. To manually force a sync, you can run `mage repo:update`.

### Protos

To ensure we stay compatible with the Python implementation, we aim to generate as much as possible based on the `.proto` files.

By running

```bash
mage generate
```

Go code will be generated. Use the protos files from `.mlflow.repo` repository.

This includes the generation of:

- Structs for each endpoint. ([pkg/protos](./pkg/protos/service.pb.go))
- Go interfaces for each service. ([pkg/contract/service/*.g.go](./pkg/contract/service/tracking.g.go))
- [fiber](https://gofiber.io/) routes for each endpoint. ([pkg/server/routes/*.g.go](./pkg/server/routes/tracking.g.go))

If there is any change in the proto files, this should ripple into the Go code.

### Request validation

We use [Go validator](https://github.com/go-playground/validator) to validate all incoming request structs.
As the proto files don't specify any validation rules, we map them manually in [pkg/cmd/generate/validations.go](./cmd/generate/validations.go).

Once the mapping has been done, validation will be invoked automatically in the generated fiber code.

When the need arises, we can write custom validation function in [pkg/validation/validation.go](./validation/validation.go).

### Data access

Initially, we want to focus on supporting Postgres SQL. We chose [Gorm](https://gorm.io/) as ORM to interact with the database.

We do not generate any Go code based on the database schema. Gorm has generation capabilities but they didn't fit our needs. The plan would be to eventually assert the current code still matches the database schema via an integration test.

All the models use pointers for their fields. We do this for performance reasons and to distinguish between zero values and null values.

### Linting

We have enabled various linters from [golangci-lint](https://golangci-lint.run/), you can run these via:

```bash
pre-commit run golangci-lint --all-files
```

Sometimes `golangci-lint` can complain about unrelated files, run `golangci-lint cache clean` to clear the cache.

## Usage

### Go server

To enable use of the Go server, users can run the `mlflow-go server` command.

```bash
# Start the Go server with a database URI
# Other databases are supported as well: sqlite, mysql and mssql
mlflow-go server --backend-store-uri postgresql://postgres:postgres@localhost:5432/postgres
```

This will launch the python process as usual. Within Python, a random port is chosen to start the existing server and a Go child process is spawned. The Go server will use the user specified port (5000 by default) and spawn the actual Python server as its own child process (`gunicorn` or `waitress`).
Any incoming requests the Go server cannot process will be proxied to the existing Python server.

Any Go-specific options can be passed with `--go-opts`, which takes a comma-separated list of key-value pairs.

```bash
mlflow-go server --backend-store-uri postgresql://postgres:postgres@localhost:5432/postgres --go-opts log_level=debug,shutdown_timeout=5s
```

MLflow client could be pointed the Go server:

```python
import mlflow

# Use the Go server
mlflow.set_tracking_uri("http://localhost:5000")

# Use MLflow as usual
mlflow.set_experiment("my-experiment")

with mlflow.start_run():
    mlflow.log_param("param", 1)
    mlflow.log_metric("metric", 2)
```

### Building the Go binary

To ensure everything still compiles:

```bash
go build -o /dev/null ./pkg/cmd/server
```

or

```bash
python -m mlflow_go.lib . /tmp
```

### Client-side Go implementation

```python
import mlflow
import mlflow_go

# Enable the Go client implementation (disabled by default)
mlflow_go.enable_go()

# Set the tracking URI (you can also set it via the environment variable MLFLOW_TRACKING_URI)
# Currently only database URIs are supported
mlflow.set_tracking_uri("sqlite:///mlflow.db")

# Use MLflow as usual
mlflow.set_experiment("my-experiment")

with mlflow.start_run():
    mlflow.log_param("param", 1)
    mlflow.log_metric("metric", 2)
```

### Go store in Python

```python
import logging
import mlflow
import mlflow_go

# Enable debug logging
logging.basicConfig()
logging.getLogger('mlflow_go').setLevel(logging.DEBUG)

# Enable the Go client implementation (disabled by default)
mlflow_go.enable_go()

# Instantiate the tracking store with a database URI
tracking_store = mlflow.tracking._tracking_service.utils._get_store('sqlite:///mlflow.db')

# Call any tracking store method
tracking_store.get_experiment(0)

# Instantiate the model registry store with a database URI
model_registry_store = mlflow.tracking._model_registry.utils._get_store('sqlite:///mlflow.db')

# Call any model registry store method
model_registry_store.get_latest_versions("model")
```

## Misc

### Debug Failing Tests

Sometimes, it can be very useful to modify failing tests and use `print` statements to display the current state or differences between objects from Python or Go services.

Adding `"-vv"` to the `pytest` command in `magefiles/tests.go` can also provide more information when assertions are not met.

### Targeting Local Postgres in Python Tests

At times, you might want to apply store calls to your local database to investigate certain read operations via the local tracking server.

You can achieve this by changing:

```python
def test_search_runs_datasets(store: SqlAlchemyStore):
```

to:

```python
def test_search_runs_datasets():
    db_uri = "postgresql://postgres:postgres@localhost:5432/postgres"
    artifact_uri = Path("/tmp/artifacts")
    artifact_uri.mkdir(exist_ok=True)
    store = SqlAlchemyStore(db_uri, artifact_uri.as_uri())
```

in the test file located in `.mlflow.repo`.

### Supported endpoints

The currently supported endpoints can be found by running mage command:

```bash
mage endpoints
```

