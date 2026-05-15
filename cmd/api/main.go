package main

import (
	"net/http"

	"lockbox/internal/app/auth"
	"lockbox/internal/cache"
	"lockbox/internal/config"
	"lockbox/internal/database"
	"lockbox/internal/database/repository"
	appHttp "lockbox/internal/http"
	"lockbox/internal/http/handler"
	"lockbox/internal/http/middleware"
	"lockbox/internal/http/router"
	"lockbox/internal/http/server"
	"lockbox/internal/idgen"
	"lockbox/internal/logger"
	"lockbox/internal/smtp"

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
			config.NewRedisConfig,
			config.NewSMTPConfig,
		),
	)

	infraModule := fx.Module("infra",
		fx.Provide(
			idgen.NewSnowflakeGen,
			logger.NewLogger,
		),
	)

	databaseModule := fx.Module("database",
		fx.Provide(
			database.NewPostgresConnection,
			cache.NewRedisClient,
			cache.NewStore,
			smtp.NewSMTPConfig,
			smtp.NewMailService,
		),
	)

	authModule := fx.Module("auth",
		fx.Provide(
			repository.NewAuthRepository,
			auth.NewAuthService,
			handler.NewHandler,
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
		authModule,

		fx.Invoke(
			func(_ *http.Server, _ *appHttp.CookieManager, _ *gorm.DB) {},
			router.NewRoutes,
		),
	))

	app.Run()
}
