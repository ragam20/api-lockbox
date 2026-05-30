package config

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog"
)

func (e *Env) Validate() error {
	var errs []string

	validateAppConfig(e.App, &errs)
	validateJWTConfig(e.JWTConfig, &errs)
	validateDatabaseConfig(e.Database, &errs)
	validateRedisConfig(e.RedisConfig, &errs)
	validateSMTPConfig(e.SMTPConfig, &errs)

	if len(errs) == 0 {
		return nil
	}

	return fmt.Errorf("invalid config: %s", strings.Join(errs, "; "))
}

func validateAppConfig(cfg AppConfig, errs *[]string) {
	if strings.TrimSpace(cfg.Env) == "" {
		*errs = append(*errs, "app.env is required")
	}

	if strings.TrimSpace(cfg.Version) == "" {
		*errs = append(*errs, "app.version is required")
	}

	if cfg.Port <= 0 || cfg.Port > 65535 {
		*errs = append(*errs, "app.port must be between 1 and 65535")
	}

	if strings.TrimSpace(cfg.LogLevel) == "" {
		*errs = append(*errs, "app.log_level is required")
	} else if _, err := zerolog.ParseLevel(cfg.LogLevel); err != nil {
		*errs = append(*errs, "app.log_level is invalid")
	}

	if cfg.NodeID < 0 {
		*errs = append(*errs, "app.node_id must be >= 0")
	}

	if cfg.GlobalRequestTimeout <= 0 {
		*errs = append(*errs, "app.global_request_timeout must be > 0")
	}

	if strings.TrimSpace(cfg.FrontendURL) == "" {
		*errs = append(*errs, "app.frontend_url is required")
	}
}

func validateJWTConfig(cfg JWTConfig, errs *[]string) {
	if strings.TrimSpace(cfg.Secret) == "" {
		*errs = append(*errs, "jwt.secret is required")
	}

	if cfg.Expired <= 0 {
		*errs = append(*errs, "jwt.expired must be > 0")
	}

	if cfg.RefreshExp <= 0 {
		*errs = append(*errs, "jwt.refresh_exp must be > 0")
	}
}

func validateDatabaseConfig(cfg DatabaseConfig, errs *[]string) {
	if strings.TrimSpace(cfg.URL) == "" {
		*errs = append(*errs, "db.url is required")
	}
}

func validateRedisConfig(cfg RedisConfig, errs *[]string) {
	if strings.TrimSpace(cfg.URL) == "" {
		*errs = append(*errs, "redis.url is required")
	}
}

func validateSMTPConfig(cfg SMTPConfig, errs *[]string) {
	if cfg.Port <= 0 || cfg.Port > 65535 {
		*errs = append(*errs, "smtp.port must be between 1 and 65535")
	}

	if strings.TrimSpace(cfg.Host) == "" {
		*errs = append(*errs, "smtp.host is required")
	}

	if strings.TrimSpace(cfg.User) == "" {
		*errs = append(*errs, "smtp.user is required")
	}

	if strings.TrimSpace(cfg.Pass) == "" {
		*errs = append(*errs, "smtp.pass is required")
	}
}
