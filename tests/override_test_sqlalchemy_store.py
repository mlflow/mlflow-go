import pytest
from mlflow.exceptions import MlflowException
from mlflow.store.tracking.sqlalchemy_store import SqlAlchemyStore


def test_log_batch_internal_error(store: SqlAlchemyStore):
    ()


def test_log_batch_params_max_length_value(store: SqlAlchemyStore, monkeypatch):
    ()


def test_log_param_max_length_value(store: SqlAlchemyStore, monkeypatch):
    ()


def test_set_tag(store: SqlAlchemyStore, monkeypatch):
    ()


def test_log_batch_null_metrics(store: SqlAlchemyStore):
    ()


def test_sqlalchemy_store_behaves_as_expected_with_inmemory_sqlite_db(monkeypatch):
    ()


def test_search_experiments_max_results_validation(store: SqlAlchemyStore):
    with pytest.raises(
        MlflowException,
        match=r"Invalid value 0 for parameter 'max_results' supplied",
    ):
        store.search_experiments(max_results=0)
    with pytest.raises(
        MlflowException,
        match=r"Invalid value 1000000 for parameter 'max_results' supplied",
    ):
        store.search_experiments(max_results=1_000_000)
