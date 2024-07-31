# Go backend for MLflow

In order to increase the performance of the tracking server and the various stores, we propose to rewrite the server and store implementation in Go.

## Usage

### Installation

This package is not yet available on PyPI and currently requires the [Go SDK](https://go.dev) to be installed.

You can then install the package via pip:
```bash
pip install git+https://github.com/jgiannuzzi/mlflow-go.git
```

### Using the Go server

```bash
# With SQLite
mlflow-go server --backend-store-uri sqlite:///mlflow.db

# With Postgres
mlflow-go server --backend-store-uri postgresql://postgres:postgres@localhost:5432/postgres
```

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

### Using the client-side Go implementation

```python
import mlflow
import mlflow_go

# Enable the Go client implementation (disabled by default)
mlflow_go.enable_go()

# With SQLite
mlflow.set_tracking_uri("sqlite:///mlflow.db")

# With Postgres
#mlflow.set_tracking_uri("postgresql://postgres:postgres@localhost:5432/postgres")

# Use MLflow as usual
mlflow.set_experiment("my-experiment")

with mlflow.start_run():
    mlflow.log_param("param", 1)
    mlflow.log_metric("metric", 2)
```

## Temp stuff

### Dev setup

```bash
# Install our Python package and its dependencies
pip install -e .

# Install the dreaded psycho
pip install psycopg2-binary

# Archive the MLFlow pre-built UI
tar -C /usr/local/python/current/lib/python3.8/site-packages/mlflow -czvf ./ui.tgz ./server/js/build

# Clone the MLflow repo
git clone https://github.com/jgiannuzzi/mlflow.git -b master .mlflow.repo

# Add the UI back to it
tar -C .mlflow.repo/mlflow -xzvf ./ui.tgz

# Install it in editable mode
pip install -e .mlflow.repo
```

or run `mage temp`.

### Run the tests manually

```bash
# Build the Go binary in a temporary directory
libpath=$(mktemp -d)
python -m mlflow_go.lib . $libpath

# Run the tests (currently just the server ones)
MLFLOW_GO_LIBRARY_PATH=$libpath pytest --confcutdir=. \
  .mlflow.repo/tests/tracking/test_rest_tracking.py \
  .mlflow.repo/tests/tracking/test_model_registry.py \
  .mlflow.repo/tests/store/tracking/test_sqlalchemy_store.py \
  .mlflow.repo/tests/store/model_registry/test_sqlalchemy_store.py \
  -k 'not [file'

# Remove the Go binary
rm -rf $libpath

# If you want to run a specific test with more verbosity
# -s for live output
# --log-level=debug for more verbosity (passed down to the Go server/stores)
MLFLOW_GO_LIBRARY_PATH=$libpath pytest --confcutdir=. \
  .mlflow.repo/tests/tracking/test_rest_tracking.py::test_create_experiment_validation \
  -k 'not [file' \
  -s --log-level=debug
```

Or run the `mage test:python` target.

## General setup

### Mage

This repository uses [mage](https://magefile.org/) to streamline some utilily functions.

```bash
# Install mage (already done in the dev container)
go install github.com/magefile/mage@v1.15.0

# See all targets
mage

# Execute single target
mage dev
```

The beauty of Mage is that we can use regular Go code for our scripting.  
That being said, we are not married to this tool.

### mlflow source code

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

This incudes the generation of:

- Structs for each endpoint. ([pkg/protos](./protos/service.pb.go))
- Go interfaces for each service. ([pkg/contract/service/*.g.go](./contract/service/tracking.g.go))
- [fiber](https://gofiber.io/) routes for each endpoint. ([pkg/server/routes/*.g.go](./server/routes/tracking.g.go))

If there is any change in the proto files, this should ripple into the Go code.

## Launching the Go server

To enable use of the Go server, users can run the `mlflow-go server` command.

```bash
mlflow-go server --backend-store-uri postgresql://postgres:postgres@localhost:5432/postgres
```

This will launch the python process as usual. Within Python, a random port is chosen to start the existing server and a Go child process is spawned. The Go server will use the user specified port (5000 by default) and spawn the actual Python server as its own child process (`gunicorn` or `waitress`).
Any incoming requests the Go server cannot process will be proxied to the existing Python server.

Any Go-specific options can be passed with `--go-opts`, which takes a comma-separated list of key-value pairs.

```bash
mlflow-go server --backend-store-uri postgresql://postgres:postgres@localhost:5432/postgres --go-opts log_level=debug,shutdown_timeout=5s
```

## Building the Go binary

To ensure everything still compiles:

```bash
go build -o /dev/null ./pkg/cmd/server
```

## Request validation

We use [Go validator](https://github.com/go-playground/validator) to validate all incoming request structs.
As the proto files don't specify any validation rules, we map them manually in [pkg/cmd/generate/validations.go](./cmd/generate/validations.go).

Once the mapping has been done, validation will be invoked automatically in the generated fiber code.

When the need arises, we can write custom validation function in [pkg/validation/validation.go](./validation/validation.go).

## Data access

Initially, we want to focus on supporting Postgres SQL. We chose [Gorm](https://gorm.io/) as ORM to interact with the database.

We do not generate any Go code based on the database schema. Gorm has generation capabilities but they didn't fit our needs. The plan would be to eventually assert the current code stil matches the database schema via an intergration test.

All the models use pointers for their fields. We do this for performance reasons and to distinguish between zero values and null values.

## Testing strategy

> [!WARNING]
> TODO rewrite this whole section

The Python integration tests have been adapted to also run against the Go implementation. Just run them as usual, e.g.

```bash
pytest tests/tracking/test_rest_tracking.py
```

To run only the tests targetting the Go implementation, you can use the `-k` flag:

```bash
pytest tests/tracking/test_rest_tracking.py -k '[go-'
```

If you'd like to run a specific test and see its output 'live', you can use the `-s` flag:

```bash
pytest -s "tests/tracking/test_rest_tracking.py::test_create_experiment_validation[go-postgresql]"
```

See the [pytest documentation](https://docs.pytest.org/en/8.2.x/how-to/usage.html#specifying-which-tests-to-run) for more details.

## Supported endpoints

The currently supported endpoints can be found in [mlflow/cmd/generate/endspoints.go](./cmd/generate/endspoints.go).

## Linters

We have enabled various linters from [golangci-lint](https://golangci-lint.run/), you can run these via:

```bash
pre-commit run golangci-lint --all-files
```
