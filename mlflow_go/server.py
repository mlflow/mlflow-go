import json
from contextlib import contextmanager

from mlflow_go.lib import get_lib


def launch_server(**config):
    config_bytes = json.dumps(config).encode("utf-8")

    # start the Go server and check for errors
    ret = get_lib().LaunchServer(config_bytes, len(config_bytes))
    if ret != 0:
        raise Exception(f"Non-zero exit code: {ret}")


@contextmanager
def server(**config):
    config_bytes = json.dumps(config).encode("utf-8")

    # start the Go server and check for errors
    id = get_lib().LaunchServerAsync(config_bytes, len(config_bytes))
    if id < 0:
        raise Exception(f"Non-zero exit code: {id}")

    try:
        yield
    finally:
        # stop the Go server and check for errors
        ret = get_lib().StopServer(id)
        if ret != 0:
            raise Exception(f"Non-zero exit code: {ret}")
