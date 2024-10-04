from mlflow.store.tracking.sqlalchemy_store import SqlAlchemyStore

from mlflow_go.store.tracking import TrackingStore

SqlAlchemyStore = TrackingStore(SqlAlchemyStore)
