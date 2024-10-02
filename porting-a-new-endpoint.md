# Porting a New Endpoint

As mentioned elsewhere, the Go implementation of the tracking server is currently incomplete. This guide outlines how to implement a missing endpoint.

## View Current Endpoint Status

Run `mage endpoints` to view which endpoints are not yet implemented.

In this guide, we will implement `deleteTag`.

```
+------------------------+------------------------------+-------------+
|        SERVICE         |           ENDPOINT           | IMPLEMENTED |
+------------------------+------------------------------+-------------+
...
+------------------------+------------------------------+-------------+
| MlflowService          | deleteTag                    |     No      |
+------------------------+------------------------------+-------------+
```

## Enable the Endpoint for Generation

Add the missing endpoint to [endpoints.go](magefiles/generate/endpoints.go) under the correct service. After that, run `mage generate`.

This should update the following files:

### Add a New Route

In [pkg/server/routes/tracking.g.go](pkg/server/routes/tracking.g.go), add the following:

```go
	app.Post("/mlflow/runs/delete-tag", func(ctx *fiber.Ctx) error {
		input := &protos.DeleteTag{}
		if err := parser.ParseBody(ctx, input); err != nil {
			return err
		}
		output, err := service.DeleteTag(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
```

This means that our Go server will now process the API request.

### New Service Method

In [pkg/contract/service/tracking.g.go](pkg/contract/service/tracking.g.go), the interface has a new function that we need to implement, which is used in the routes file above.

```go
type TrackingService interface {
    ...
	DeleteTag(ctx context.Context, input *protos.DeleteTag) (*protos.DeleteTag_Response, *contract.Error)
}
```

### New FFI Endpoint

In [pkg/lib/tracking.g.go](pkg/lib/tracking.g.go), we need to expose that same service method so it can be called directly from Python via FFI. More on that later.

```go
//export TrackingServiceDeleteTag
func TrackingServiceDeleteTag(serviceID int64, requestData unsafe.Pointer, requestSize C.int, responseSize *C.int) unsafe.Pointer {
	service, err := trackingServices.Get(serviceID)
	if err != nil {
		return makePointerFromError(err, responseSize)
	}
	return invokeServiceMethod(service.DeleteTag, new(protos.DeleteTag), requestData, requestSize, responseSize)
}
```

## Input Validation

A first step to port an endpoint would be to check the request validation happening at the HTTP level. Open [handlers.py](.mlflow.repo/mlflow/server/handlers.py) (from the `.mlflow.repo`) to see which fields are required.

```python
@catch_mlflow_exception
@_disable_if_artifacts_only
def _delete_tag():
    request_message = _get_request_message(
        DeleteTag(),
        schema={
            "run_id": [_assert_required, _assert_string],
            "key": [_assert_required, _assert_string],
        },
    )
    _get_tracking_store().delete_tag(request_message.run_id, request_message.key)
    response_message = DeleteTag.Response()
    response = Response(mimetype="application/json")
    response.set_data(message_to_json(response_message))
    return response
```

We notice that `run_id` and `key` are required. To ensure parity, we need to update the input struct `DeleteTag` from the generated proto code ([pkg/protos/service.pb.go](pkg/protos/service.pb.go)). ⚠️ Since this file was generated, we don't want to modify it directly. Instead, we will configure validations in [validations.go](magefiles/generate/validations.go).

```go
var validations = map[string]string{
    ...
	"DeleteTag_RunId":                    "required",
	"DeleteTag_Key":                      "required",
}
```

After that, run `mage generate` again and check if our fields in `DeleteTag` now contain `validate:"required"`.

## Service Implementation

We aim to keep the Go implementation as close to the Python code as possible. That's why we have a thin service layer and keep most logic in the (SQL) store instead of the service. This keeps things similar to Python, making it easier to port and compare the code.

In the service, we need to call the store with the same arguments as the Python store would be called. This is typically where we convert proto structs to entities. In the case of `deleteTag`, we can just pass the two arguments to the store.

```go
func (ts TrackingService) DeleteTag(ctx context.Context, input *protos.DeleteTag) (*protos.DeleteTag_Response, *contract.Error) {
    // Note that Store.DeleteTag does not exist yet; this just mirrors what happens in _delete_tag() in handlers.py
	if err := ts.Store.DeleteTag(ctx, input.GetRunId(), input.GetKey()); err != nil {
		return nil, err
	}

	return &protos.DeleteTag_Response{}, nil
}
```

## Store Implementation

In the future, we hope to implement the file store in Go as well. That’s why we use a store interface to abstract away all SQL details, which also makes testing easier.

Update [store.go](pkg/tracking/store/store.go):

```go
	RunTrackingStore interface {
		...
		DeleteTag(ctx context.Context, runID, key string) *contract.Error
	}
```

Then, add an empty implementation in [store/sql/tags.go](pkg/tracking/store/sql/tags.go):

```go
func (s TrackingSQLStore) DeleteTag(
	ctx context.Context, runID string, key string,
) *contract.Error {
	return nil
}
```

Lastly, update the `TrackingStore` interface mock via `go generate ./...`.

If everything goes well, you should be able to run `mage test:unit` to execute the Go unit tests. This serves as a sanity check to verify that the Go code is compiling.

### The Actual Store Implementation

At this point, we need to write the actual implementation. Again, we want to stay true to the Python code, so please refer to [sqlalchemy_store.py](.mlflow.repo/mlflow/store/tracking/sqlalchemy_store.py).

It is important that the same data and exceptions are returned to ensure the Go implementation behaves like the Python one, allowing users to seamlessly switch to Go while having the same experience.

Don’t hesitate to challenge the current implementation of the Python code. You might find opportunities to revisit and improve parts of it.

## Create the Python Binding

To access our new endpoint in the Python FFI binding, we need to update our Python store as well. Update [mlflow_go/store/tracking.py](mlflow_go/store/tracking.py) and add the method matching our endpoint from [.mlflow.repo/mlflow/store/tracking/sqlalchemy_store.py](.mlflow.repo/mlflow/store/tracking/sqlalchemy_store.py):

```python
def delete_tag(self, run_id, key):
	request = DeleteTag(run_id=run_id, key=key)
	self.service.call_endpoint(get_lib().TrackingServiceDeleteTag, request)
```

Note that `DeleteTag` and `TrackingServiceDeleteTag` will match our newly generated code in [pkg/lib/tracking.g.go](pkg/lib/tracking.g.go).

## Testing

Depending on the endpoint you are porting, you may want to add Go unit tests or rely on the existing Python integration tests. It really depends on the complexity of the endpoint. 

As a rule of thumb, add Go unit tests for clear-cut isolated parts, like validation logic, and keep integration tests for the actual database layer. Avoid tests that are too closely coupled to the implementation. It’s sometimes acceptable not to have any Go unit tests if the existing Python tests adequately cover the endpoint.

## Run the Python Integration Test

Run `mage test:python` and verify that our Go implementation passes the existing tests.

There is one caveat to these tests; occasionally, they may have a Python bias, meaning that the tests pass due to Python's dynamic nature, while our Go tests might fail because they are strongly typed. Another issue may arise if the Python implementation does not consistently return the same error messages. Therefore, it may be necessary to submit a PR to [mlflow](https://github.com/mlflow/mlflow) to adjust the existing tests.

Some examples include:
- https://github.com/mlflow/mlflow/pull/13233
- https://github.com/mlflow/mlflow/pull/13128
- https://github.com/mlflow/mlflow/issues/12550