package config

import "time"

type AppConfig struct {
	Env                  string        `mapstructure:"env"`
	Version              string        `mapstructure:"version"`
	Port                 int           `mapstructure:"port"`
	LogLevel             string        `mapstructure:"log_level"`
	NodeID               int64         `mapstructure:"node_id"`
	GlobalRequestTimeout time.Duration `mapstructure:"global_request_timeout"`
	AllowedOrigins       []string      `mapstructure:"allowed_origins"`
	FrontendURL          string        `mapstructure:"frontend_url"`
}

type DatabaseConfig struct {
	URL string `mapstructure:"url"`
}

type JWTConfig struct {
	Secret  string        `mapstructure:"secret"`
	Expired time.Duration `mapstructure:"expired"`
}

type RedisConfig struct {
	URL string `mapstructure:"url"`
}

type SMTPConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
	User string `mapstructure:"user"`
	Pass string `mapstructure:"pass"`
}

type Config struct {
	App         AppConfig
	Database    DatabaseConfig
	JWTConfig   JWTConfig
	RedisConfig RedisConfig
	SMTPConfig  SMTPConfig
}

func NewConfig(env *Env) *Config {
	return &Config{
		App:         env.App,
		Database:    env.Database,
		JWTConfig:   env.JWTConfig,
		RedisConfig: env.RedisConfig,
		SMTPConfig:  env.SMTPConfig,
	}
}

func NewAppConfig(cfg *Config) *AppConfig {
	return &cfg.App
}

func NewDatabaseConfig(cfg *Config) *DatabaseConfig {
	return &cfg.Database
}

func NewJWTConfig(cfg *Config) *JWTConfig {
	return &cfg.JWTConfig
}

func NewRedisConfig(cfg *Config) *RedisConfig {
	return &cfg.RedisConfig
}

func NewSMTPConfig(cfg *Config) *SMTPConfig {
	return &cfg.SMTPConfig
}
