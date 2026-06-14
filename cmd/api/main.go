package main

import (
	"log"
	"net/http"

	"lockbox/internal/app/auth"
	"lockbox/internal/app/tools"
	"lockbox/internal/app/user"
	"lockbox/internal/app/vault"
	"lockbox/internal/cache"
	"lockbox/internal/config"
	"lockbox/internal/database"
	"lockbox/internal/database/repository"
	dbtx "lockbox/internal/database/transaction"
	appHttp "lockbox/internal/http"
	"lockbox/internal/http/handler"
	"lockbox/internal/http/middleware"
	"lockbox/internal/http/router"
	"lockbox/internal/http/server"
	"lockbox/internal/idgen"
	"lockbox/internal/logger"
	"lockbox/internal/smtp"
	"lockbox/internal/storage/cloudinary"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

func main() {
	configModule := fx.Module(
		"config",
		fx.Provide(
			config.NewEnv,
			config.NewConfig,
			config.NewAppConfig,
			config.NewDatabaseConfig,
			config.NewRedisConfig,
			config.NewSMTPConfig,
			config.NewJWTConfig,
			config.NewStorageConfig,
			config.NewCloudinaryConfig,
		),
	)

	infraModule := fx.Module(
		"infra",
		fx.Provide(
			idgen.NewIDGen,
			logger.NewLogger,
			smtp.NewSMTPConfig,
			smtp.NewMailService,
			cloudinary.NewCloudinaryService,
		),
	)

	databaseModule := fx.Module(
		"database",
		fx.Provide(
			database.NewPostgresConnection,
			dbtx.NewTransactionManager,
			cache.NewRedisClient,
			cache.NewStore,
		),
	)

	repositoryModule := fx.Module(
		"repository",
		fx.Provide(
			repository.NewAuthRepository,
			repository.NewUserRepository,
			repository.NewVaultRepository,
		),
	)

	authModule := fx.Module(
		"auth",
		fx.Provide(
			auth.NewTokenManager,
			auth.NewAuthService,
			handler.NewAuthHandler,
		),
	)

	userModule := fx.Module(
		"user",
		fx.Provide(
			user.NewUserService,
			handler.NewUserHandler,
		),
	)

	vaultModule := fx.Module(
		"vault",
		fx.Provide(
			vault.NewVaultService,
			tools.NewToolService,
			handler.NewVaultHandler,
		),
	)

	appHTTPModule := fx.Module(
		"apphttp",
		fx.Provide(
			middleware.NewErrorMiddleware,
			middleware.NewRecoveryMiddleware,
			middleware.NewVerifyJWTMiddleware,
			middleware.NewRequestTimeoutMiddleware,
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
		repositoryModule,
		authModule,
		userModule,
		vaultModule,

		fx.Invoke(
			func(_ *http.Server, _ *appHttp.CookieManager, _ *gorm.DB) {},
			router.NewRoutes,
		),
	))

	if err := app.Err(); err != nil {
		log.Fatal(err)
	}

	app.Run()
}
