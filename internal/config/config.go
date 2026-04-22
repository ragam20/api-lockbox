package config

import "time"

type AppConfig struct {
	Env            string        `mapstructure:"env"`
	Version        string        `mapstructure:"version"`
	Port           int           `mapstructure:"port"`
	LogLevel       string        `mapstructure:"log_level"`
	RequestTimeout time.Duration `mapstructure:"request_timeout"`
	AllowedOrigins []string      `mapstructure:"allowed_origins"`
}

type DatabaseConfig struct {
	URL string `mapstructure:"url"`
}

type Config struct {
	App      AppConfig
	Database DatabaseConfig
}

func NewConfig(env *Env) *Config {
	return &Config{
		App:      env.App,
		Database: env.Database,
	}
}

func NewAppConfig(cfg *Config) *AppConfig {
	return &cfg.App
}

func NewDatabaseConfig(cfg *Config) *DatabaseConfig {
	return &cfg.Database
}
