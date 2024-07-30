// Code generated by mlflow/go/cmd/generate/main.go. DO NOT EDIT.

package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mlflow/mlflow-go/pkg/contract/service"
	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/server/parser"
)

func RegisterTrackingServiceRoutes(service service.TrackingService, parser *parser.HTTPRequestParser, app *fiber.App) {
	app.Get("/mlflow/experiments/get-by-name", func(ctx *fiber.Ctx) error {
		input := &protos.GetExperimentByName{}
		if err := parser.ParseQuery(ctx, input); err != nil {
			return err
		}
		output, err := service.GetExperimentByName(input)
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
		output, err := service.CreateExperiment(input)
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
		output, err := service.GetExperiment(input)
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
		output, err := service.DeleteExperiment(input)
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
		output, err := service.RestoreExperiment(input)
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
		output, err := service.CreateRun(input)
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
		output, err := service.SearchRuns(input)
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
		output, err := service.LogBatch(input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
}
