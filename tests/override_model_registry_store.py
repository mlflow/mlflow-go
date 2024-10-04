from mlflow.store.model_registry.sqlalchemy_store import SqlAlchemyStore

from mlflow_go.store.model_registry import ModelRegistryStore

SqlAlchemyStore = ModelRegistryStore(SqlAlchemyStore)
