// Code generated by mlflow/go/cmd/generate/main.go. DO NOT EDIT.

package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mlflow/mlflow-go/pkg/server/parser"
	"github.com/mlflow/mlflow-go/pkg/contract/service"
	"github.com/mlflow/mlflow-go/pkg/utils"
	"github.com/mlflow/mlflow-go/pkg/protos"
)

func RegisterModelRegistryServiceRoutes(service service.ModelRegistryService, parser *parser.HTTPRequestParser, app *fiber.App) {
	app.Post("/mlflow/registered-models/rename", func(ctx *fiber.Ctx) error {
		input := &protos.RenameRegisteredModel{}
		if err := parser.ParseBody(ctx, input); err != nil {
			return err
		}
		output, err := service.RenameRegisteredModel(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
	app.Patch("/mlflow/registered-models/update", func(ctx *fiber.Ctx) error {
		input := &protos.UpdateRegisteredModel{}
		if err := parser.ParseBody(ctx, input); err != nil {
			return err
		}
		output, err := service.UpdateRegisteredModel(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
	app.Delete("/mlflow/registered-models/delete", func(ctx *fiber.Ctx) error {
		input := &protos.DeleteRegisteredModel{}
		if err := parser.ParseBody(ctx, input); err != nil {
			return err
		}
		output, err := service.DeleteRegisteredModel(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
	app.Get("/mlflow/registered-models/get", func(ctx *fiber.Ctx) error {
		input := &protos.GetRegisteredModel{}
		if err := parser.ParseQuery(ctx, input); err != nil {
			return err
		}
		output, err := service.GetRegisteredModel(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
	app.Post("/mlflow/registered-models/get-latest-versions", func(ctx *fiber.Ctx) error {
		input := &protos.GetLatestVersions{}
		if err := parser.ParseBody(ctx, input); err != nil {
			return err
		}
		output, err := service.GetLatestVersions(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
	app.Get("/mlflow/registered-models/get-latest-versions", func(ctx *fiber.Ctx) error {
		input := &protos.GetLatestVersions{}
		if err := parser.ParseQuery(ctx, input); err != nil {
			return err
		}
		output, err := service.GetLatestVersions(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
	app.Delete("/mlflow/model-versions/delete", func(ctx *fiber.Ctx) error {
		input := &protos.DeleteModelVersion{}
		if err := parser.ParseBody(ctx, input); err != nil {
			return err
		}
		output, err := service.DeleteModelVersion(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
	app.Get("/mlflow/model-versions/get", func(ctx *fiber.Ctx) error {
		input := &protos.GetModelVersion{}
		if err := parser.ParseQuery(ctx, input); err != nil {
			return err
		}
		output, err := service.GetModelVersion(utils.NewContextWithLoggerFromFiberContext(ctx), input)
		if err != nil {
			return err
		}
		return ctx.JSON(output)
	})
}
