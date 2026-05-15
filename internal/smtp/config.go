package smtp

import (
	"crypto/tls"

	"lockbox/internal"
	"lockbox/internal/config"

	"gopkg.in/gomail.v2"
)

type SMTPConfig struct {
	Dialer *gomail.Dialer
	From   string
}

func NewSMTPConfig(cfg *config.SMTPConfig, appCfg *config.AppConfig) *SMTPConfig {
	d := gomail.NewDialer(cfg.Host, cfg.Port, cfg.User, cfg.Pass)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: !internal.IsProduction(appCfg.Env)}

	return &SMTPConfig{
		Dialer: d,
		From:   cfg.User,
	}
}
