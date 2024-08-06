import json

from google.protobuf.message import DecodeError
from mlflow.exceptions import MlflowException
from mlflow.protos.databricks_pb2 import INTERNAL_ERROR, ErrorCode

from mlflow_go.lib import get_ffi, get_lib


class _ServiceProxy:
    def __init__(self, id):
        self.id = id

    def call_endpoint(self, endpoint, request):
        request_data = request.SerializeToString()
        response_size = get_ffi().new("int*")

        response_data = endpoint(
            self.id,
            request_data,
            len(request_data),
            response_size,
        )

        response_bytes = get_ffi().buffer(response_data, response_size[0])[:]
        get_lib().free(response_data)

        try:
            response = type(request).Response()
            response.ParseFromString(response_bytes)
            return response
        except DecodeError:
            try:
                e = json.loads(response_bytes)
                error_code = e.get("error_code", ErrorCode.Name(INTERNAL_ERROR))
                raise MlflowException(
                    message=e["message"],
                    error_code=ErrorCode.Value(error_code),
                ) from None
            except json.JSONDecodeError as e:
                raise MlflowException(
                    message=f"Failed to parse response: {e}",
                )
