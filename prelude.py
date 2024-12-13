import tempfile
import os
from pathlib import Path
from mlflow.store.tracking.sqlalchemy_store import SqlAlchemyStore
from mlflow_go.store.tracking import TrackingStore

DB_URI = "sqlite:///"
ARTIFACT_URI = "artifact_folder"

SqlAlchemyStore = TrackingStore(SqlAlchemyStore)

tmp_path = tempfile.gettempdir()
db_file = os.path.join(tmp_path, 'temp.db')
db_uri = f"{DB_URI}{db_file}"
artifact_uri = Path(os.path.join(tmp_path, "artifacts"))
artifact_uri.mkdir(exist_ok=True)
store = SqlAlchemyStore(db_uri, artifact_uri.as_uri())
store._dispose_engine()
del store

db_file = Path(db_file)
if db_file.exists():
    db_file.unlink()
    print(f"{db_file} has been deleted.")