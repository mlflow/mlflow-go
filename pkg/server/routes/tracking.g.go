// Code generated by mlflow/go/cmd/generate/main.go. DO NOT EDIT.

package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mlflow/mlflow-go/pkg/server/parser"
	"github.com/mlflow/mlflow-go/pkg/contract/service"
	"github.com/mlflow/mlflow-go/pkg/utils"
	"github.com/mlflow/mlflow-go/pkg/protos"
)

func RegisterTrackingServiceRoutes(service service.TrackingService, parser *parser.HTTPRequestParser, app *fiber.App) {
	app.Get("/mlflow/experiments/get-by-name", func(ctx *fiber.Ctx) error {
		input := &protos.GetExperimentByName{}
		if err := parser.ParseQuery(ctx, input); err != nil {
			return err
		}
		output, err := service.GetExperimentByName(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
	app.Post("/mlflow/experiments/create", func(ctx *fiber.Ctx) error {
		input := &protos.CreateExperiment{}
		if err := parser.ParseBody(ctx, input); err != nil {
			return err
		}
		output, err := service.CreateExperiment(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
	app.Post("/mlflow/experiments/search", func(ctx *fiber.Ctx) error {
		input := &protos.SearchExperiments{}
		if err := parser.ParseBody(ctx, input); err != nil {
			return err
		}
		output, err := service.SearchExperiments(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
	app.Get("/mlflow/experiments/search", func(ctx *fiber.Ctx) error {
		input := &protos.SearchExperiments{}
		if err := parser.ParseQuery(ctx, input); err != nil {
			return err
		}
		output, err := service.SearchExperiments(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
	app.Get("/mlflow/experiments/get", func(ctx *fiber.Ctx) error {
		input := &protos.GetExperiment{}
		if err := parser.ParseQuery(ctx, input); err != nil {
			return err
		}
		output, err := service.GetExperiment(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
	app.Post("/mlflow/experiments/delete", func(ctx *fiber.Ctx) error {
		input := &protos.DeleteExperiment{}
		if err := parser.ParseBody(ctx, input); err != nil {
			return err
		}
		output, err := service.DeleteExperiment(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
	app.Post("/mlflow/experiments/restore", func(ctx *fiber.Ctx) error {
		input := &protos.RestoreExperiment{}
		if err := parser.ParseBody(ctx, input); err != nil {
			return err
		}
		output, err := service.RestoreExperiment(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
	app.Post("/mlflow/experiments/update", func(ctx *fiber.Ctx) error {
		input := &protos.UpdateExperiment{}
		if err := parser.ParseBody(ctx, input); err != nil {
			return err
		}
		output, err := service.UpdateExperiment(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
	app.Post("/mlflow/runs/create", func(ctx *fiber.Ctx) error {
		input := &protos.CreateRun{}
		if err := parser.ParseBody(ctx, input); err != nil {
			return err
		}
		output, err := service.CreateRun(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
	app.Post("/mlflow/runs/update", func(ctx *fiber.Ctx) error {
		input := &protos.UpdateRun{}
		if err := parser.ParseBody(ctx, input); err != nil {
			return err
		}
		output, err := service.UpdateRun(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
	app.Post("/mlflow/runs/delete", func(ctx *fiber.Ctx) error {
		input := &protos.DeleteRun{}
		if err := parser.ParseBody(ctx, input); err != nil {
			return err
		}
		output, err := service.DeleteRun(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
	app.Post("/mlflow/runs/restore", func(ctx *fiber.Ctx) error {
		input := &protos.RestoreRun{}
		if err := parser.ParseBody(ctx, input); err != nil {
			return err
		}
		output, err := service.RestoreRun(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
	app.Post("/mlflow/runs/log-metric", func(ctx *fiber.Ctx) error {
		input := &protos.LogMetric{}
		if err := parser.ParseBody(ctx, input); err != nil {
			return err
		}
		output, err := service.LogMetric(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
	app.Post("/mlflow/runs/log-parameter", func(ctx *fiber.Ctx) error {
		input := &protos.LogParam{}
		if err := parser.ParseBody(ctx, input); err != nil {
			return err
		}
		output, err := service.LogParam(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
	app.Post("/mlflow/experiments/set-experiment-tag", func(ctx *fiber.Ctx) error {
		input := &protos.SetExperimentTag{}
		if err := parser.ParseBody(ctx, input); err != nil {
			return err
		}
		output, err := service.SetExperimentTag(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
	app.Patch("/mlflow/traces/:request_id/tags", func(ctx *fiber.Ctx) error {
		input := &protos.SetTraceTag{}
		if err := parser.ParseBody(ctx, input); err != nil {
			return err
		}
		output, err := service.SetTraceTag(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
	app.Delete("/mlflow/traces/:request_id/tags", func(ctx *fiber.Ctx) error {
		input := &protos.DeleteTraceTag{}
		if err := parser.ParseBody(ctx, input); err != nil {
			return err
		}
		output, err := service.DeleteTraceTag(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
	app.Post("/mlflow/runs/delete-tag", func(ctx *fiber.Ctx) error {
		input := &protos.DeleteTag{}
		if err := parser.ParseBody(ctx, input); err != nil {
			return err
		}
		output, err := service.DeleteTag(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
	app.Get("/mlflow/runs/get", func(ctx *fiber.Ctx) error {
		input := &protos.GetRun{}
		if err := parser.ParseQuery(ctx, input); err != nil {
			return err
		}
		output, err := service.GetRun(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
	app.Post("/mlflow/runs/search", func(ctx *fiber.Ctx) error {
		input := &protos.SearchRuns{}
		if err := parser.ParseBody(ctx, input); err != nil {
			return err
		}
		output, err := service.SearchRuns(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
	app.Post("/mlflow/runs/log-batch", func(ctx *fiber.Ctx) error {
		input := &protos.LogBatch{}
		if err := parser.ParseBody(ctx, input); err != nil {
			return err
		}
		output, err := service.LogBatch(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
	app.Post("/mlflow/runs/log-inputs", func(ctx *fiber.Ctx) error {
		input := &protos.LogInputs{}
		if err := parser.ParseBody(ctx, input); err != nil {
			return err
		}
		output, err := service.LogInputs(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
}
