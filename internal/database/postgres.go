package database

import (
	"lockbox/internal/config"
	"lockbox/internal/database/entity"
	appLogger "lockbox/internal/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewPostgresConnection(cfg *config.DatabaseConfig, appCfg *config.AppConfig, log appLogger.Logger) (*gorm.DB, error) {
	gormCfg := &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Warn),
		TranslateError: true,
	}

	pgConfig := postgres.Config{
		DSN: cfg.URL,
	}

	if appCfg.Env == "production" {
		gormCfg.PrepareStmt = false
		pgConfig.PreferSimpleProtocol = true
	}

	db, err := gorm.Open(postgres.New(pgConfig), gormCfg)
	if err != nil {
		log.Error("failed to connect database", err, map[string]any{})
		return nil, err
	}

	if appCfg.Env == "development" {
		if err := db.AutoMigrate(
			entity.UserEntity{},
			entity.ArgonConfigEntity{},
			entity.VaultEntity{},
			entity.VaultItemEntity{},
			entity.UserSessionEntity{},
		); err != nil {
			log.Error("failed to migrate", err, map[string]any{})
			return nil, err
		}
	}

	return db, nil
}
