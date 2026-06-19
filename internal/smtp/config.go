package smtp

import (
	"crypto/tls"

	"lockbox/internal/config"

	"gopkg.in/gomail.v2"
)

type SMTPConfig struct {
	Dialer *gomail.Dialer
	From   string
}

func NewSMTPConfig(cfg *config.SMTPConfig, appCfg *config.AppConfig) *SMTPConfig {
	d := gomail.NewDialer(cfg.Host, cfg.Port, cfg.User, cfg.Pass)

	if cfg.TLS {
		d.TLSConfig = &tls.Config{InsecureSkipVerify: cfg.TLS}
	}

	return &SMTPConfig{
		Dialer: d,
		From:   cfg.User,
	}
}
