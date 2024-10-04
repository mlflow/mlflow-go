package config_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/mlflow/mlflow-go/pkg/config"
)

type validSample struct {
	input    string
	duration config.Duration
}

func TestValidDuration(t *testing.T) {
	t.Parallel()

	samples := []validSample{
		{input: "1000", duration: config.Duration{Duration: 1000 * time.Nanosecond}},
		{input: `"1s"`, duration: config.Duration{Duration: 1 * time.Second}},
		{input: `"2h45m"`, duration: config.Duration{Duration: 2*time.Hour + 45*time.Minute}},
	}

	for _, sample := range samples {
		currentSample := sample
		t.Run(currentSample.input, func(t *testing.T) {
			t.Parallel()

			jsonConfig := fmt.Sprintf(`{ "shutdown_timeout": %s }`, currentSample.input)

			var cfg config.Config

			err := json.Unmarshal([]byte(jsonConfig), &cfg)
			require.NoError(t, err)

			require.Equal(t, currentSample.duration, cfg.ShutdownTimeout)
		})
	}
}

func TestInvalidDuration(t *testing.T) {
	t.Parallel()

	var cfg config.Config

	if err := json.Unmarshal([]byte(`{ "shutdown_timeout": "two seconds" }`), &cfg); err == nil {
		t.Error("expected error")
	}
}
