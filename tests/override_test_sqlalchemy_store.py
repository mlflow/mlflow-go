from mlflow.store.tracking.sqlalchemy_store import SqlAlchemyStore


def test_log_batch_internal_error(store: SqlAlchemyStore):
    ()


def test_log_batch_params_max_length_value(store: SqlAlchemyStore, monkeypatch):
    ()
