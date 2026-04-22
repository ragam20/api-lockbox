package main

import (
	"net/http"

	"lockbox/internal/config"
	"lockbox/internal/database"
	appHttp "lockbox/internal/http"
	"lockbox/internal/http/middleware"
	"lockbox/internal/http/router"
	"lockbox/internal/http/server"
	"lockbox/internal/idgen"
	"lockbox/internal/logger"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

func main() {
	configModule := fx.Module("config",
		fx.Provide(
			config.NewEnv,
			config.NewConfig,
			config.NewAppConfig,
			config.NewDatabaseConfig,
		),
	)

	databaseModule := fx.Module("database",
		fx.Provide(
			database.NewPostgresConnection,
		),
	)

	infraModule := fx.Module("infra",
		fx.Provide(
			idgen.NewSnowflakeGen,
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
		databaseModule,
		infraModule,
		appHTTPModule,

		fx.Invoke(
			func(_ *http.Server, _ *appHttp.CookieManager, _ *gorm.DB) {},
			router.NewRoutes,
		),
	))

	app.Run()
}
