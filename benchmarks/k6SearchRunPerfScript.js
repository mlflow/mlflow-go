import http from "k6/http";

// Set base url from environment variable, the variable ' -e MLFLOW_TRACKING_URI = xxxx ' must be added to command line argument
const base_url = __ENV.MLFLOW_TRACKING_URI + "/api/2.0/mlflow";

if (!base_url.startsWith("http")) {
    throw new Error("MLFLOW_TRACKING_URI must be a valid URL, starting with http(s)");
}

function searchRuns() {
    // 37 was chosen as experiment number with 23525 runs
    // Found using
    // SELECT experiment_id, count(experiment_id) As count
    // FROM runs
    // GROUP BY experiment_id
    // ORDER BY count DESC;
    const experiment_id = "37";

    const run_response = http.post(
        `${base_url}/runs/search`,
        JSON.stringify({
            experiment_ids: [experiment_id],
            max_results: 1000
        }),
        {
            headers: {
                "Content-Type": "application/json",
            },
        }
    );
}

export default function () {
    searchRuns()
}