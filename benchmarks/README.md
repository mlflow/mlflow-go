# Benchmarks

The primary goal of this Go implementation of the MLFlow tracking server is to enhance performance. We conducted initial benchmarks using `k6` for the following operations:

- SearchRuns
- SearchExperiments
- CreateRun
- LogMetricBatch10
- LogMetricBatch500
- LogParamBatch10
- LogParamBatch100

Please note that in the LogMetricBatch500, issue [#81](https://github.com/mlflow/mlflow-go/issues/81) was identified and therefore excluded from the results.

## Approach

We start the MLflow server with the command:

```shell
uv run mlflow -- server --backend-store-uri "postgresql://postgres:postgres@localhost:5432/postgres"
```

We then start MLflow-Go using:

```shell
mage dev
```

(Note: `MLFLOW_SQLALCHEMYSTORE_ECHO=0` and `log_level=fatal` were adjusted to minimize logging impact.)

Next, we initiate each test with the following command:

```shell
k6 run -e MLFLOW_TRACKING_URI=http://localhost:5000 k6LogBatchPerfScript.js -u 20 -d 30s
```

Benchmark results were recorded manually:

| Test Case                 | MLflow Avg Duration (ms) | MLflow Iterations | MLflow-Go Avg Duration (ms)  | MLflow-Go Iterations |
|---------------------------|--------------------------|-------------------|------------------------------|----------------------|
| LogMetricBatch10          | 58.43                    | 10224             | 18.19                        | 32827                |
| LogMetricBatch500         | 1200                     | 504               | 13.12                        | 56                   |
| LogParamBatch10           | 12.76                    | 46346             | 7.97                         | 74127                |
| LogParamBatch100          | 19.85                    | 29635             | 12.02                        | 47940                |
| CreateRun                 | 25.33                    | 23525             | 13.71                        | 43431                |
| SearchRuns                | 1004                     | 582               | 104.42                       | 5753                 |
| SearchExperiments         | 20.85                    | 28583             | 8.88                         | 67065                |

![Duration Results](results_duration.png)
![Iteration Results](results_iteration.png)

We welcome contributions to automate this process!