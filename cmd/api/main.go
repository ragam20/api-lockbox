package main

import (
	"net/http"

	"lockbox/internal/config"
	appHttp "lockbox/internal/http"
	"lockbox/internal/http/middleware"
	"lockbox/internal/http/router"
	"lockbox/internal/http/server"
	"lockbox/internal/logger"

	"go.uber.org/fx"
)

func main() {
	configModule := fx.Module("config",
		fx.Provide(
			config.NewEnv,
			config.NewConfig,
			config.NewAppConfig,
		),
	)

	infraModule := fx.Module("infra",
		fx.Provide(
			logger.NewLogger,
		),
	)

	appHTTPModule := fx.Module("apphttp",
		fx.Provide(
			middleware.NewRequestTimeoutMiddleware,
			middleware.NewErrorMiddleware,
			middleware.NewRecoveryMiddleware,

			server.NewGinEngine,
			server.NewHTTPServer,

			appHttp.NewCookieManager,
		),
	)

	app := fx.New(fx.Options(
		configModule,
		infraModule,
		appHTTPModule,

		fx.Invoke(
			func(_ *http.Server, _ *appHttp.CookieManager) {},
			router.NewRoutes,
		),
	))

	app.Run()
}
