package utils

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"

	"github.com/mlflow/mlflow-go/pkg/config"
)

type loggerKey struct{}

func NewContextWithLogger(ctx context.Context, logger *logrus.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// NewContextWithLoggerFromFiberContext transfer logger from Fiber context to a normal context.Context object.
func NewContextWithLoggerFromFiberContext(c *fiber.Ctx) context.Context {
	logger := GetLoggerFromContext(c.UserContext())

	return NewContextWithLogger(c.Context(), logger)
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
