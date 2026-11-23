package main

import (
	"ftools/config"
	"ftools/internal/controller"
	"ftools/internal/controller/web"
	"ftools/internal/service"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {

	fx.New(
		fx.Provide(zap.NewProduction),
		config.Module,
		service.Module,
		controller.Module,
		web.Module,
	).Run()
}
