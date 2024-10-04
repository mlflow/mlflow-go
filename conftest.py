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
        (
            "mlflow.store.tracking.sqlalchemy_store.SqlAlchemyStore",
            "tests/override_tracking_store.py",
        ),
        (
            "mlflow.store.model_registry.sqlalchemy_store.SqlAlchemyStore",
            "tests/override_model_registry_store.py",
        ),
        # This test will patch some Python internals to invoke an internal exception.
        # We cannot do this in Go.
        (
            "tests.store.tracking.test_sqlalchemy_store.test_log_batch_internal_error",
            "tests/override_test_sqlalchemy_store.py",
        ),
        # This test uses monkeypatch.setenv which does not flow through to the
        (
            "tests.store.tracking.test_sqlalchemy_store.test_log_batch_params_max_length_value",
            "tests/override_test_sqlalchemy_store.py",
        ),
        # This tests calls the store using invalid metric entity that cannot be converted
        # to its proto counterpart.
        # Example: entities.Metric("invalid_metric", None, (int(time.time() * 1000)), 0).to_proto()
        (
            "tests.store.tracking.test_sqlalchemy_store.test_log_batch_null_metrics",
            "tests/override_test_sqlalchemy_store.py",
        ),
        # We do not support applying the SQL schema to sqlite like Python does.
        # So we do not support sqlite:////:memory: database.
        (
            "tests.store.tracking.test_sqlalchemy_store.test_sqlalchemy_store_behaves_as_expected_with_inmemory_sqlite_db",
            "tests/override_test_sqlalchemy_store.py",
        ),
    ):
        func_name = func_to_patch.rsplit(".", 1)[1]
        new_func_file = (
            pathlib.Path(__file__).parent.joinpath(new_func_file_relative).resolve().as_posix()
        )

        new_func = load_new_function(new_func_file, func_name)

        _logger.info(f"Patching function: {func_to_patch}")
        patch(func_to_patch, new_func).start()
