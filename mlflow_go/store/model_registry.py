import json
import logging

from mlflow.entities.model_registry import (
    ModelVersion,
    RegisteredModel,
)
from mlflow.protos.model_registry_pb2 import (
    CreateRegisteredModel,
    GetLatestVersions,
)

from mlflow_go import is_go_enabled
from mlflow_go.lib import get_lib
from mlflow_go.store._service_proxy import _ServiceProxy

_logger = logging.getLogger(__name__)


class _ModelRegistryStore:
    def __init__(self, *args, **kwargs):
        store_uri = args[0] if len(args) > 0 else kwargs.get("db_uri", kwargs.get("root_directory"))
        config = json.dumps(
            {
                "model_registry_store_uri": store_uri,
                "log_level": logging.getLevelName(_logger.getEffectiveLevel()),
            }
        ).encode("utf-8")
        self.service = _ServiceProxy(get_lib().CreateModelRegistryService(config, len(config)))
        super().__init__(store_uri)

    def __del__(self):
        if hasattr(self, "service"):
            get_lib().DestroyModelRegistryService(self.service.id)

    def get_latest_versions(self, name, stages=None):
        request = GetLatestVersions(
            name=name,
            stages=stages,
        )
        response = self.service.call_endpoint(
            get_lib().ModelRegistryServiceGetLatestVersions, request
        )
        return [ModelVersion.from_proto(mv) for mv in response.model_versions]

    def create_registered_model(self, name, tags=None, description=None):
        request = CreateRegisteredModel(
            name=name,
            tags=[tag.to_proto() for tag in tags] if tags else [],
            description=description,
        )
        response = self.service.call_endpoint(
            get_lib().ModelRegistryServiceCreateRegisteredModel, request
        )
        entity = RegisteredModel.from_proto(response.registered_model)
        if not response.registered_model.HasField("description"):
            entity.description = None

        return entity


def ModelRegistryStore(cls):
    return type(cls.__name__, (_ModelRegistryStore, cls), {})


def _get_sqlalchemy_store(store_uri):
    from mlflow.store.model_registry.sqlalchemy_store import SqlAlchemyStore

    if is_go_enabled():
        SqlAlchemyStore = ModelRegistryStore(SqlAlchemyStore)

    return SqlAlchemyStore(store_uri)
