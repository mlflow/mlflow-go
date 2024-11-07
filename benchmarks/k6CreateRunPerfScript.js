import http from "k6/http";

// Set base url from environment variable, the variable ' -e MLFLOW_TRACKING_URI = xxxx ' must be added to command line argument
const base_url = __ENV.MLFLOW_TRACKING_URI + "/api/2.0/mlflow";

if (!base_url.startsWith("http")) {
    throw new Error("MLFLOW_TRACKING_URI must be a valid URL, starting with http(s)");
}

export function setup() {
    const experiment_response = http.post(
        `${base_url}/experiments/create`,
        JSON.stringify({
            name: `exp_k6_${Date.now()}`,
            tags: [
                {
                    key: "description",
                    value: "k6 experiment",
                },
            ],
        }),
        {
            headers: {
                "Content-Type": "application/json",
            },
        }
    );

    const createExperimentResponse = experiment_response.json();
    const experiment_id = createExperimentResponse.experiment_id;
    return experiment_id;
}

function createRun({ experiment_id }) {
    const run_response = http.post(
        `${base_url}/runs/create`,
        JSON.stringify({
            experiment_id: experiment_id,
            start_time: Date.now(),
            tags: [
                {
                    key: "mlflow.user",
                    value: "k6",
                },
            ],
        }),
        {
            headers: {
                "Content-Type": "application/json",
            },
        }
    );

    const runResponseJson = run_response.json();
    const runInfo = runResponseJson.run.info;
    return runInfo.run_id || runInfo.runId;
}

export default function (experiment_id) {
    createRun({ experiment_id });
}