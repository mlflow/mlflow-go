import json
import logging
from typing import Dict

from mlflow.entities import (
    Experiment,
    Run,
    RunInfo,
    TraceInfo,
    ViewType,
)
from mlflow.environment_variables import MLFLOW_TRUNCATE_LONG_VALUES
from mlflow.exceptions import MlflowException
from mlflow.protos import databricks_pb2
from mlflow.protos.service_pb2 import (
    CreateExperiment,
    CreateRun,
    DeleteExperiment,
    DeleteRun,
    DeleteTag,
    DeleteTraceTag,
    GetExperiment,
    GetExperimentByName,
    GetRun,
    LogBatch,
    LogMetric,
    LogParam,
    RestoreExperiment,
    RestoreRun,
    SearchExperiments,
    SearchRuns,
    SetTag,
    SetTraceTag,
    StartTrace,
    TraceRequestMetadata,
    TraceTag,
    UpdateExperiment,
    UpdateRun,
)
from mlflow.store.entities import PagedList
from mlflow.store.tracking import SEARCH_MAX_RESULTS_DEFAULT
from mlflow.utils.uri import resolve_uri_if_local

from mlflow_go import is_go_enabled
from mlflow_go.lib import get_lib
from mlflow_go.store._service_proxy import _ServiceProxy

_logger = logging.getLogger(__name__)


class _TrackingStore:
    def __init__(self, *args, **kwargs):
        store_uri = args[0] if len(args) > 0 else kwargs.get("db_uri", kwargs.get("root_directory"))
        default_artifact_root = (
            args[1]
            if len(args) > 1
            else kwargs.get("default_artifact_root", kwargs.get("artifact_root_uri"))
        )
        config = json.dumps(
            {
                "log_level": logging.getLevelName(_logger.getEffectiveLevel()),
                "python_tests_env": {
                    "MLFLOW_TRUNCATE_LONG_VALUES": MLFLOW_TRUNCATE_LONG_VALUES.get()
                },
                "tracking_store_uri": store_uri,
                "default_artifact_root": resolve_uri_if_local(default_artifact_root),
            }
        ).encode("utf-8")
        self.service = _ServiceProxy(get_lib().CreateTrackingService(config, len(config)))
        super().__init__(store_uri, default_artifact_root)

    def __del__(self):
        if hasattr(self, "service"):
            get_lib().DestroyTrackingService(self.service.id)

    def get_experiment(self, experiment_id):
        request = GetExperiment(experiment_id=str(experiment_id))
        response = self.service.call_endpoint(get_lib().TrackingServiceGetExperiment, request)
        return Experiment.from_proto(response.experiment)

    def get_experiment_by_name(self, experiment_name):
        request = GetExperimentByName(experiment_name=experiment_name)
        try:
            response = self.service.call_endpoint(
                get_lib().TrackingServiceGetExperimentByName, request
            )
            return Experiment.from_proto(response.experiment)
        except MlflowException as e:
            if e.error_code == databricks_pb2.ErrorCode.Name(
                databricks_pb2.RESOURCE_DOES_NOT_EXIST
            ):
                return None
            raise

    def create_experiment(self, name, artifact_location=None, tags=None):
        request = CreateExperiment(
            name=name,
            artifact_location=artifact_location,
            tags=[tag.to_proto() for tag in tags] if tags else [],
        )
        response = self.service.call_endpoint(get_lib().TrackingServiceCreateExperiment, request)
        return response.experiment_id

    def delete_experiment(self, experiment_id):
        request = DeleteExperiment(experiment_id=str(experiment_id))
        self.service.call_endpoint(get_lib().TrackingServiceDeleteExperiment, request)

    def restore_experiment(self, experiment_id):
        request = RestoreExperiment(experiment_id=str(experiment_id))
        self.service.call_endpoint(get_lib().TrackingServiceRestoreExperiment, request)

    def rename_experiment(self, experiment_id, new_name):
        request = UpdateExperiment(experiment_id=str(experiment_id), new_name=new_name)
        self.service.call_endpoint(get_lib().TrackingServiceUpdateExperiment, request)

    def get_run(self, run_id):
        request = GetRun(run_uuid=run_id, run_id=run_id)
        response = self.service.call_endpoint(get_lib().TrackingServiceGetRun, request)
        return Run.from_proto(response.run)

    def create_run(self, experiment_id, user_id, start_time, tags, run_name):
        request = CreateRun(
            experiment_id=str(experiment_id),
            user_id=user_id,
            start_time=start_time,
            tags=[tag.to_proto() for tag in tags] if tags else [],
            run_name=run_name,
        )
        response = self.service.call_endpoint(get_lib().TrackingServiceCreateRun, request)
        return Run.from_proto(response.run)

    def delete_run(self, run_id):
        request = DeleteRun(run_id=run_id)
        self.service.call_endpoint(get_lib().TrackingServiceDeleteRun, request)

    def restore_run(self, run_id):
        request = RestoreRun(run_id=run_id)
        self.service.call_endpoint(get_lib().TrackingServiceRestoreRun, request)

    def update_run(self, run_id, run_status, end_time, run_name):
        request = UpdateRun(
            run_uuid=run_id,
            run_id=run_id,
            status=run_status,
            end_time=end_time,
            run_name=run_name,
        )
        response = self.service.call_endpoint(get_lib().TrackingServiceUpdateRun, request)
        return RunInfo.from_proto(response.run_info)

    def _search_runs(
        self, experiment_ids, filter_string, run_view_type, max_results, order_by, page_token
    ):
        request = SearchRuns(
            experiment_ids=[str(experiment_id) for experiment_id in experiment_ids],
            filter=filter_string,
            run_view_type=ViewType.to_proto(run_view_type),
            max_results=max_results,
            order_by=order_by,
            page_token=page_token,
        )
        response = self.service.call_endpoint(get_lib().TrackingServiceSearchRuns, request)
        runs = [Run.from_proto(proto_run) for proto_run in response.runs]
        return runs, (response.next_page_token or None)

    def log_batch(self, run_id, metrics, params, tags):
        request = LogBatch(
            run_id=run_id,
            metrics=[metric.to_proto() for metric in metrics],
            params=[param.to_proto() for param in params],
            tags=[tag.to_proto() for tag in tags],
        )
        self.service.call_endpoint(get_lib().TrackingServiceLogBatch, request)

    def log_metric(self, run_id, metric):
        request = LogMetric(
            run_id=run_id,
            key=metric.key,
            value=metric.value,
            timestamp=metric.timestamp,
            step=metric.step,
        )
        self.service.call_endpoint(get_lib().TrackingServiceLogMetric, request)

    def log_param(self, run_id, param):
        request = LogParam(
            run_id=run_id,
            key=param.key,
            value=param.value,
        )
        self.service.call_endpoint(get_lib().TrackingServiceLogParam, request)

    def set_trace_tag(self, request_id: str, key: str, value: str):
        request = SetTraceTag(
            key=key,
            value=value,
            request_id=request_id,
        )
        self.service.call_endpoint(get_lib().TrackingServiceSetTraceTag, request)

    def delete_tag(self, run_id, key):
        request = DeleteTag(run_id=run_id, key=key)
        self.service.call_endpoint(get_lib().TrackingServiceDeleteTag, request)

    def delete_trace_tag(self, request_id: str, key: str):
        request = DeleteTraceTag(
            request_id=request_id,
            key=key,
        )
        self.service.call_endpoint(get_lib().TrackingServiceDeleteTraceTag, request)

    def search_experiments(
        self,
        view_type=ViewType.ACTIVE_ONLY,
        max_results=SEARCH_MAX_RESULTS_DEFAULT,
        filter_string=None,
        order_by=None,
        page_token=None,
    ):
        request = SearchExperiments(
            view_type=view_type,
            max_results=max_results,
            filter=filter_string,
            order_by=order_by,
            page_token=page_token,
        )
        response = self.service.call_endpoint(get_lib().TrackingServiceSearchExperiments, request)
        experiments = [
            Experiment.from_proto(proto_experiment) for proto_experiment in response.experiments
        ]
        return PagedList(experiments, (response.next_page_token or None))

    def set_tag(self, run_id, tag):
        request = SetTag(run_id=run_id, key=tag.key, value=tag.value)
        self.service.call_endpoint(get_lib().TrackingServiceSetTag, request)

    def start_trace(
        self,
        experiment_id: str,
        timestamp_ms: int,
        request_metadata: Dict[str, str],
        tags: Dict[str, str],
    ) -> TraceInfo:
        request = StartTrace(
            experiment_id=experiment_id,
            timestamp_ms=timestamp_ms,
            request_metadata=[
                TraceRequestMetadata(key=key, value=value)
                for key, value in request_metadata.items()
            ]
            if request_metadata
            else [],
            tags=[TraceTag(key=key, value=value) for key, value in tags.items()]
            if request_metadata
            else [],
        )
        response = self.service.call_endpoint(get_lib().TrackingServiceStartTrace, request)
        return TraceInfo.from_proto(response.trace_info)


def TrackingStore(cls):
    return type(cls.__name__, (_TrackingStore, cls), {})


def _get_sqlalchemy_store(store_uri, artifact_uri):
    from mlflow.store.tracking import DEFAULT_LOCAL_FILE_AND_ARTIFACT_PATH
    from mlflow.store.tracking.sqlalchemy_store import SqlAlchemyStore

    if is_go_enabled():
        SqlAlchemyStore = TrackingStore(SqlAlchemyStore)

    if artifact_uri is None:
        artifact_uri = DEFAULT_LOCAL_FILE_AND_ARTIFACT_PATH

    return SqlAlchemyStore(store_uri, artifact_uri)
