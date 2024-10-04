package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	as "github.com/mlflow/mlflow-go/pkg/artifacts/service"
	mr "github.com/mlflow/mlflow-go/pkg/model_registry/service"
	ts "github.com/mlflow/mlflow-go/pkg/tracking/service"

	"github.com/mlflow/mlflow-go/pkg/config"
	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/server/parser"
	"github.com/mlflow/mlflow-go/pkg/server/routes"
	"github.com/mlflow/mlflow-go/pkg/utils"
)

//nolint:funlen
func configureApp(ctx context.Context, cfg *config.Config) (*fiber.App, error) {
	//nolint:mnd
	app := fiber.New(fiber.Config{
		BodyLimit:      16 * 1024 * 1024,
		ReadBufferSize: 16384,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   600 * time.Second,
		IdleTimeout:    120 * time.Second,
		ServerHeader:   "mlflow/" + cfg.Version,
		JSONEncoder: func(v interface{}) ([]byte, error) {
			if protoMessage, ok := v.(proto.Message); ok {
				return protojson.Marshal(protoMessage)
			}

			return json.Marshal(v)
		},
		JSONDecoder: func(data []byte, v interface{}) error {
			if protoMessage, ok := v.(proto.Message); ok {
				return protojson.Unmarshal(data, protoMessage)
			}

			return json.Unmarshal(data, v)
		},
		DisableStartupMessage: true,
	})

	app.Use(compress.New())
	app.Use(recover.New(recover.Config{EnableStackTrace: true}))
	app.Use(logger.New(logger.Config{
		Format: "${status} - ${latency} ${method} ${path}\n",
		Output: utils.GetLoggerFromContext(ctx).Writer(),
	}))
	app.Use(func(c *fiber.Ctx) error {
		c.SetUserContext(ctx)

		return c.Next()
	})

	apiApp, err := newAPIApp(ctx, cfg)
	if err != nil {
		return nil, err
	}

	app.Mount("/api/2.0", apiApp)
	app.Mount("/ajax-api/2.0", apiApp)

	if cfg.StaticFolder != "" {
		app.Static("/static-files", cfg.StaticFolder)
		app.Get("/", func(c *fiber.Ctx) error {
			return c.SendFile(filepath.Join(cfg.StaticFolder, "index.html"))
		})
	}

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})
	app.Get("/version", func(c *fiber.Ctx) error {
		return c.SendString(cfg.Version)
	})

	if cfg.PythonAddress != "" {
		app.Use(proxy.BalancerForward([]string{cfg.PythonAddress}))
	}

	return app, nil
}

func launchServer(ctx context.Context, cfg *config.Config) error {
	logger := utils.GetLoggerFromContext(ctx)

	app, err := configureApp(ctx, cfg)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()

		logger.Info("Shutting down MLflow Go server")

		if err := app.ShutdownWithTimeout(cfg.ShutdownTimeout.Duration); err != nil {
			logger.Errorf("Failed to gracefully shutdown MLflow Go server: %v", err)
		}
	}()

	if cfg.PythonAddress != "" {
		logger.Debugf("Waiting for Python server to be ready on http://%s", cfg.PythonAddress)

		for {
			dialer := &net.Dialer{}
			conn, err := dialer.DialContext(ctx, "tcp", cfg.PythonAddress)

			if err == nil {
				conn.Close()

				break
			}

			if errors.Is(err, context.Canceled) {
				return fmt.Errorf("failed to connect to Python server: %w", err)
			}

			time.Sleep(50 * time.Millisecond) //nolint:mnd
		}
		logger.Debugf("Python server is ready on http://%s", cfg.PythonAddress)
	}

	logger.Infof("Launching MLflow Go server on http://%s", cfg.Address)

	err = app.Listen(cfg.Address)
	if err != nil {
		return fmt.Errorf("failed to start MLflow Go server: %w", err)
	}

	return nil
}

func newFiberConfig() fiber.Config {
	return fiber.Config{
		ErrorHandler: func(context *fiber.Ctx, err error) error {
			var contractError *contract.Error
			if !errors.As(err, &contractError) {
				code := protos.ErrorCode_INTERNAL_ERROR

				var f *fiber.Error
				if errors.As(err, &f) {
					switch f.Code {
					case fiber.StatusBadRequest:
						code = protos.ErrorCode_BAD_REQUEST
					case fiber.StatusServiceUnavailable:
						code = protos.ErrorCode_SERVICE_UNDER_MAINTENANCE
					case fiber.StatusNotFound:
						code = protos.ErrorCode_ENDPOINT_NOT_FOUND
					}
				}

				contractError = contract.NewError(code, err.Error())
			}

			var logFn func(format string, args ...any)

			logger := utils.GetLoggerFromContext(context.Context())
			switch contractError.StatusCode() {
			case fiber.StatusBadRequest:
				logFn = logger.Infof
			case fiber.StatusServiceUnavailable:
				logFn = logger.Warnf
			case fiber.StatusNotFound:
				logFn = logger.Debugf
			default:
				logFn = logger.Errorf
			}

			logFn("Error encountered in %s %s: %s", context.Method(), context.Path(), err)

			return context.Status(contractError.StatusCode()).JSON(contractError)
		},
	}
}

func newAPIApp(ctx context.Context, cfg *config.Config) (*fiber.App, error) {
	app := fiber.New(newFiberConfig())

	parser, err := parser.NewHTTPRequestParser()
	if err != nil {
		return nil, fmt.Errorf("failed to create new HTTP request parser: %w", err)
	}

	trackingService, err := ts.NewTrackingService(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create new tracking service: %w", err)
	}

	routes.RegisterTrackingServiceRoutes(trackingService, parser, app)

	modelRegistryService, err := mr.NewModelRegistryService(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create new model registry service: %w", err)
	}

	routes.RegisterModelRegistryServiceRoutes(modelRegistryService, parser, app)

	artifactService, err := as.NewArtifactsService(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create new artifacts service: %w", err)
	}

	routes.RegisterArtifactsServiceRoutes(artifactService, parser, app)

	return app, nil
}
