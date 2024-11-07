import http from "k6/http";

// Set base url from environment variable, the variable ' -e MLFLOW_TRACKING_URI = xxxx ' must be added to command line argument
const base_url = __ENV.MLFLOW_TRACKING_URI + "/api/2.0/mlflow";

if (!base_url.startsWith("http")) {
    throw new Error("MLFLOW_TRACKING_URI must be a valid URL, starting with http(s)");
}

function searchExperiments() {
    const response = http.post(
        `${base_url}/experiments/search`,
        JSON.stringify({
            max_results: 1000,
        }),
        {
            headers: {
                "Content-Type": "application/json",
            },
        }
    );
}

export default function () {
    searchExperiments()
}