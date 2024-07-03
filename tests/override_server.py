import contextlib
import logging
import os
import sys
from subprocess import Popen

import mlflow
import pytest
from mlflow.server import ARTIFACT_ROOT_ENV_VAR, BACKEND_STORE_URI_ENV_VAR
from mlflow.server.handlers import ModelRegistryStoreRegistryWrapper, TrackingStoreRegistryWrapper
from mlflow.utils import find_free_port

from tests.helper_functions import LOCALHOST
from tests.tracking.integration_test_utils import _await_server_up_or_die

_logger = logging.getLogger(__name__)


@contextlib.contextmanager
def _init_server(backend_uri, root_artifact_uri, extra_env=None, app="mlflow.server:app"):
    """
    Launch a new REST server using the tracking store specified by backend_uri and root artifact
    directory specified by root_artifact_uri.
    :returns A string URL to the server.
    """
    scheme = backend_uri.split("://")[0]
    if ("sqlite" or "postgresql" or "mysql" or "mssql") not in scheme:
        pytest.skip(f'Unsupported scheme "{scheme}" for the Go server')

    mlflow.set_tracking_uri(None)

    server_port = find_free_port()
    python_port = find_free_port()
    url = f"http://{LOCALHOST}:{server_port}"

    _logger.info(
        f"Initializing stores with backend URI {backend_uri} and artifact root {root_artifact_uri}"
    )
    TrackingStoreRegistryWrapper().get_store(backend_uri, root_artifact_uri)
    ModelRegistryStoreRegistryWrapper().get_store(backend_uri, root_artifact_uri)

    _logger.info(
        f"Launching tracking server on {url} with backend URI {backend_uri} and "
        f"artifact root {root_artifact_uri}"
    )
    with Popen(
        [
            sys.executable,
            "-m",
            "mlflow_go.cli",
            "server",
            "--host",
            LOCALHOST,
            "--port",
            str(server_port),
            "--backend-store-uri",
            backend_uri,
            "--default-artifact-root",
            root_artifact_uri,
            "--go-opts",
            ",".join(
                (
                    "log_level=debug",
                    f"python_address={LOCALHOST}:{python_port}",
                    f"python_command={sys.executable} -m flask --app {app} "
                    f"run --host {LOCALHOST} --port {python_port}",
                    "shutdown_timeout=5s",
                )
            ),
        ],
        env={
            **os.environ,
            BACKEND_STORE_URI_ENV_VAR: backend_uri,
            ARTIFACT_ROOT_ENV_VAR: root_artifact_uri,
            **(extra_env or {}),
        },
    ) as proc:
        try:
            _await_server_up_or_die(server_port)
            yield url
        finally:
            proc.terminate()
