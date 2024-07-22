package utils

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/mlflow/mlflow-go/pkg/config"
)

type loggerKey struct{}

func NewContextWithLogger(ctx context.Context, logger *logrus.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

func GetLoggerFromContext(ctx context.Context) *logrus.Logger {
	logger := ctx.Value(loggerKey{})
	if logger != nil {
		logger, ok := logger.(*logrus.Logger)
		if ok {
			return logger
		}
	}

	return logrus.StandardLogger()
}

func NewLoggerFromConfig(cfg *config.Config) *logrus.Logger {
	logger := logrus.New()

	logLevel, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		logLevel = logrus.InfoLevel
		logger.Warnf("failed to parse log level: %s - assuming %q", err, logrus.InfoLevel)
	}

	logger.SetLevel(logLevel)

	return logger
}
