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
		return fmt.Errorf("failed to unmarshal duration: %w", err)
	}

	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)

		return nil
	case string:
		var err error

		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("failed to parse duration \"%s\": %w", value, err)
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
	ModelRegistryStoreURI string   `json:"model_registry_store_uri"`
	PythonAddress         string   `json:"python_address"`
	PythonCommand         []string `json:"python_command"`
	ShutdownTimeout       Duration `json:"shutdown_timeout"`
	StaticFolder          string   `json:"static_folder"`
	TrackingStoreURI      string   `json:"tracking_store_uri"`
	Version               string   `json:"version"`
}

func NewConfigFromBytes(cfgBytes []byte) (*Config, error) {
	if len(cfgBytes) == 0 {
		cfgBytes = []byte("{}")
	}

	var cfg Config
	if err := json.Unmarshal(cfgBytes, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse JSON config: %w", err)
	}

	return &cfg, nil
}

func NewConfigFromString(s string) (*Config, error) {
	return NewConfigFromBytes([]byte(s))
}
