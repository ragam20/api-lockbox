package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/spf13/viper"
)

type Env struct {
	App           AppConfig      `mapstructure:"app"`
	Database      DatabaseConfig `mapstructure:"db"`
	JWTConfig     JWTConfig      `mapstructure:"jwt"`
	RedisConfig   RedisConfig    `mapstructure:"redis"`
	SMTPConfig    SMTPConfig     `mapstructure:"smtp"`
	StorageConfig StorageConfig  `mapstructure:"storage"`
}

func NewEnv() (*Env, error) {
	env, err := loadEnv(getKeys())
	if err != nil {
		return nil, err
	}

	fmt.Println("smtp.tls:", env.SMTPConfig.TLS)
	fmt.Println("redis.tls:", env.RedisConfig.TLS)

	if err := env.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	logConfigLoad(env)

	return env, nil
}

func NewMigrationEnv() (*Env, error) {
	env, err := loadEnv(getMigrationKeys())
	if err != nil {
		return nil, err
	}

	if err := env.ValidateMigration(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return env, nil
}

func (e *Env) ValidateMigration() error {
	var errs []string

	validateMigrationAppConfig(e.App, &errs)
	validateDatabaseConfig(e.Database, &errs)

	if len(errs) == 0 {
		return nil
	}

	return fmt.Errorf("invalid config: %s", strings.Join(errs, "; "))
}

func validateMigrationAppConfig(cfg AppConfig, errs *[]string) {
	if strings.TrimSpace(cfg.Env) == "" {
		*errs = append(*errs, "app.env is required")
	}

	if strings.TrimSpace(cfg.LogLevel) == "" {
		*errs = append(*errs, "app.log_level is required")
	} else if _, err := zerolog.ParseLevel(cfg.LogLevel); err != nil {
		*errs = append(*errs, "app.log_level is invalid")
	}
}

func loadEnv(keys []string) (*Env, error) {
	appEnv := normalizeAppEnv(getAppEnv())

	v := viper.New()

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	if !isProduction(appEnv) {
		v.SetConfigName("development")
		v.SetConfigType("yaml")
		v.AddConfigPath("internal/config")
		v.AddConfigPath(".")

		if err := v.ReadInConfig(); err != nil {
			var notFound viper.ConfigFileNotFoundError

			if !errors.As(err, &notFound) {
				return nil, fmt.Errorf("read development config: %w", err)
			}

			return nil, fmt.Errorf("development.yaml not found: %w", err)
		}
	}

	if err := bindEnvs(v, keys); err != nil {
		return nil, fmt.Errorf("bind env: %w", err)
	}

	var env Env

	if err := v.Unmarshal(&env); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if appEnv != "" {
		env.App.Env = appEnv
	}

	if isProduction(appEnv) {
		applyOverrides(&env)
	}

	return &env, nil
}

func logConfigLoad(env *Env) {
	data, _ := json.Marshal(map[string]any{
		"level": "info",
		"env":   env.App.Env,
		"port":  env.App.Port,
		"msg":   "config load",
	})
	fmt.Println(string(data))
}

func getAppEnv() string {
	for _, key := range []string{"APP_ENV", "app_env"} {
		if value, ok := os.LookupEnv(key); ok {
			return value
		}
	}

	return ""
}

func normalizeAppEnv(appEnv string) string {
	return strings.ToLower(strings.TrimSpace(appEnv))
}

func applyOverrides(env *Env) {
	if origins := os.Getenv("APP_ALLOWED_ORIGINS"); origins != "" {
		env.App.AllowedOrigins = getEnvArray(origins)
	}
}

func getEnvArray(value string) []string {
	if value == "" {
		return []string{}
	}

	parts := strings.Split(value, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	return parts
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.env", "development")
	v.SetDefault("app.version", "dev")
	v.SetDefault("app.port", 8000)
	v.SetDefault("app.log_level", "debug")
	v.SetDefault("app.node_id", 0)
	v.SetDefault("app.global_request_timeout", 2*time.Minute)
	v.SetDefault("app.allowed_origins", []string{"http://localhost:5173"})
	v.SetDefault("app.frontend_url", "http://localhost:5173")

	v.SetDefault("jwt.refresh_exp", 168*time.Hour)
	v.SetDefault("jwt.expired", 15*time.Minute)

	v.SetDefault("db.url", "postgresql://postgres:postgres@localhost:5432/postgres?sslmode=disable")

	v.SetDefault("redis.url", "redis://localhost:6379")
	v.SetDefault("redis.tls", false)

	v.SetDefault("smtp.tls", false)
}

func isProduction(appEnv string) bool {
	return appEnv == "production" || appEnv == "prod"
}

func bindEnvs(v *viper.Viper, keys []string) error {
	for _, k := range keys {
		if err := v.BindEnv(k); err != nil {
			return err
		}
	}
	return nil
}

func getKeys() []string {
	return []string{
		"app.allowed_origins",
		"app.env",
		"app.log_level",
		"app.node_id",
		"app.port",
		"app.global_request_timeout",
		"app.version",
		"app.frontend_url",

		"jwt.secret",
		"jwt.expired",
		"jwt.refresh_exp",

		"db.url",

		"redis.url",
		"redis.tls",

		"smtp.port",
		"smtp.host",
		"smtp.user",
		"smtp.pass",
		"smtp.tls",

		"storage.cloudinary.cloud_name",
		"storage.cloudinary.api_secret",
		"storage.cloudinary.api_key",
	}
}

func getMigrationKeys() []string {
	return []string{
		"app.env",
		"app.log_level",
		"db.url",
	}
}
