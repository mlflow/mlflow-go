import os

_go_enabled = "MLFLOW_GO_ENABLED" in os.environ


def _set_go_enabled(enabled: bool):
    global _go_enabled
    _go_enabled = enabled


def is_go_enabled():
    return _go_enabled


def disable_go():
    _set_go_enabled(False)


def enable_go():
    _set_go_enabled(True)
