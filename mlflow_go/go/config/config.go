package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type Duration struct {
	time.Duration
}

var ErrDuration = errors.New("invalid duration")

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return fmt.Errorf("could not unmarshall duration: %w", err)
	}

	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)

		return nil
	case string:
		var err error

		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("could not parse duration \"%s\": %w", value, err)
		}

		return nil
	default:
		return ErrDuration
	}
}

type Config struct {
	Address               string   `json:"address"`
	DefaultArtifactRoot   string   `json:"default_artifact_root"`
	LogLevel              string   `json:"log_level"`
	PythonAddress         string   `json:"python_address"`
	PythonCommand         []string `json:"python_command"`
	PythonEnv             []string `json:"python_env"`
	ShutdownTimeout       Duration `json:"shutdown_timeout"`
	StaticFolder          string   `json:"static_folder"`
	TrackingStoreURI      string   `json:"tracking_store_uri"`
	ModelRegistryStoreURI string   `json:"model_registry_store_uri"`
	Version               string   `json:"version"`
}
