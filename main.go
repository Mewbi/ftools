package main

import (
	"ftools/config"
	"ftools/internal/controller"
	"ftools/internal/controller/web"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {

	fx.New(
		fx.Provide(zap.NewProduction),
		config.Module,
		controller.Module,
		web.Module,
	).Run()
}
