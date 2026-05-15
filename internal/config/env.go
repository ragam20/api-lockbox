package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Env struct {
	App         AppConfig      `mapstructure:"app"`
	Database    DatabaseConfig `mapstructure:"db"`
	JWTConfig   JWTConfig      `mapstructure:"jwt"`
	RedisConfig RedisConfig    `mapstructure:"redis"`
	SMTPConfig  SMTPConfig     `mapstructure:"smtp"`
}

func NewEnv() (*Env, error) {
	appEnv := os.Getenv("APP_ENV")
	configName := "development"
	if appEnv != "" {
		configName = appEnv
	}

	v := viper.New()
	v.SetConfigName(configName)
	v.SetConfigType("yaml")
	v.AddConfigPath("internal/config")
	v.AddConfigPath(".")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	if !isProduction(appEnv) {
		if err := v.ReadInConfig(); err != nil {
			var notFound viper.ConfigFileNotFoundError
			if !errors.As(err, &notFound) {
				return nil, fmt.Errorf("read config %q: %w", configName, err)
			}
		}
	}

	if err := bindEnvs(v, getKeys()); err != nil {
		return nil, fmt.Errorf("bind env: %w", err)
	}

	var env Env
	if err := v.Unmarshal(&env); err != nil {
		return nil, fmt.Errorf("parse config %q: %w", configName, err)
	}
	if err := env.Validate(); err != nil {
		return nil, fmt.Errorf("validate config %q: %w", configName, err)
	}

	data, _ := json.Marshal(map[string]any{
		"level": "info",
		"env":   env.App.Env,
		"port":  env.App.Port,
		"msg":   "config load",
	})
	fmt.Println(string(data))

	return &env, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.env", "development")
	v.SetDefault("app.version", "dev")
	v.SetDefault("app.port", 8000)
	v.SetDefault("app.log_level", "debug")
	v.SetDefault("app.node_id", 0)
	v.SetDefault("app.request_timeout", 5*time.Second)
	v.SetDefault("app.allowed_origins", []string{"http://localhost:5173"})
	v.SetDefault("app.frontend_url", "http://localhost:5173")
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
		"app.request_timeout",
		"app.version",
		"app.frontend_url",

		"jwt.secret",
		"jwt.expired",

		"db.url",

		"redis.url",

		"smtp.port",
		"smtp.host",
		"smtp.user",
		"smtp.pass",
	}
}
