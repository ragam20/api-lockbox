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

type Config struct {
	App AppConfig
}

func NewConfig(env *Env) *Config {
	return &Config{
		App: env.App,
	}
}

func NewAppConfig(cfg *Config) *AppConfig {
	return &cfg.App
}
