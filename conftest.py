import logging
import pathlib
from unittest.mock import patch

_logger = logging.getLogger(__name__)


def load_new_function(file_path, func_name):
    with open(file_path) as f:
        new_func_code = f.read()

    local_dict = {}
    exec(new_func_code, local_dict)
    return local_dict[func_name]


def pytest_configure(config):
    for func_to_patch, new_func_file_relative in (
        (
            "tests.tracking.integration_test_utils._init_server",
            "tests/override_server.py",
        ),
    ):
        func_name = func_to_patch.rsplit(".", 1)[1]
        new_func_file = (
            pathlib.Path(__file__).parent.joinpath(new_func_file_relative).resolve().as_posix()
        )

        new_func = load_new_function(new_func_file, func_name)

        _logger.info(f"Patching function: {func_to_patch}")
        patch(func_to_patch, new_func).start()
