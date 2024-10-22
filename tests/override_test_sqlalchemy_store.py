import time
from typing import List, Union

import pytest
from mlflow import entities
from mlflow.entities import (
    Param,
)
from mlflow.exceptions import MlflowException
from mlflow.store.tracking.sqlalchemy_store import SqlAlchemyStore
from mlflow.utils.validation import (
    MAX_TAG_VAL_LENGTH,
)


def get_current_time_millis():
    """
    Returns the time in milliseconds since the epoch as an integer number.
    """
    return int(time.time() * 1000)


def _create_experiments(store: SqlAlchemyStore, names) -> Union[str, List]:
    if isinstance(names, (list, tuple)):
        ids = []
        for name in names:
            # Sleep to ensure each experiment has a unique creation_time for
            # deterministic experiment search results
            time.sleep(0.001)
            ids.append(store.create_experiment(name=name))
        return ids

    time.sleep(0.001)
    return store.create_experiment(name=names)


def _run_factory(store: SqlAlchemyStore, config=None):
    if not config:
        config = _get_run_configs()
    if not config.get("experiment_id", None):
        config["experiment_id"] = _create_experiments(store, "test exp")

    return store.create_run(**config)


def _get_run_configs(experiment_id=None, tags=None, start_time=None):
    return {
        "experiment_id": experiment_id,
        "user_id": "Anderson",
        "start_time": get_current_time_millis() if start_time is None else start_time,
        "tags": tags,
        "run_name": "name",
    }


def _verify_logged(store, run_id, metrics, params, tags):
    run = store.get_run(run_id)
    all_metrics = sum([store.get_metric_history(run_id, key) for key in run.data.metrics], [])
    assert len(all_metrics) == len(metrics)
    logged_metrics = [(m.key, m.value, m.timestamp, m.step) for m in all_metrics]
    assert set(logged_metrics) == {(m.key, m.value, m.timestamp, m.step) for m in metrics}
    logged_tags = set(run.data.tags.items())
    assert {(tag.key, tag.value) for tag in tags} <= logged_tags
    assert len(run.data.params) == len(params)
    assert set(run.data.params.items()) == {(param.key, param.value) for param in params}


def test_log_batch_internal_error(store: SqlAlchemyStore):
    # We can't mock SQL storage on the flight to simulate Internal Error.
    # Skip such a test for now.
    ()


def test_log_batch_params_max_length_value(store: SqlAlchemyStore, monkeypatch):
    run = _run_factory(store)
    param_entities = [Param("long param", "x" * 6000), Param("short param", "xyz")]
    expected_param_entities = [
        Param("long param", "x" * 6000),
        Param("short param", "xyz"),
    ]
    store.log_batch(run.info.run_id, [], param_entities, [])
    _verify_logged(store, run.info.run_id, [], expected_param_entities, [])

    # We can't overriride on the fligh MLFLOW_TRUNCATE_LONG_VALUES ENV parameter.
    # Skip such a part of the test.
    # param_entities = [Param("long param", "x" * 6001)]
    # monkeypatch.setenv("MLFLOW_TRUNCATE_LONG_VALUES", "false")
    # with pytest.raises(MlflowException, match="exceeds the maximum length"):
    #     store.log_batch(run.info.run_id, [], param_entities, [])

    monkeypatch.setenv("MLFLOW_TRUNCATE_LONG_VALUES", "true")
    store.log_batch(run.info.run_id, [], param_entities, [])


def test_log_param_max_length_value(store: SqlAlchemyStore, monkeypatch):
    run = _run_factory(store)
    tkey = "blahmetric"
    tval = "x" * 6000
    param = entities.Param(tkey, tval)
    store.log_param(run.info.run_id, param)
    run = store.get_run(run.info.run_id)
    assert run.data.params[tkey] == str(tval)
    # We can't overriride on the fligh MLFLOW_TRUNCATE_LONG_VALUES ENV parameter.
    # Skip such a part of the test.
    # monkeypatch.setenv("MLFLOW_TRUNCATE_LONG_VALUES", "false")
    # with pytest.raises(MlflowException, match="exceeds the maximum length"):
    #     store.log_param(run.info.run_id, entities.Param(tkey, "x" * 6001))

    monkeypatch.setenv("MLFLOW_TRUNCATE_LONG_VALUES", "true")
    store.log_param(run.info.run_id, entities.Param(tkey, "x" * 6001))


def test_set_tag(store: SqlAlchemyStore, monkeypatch):
    run = _run_factory(store)

    tkey = "test tag"
    tval = "a boogie"
    new_val = "new val"
    tag = entities.RunTag(tkey, tval)
    new_tag = entities.RunTag(tkey, new_val)
    store.set_tag(run.info.run_id, tag)
    # Overwriting tags is allowed
    store.set_tag(run.info.run_id, new_tag)

    # We can't overriride on the fligh MLFLOW_TRUNCATE_LONG_VALUES ENV parameter.
    # Skip such a part of the test.
    # test setting tags that are too long fails.
    # monkeypatch.setenv("MLFLOW_TRUNCATE_LONG_VALUES", "false")
    # with pytest.raises(
    #     MlflowException, match=f"exceeds the maximum length of {MAX_TAG_VAL_LENGTH} characters"
    # ):
    #     store.set_tag(
    #         run.info.run_id, entities.RunTag("longTagKey", "a" * (MAX_TAG_VAL_LENGTH + 1))
    #     )

    monkeypatch.setenv("MLFLOW_TRUNCATE_LONG_VALUES", "true")
    store.set_tag(run.info.run_id, entities.RunTag("longTagKey", "a" * (MAX_TAG_VAL_LENGTH + 1)))

    # test can set tags that are somewhat long
    store.set_tag(run.info.run_id, entities.RunTag("longTagKey", "a" * (MAX_TAG_VAL_LENGTH - 1)))
    run = store.get_run(run.info.run_id)
    assert tkey in run.data.tags
    assert run.data.tags[tkey] == new_val


def test_log_batch_null_metrics(store: SqlAlchemyStore):
    # We can't simulate such a situation. Skip such a test for now.
    ()


def test_sqlalchemy_store_behaves_as_expected_with_inmemory_sqlite_db(monkeypatch):
    # We can't simulate such a situation. Skip such a test for now.
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
