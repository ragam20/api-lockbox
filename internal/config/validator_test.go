package config

import (
	"strings"
	"testing"
	"time"
)

func TestValidateRejectsInvalidJWTConfig(t *testing.T) {
	env := Env{
		App: AppConfig{
			Env:                  "development",
			Version:              "test",
			Port:                 8000,
			LogLevel:             "debug",
			GlobalRequestTimeout: time.Second,
			FrontendURL:          "http://localhost:5173",
		},
		Database: DatabaseConfig{URL: "postgres://user:pass@localhost:5432/db"},
		RedisConfig: RedisConfig{
			URL: "redis://localhost:6379",
		},
		SMTPConfig: SMTPConfig{
			Port: 587,
			Host: "smtp.example.com",
			User: "user",
			Pass: "pass",
		},
	}

	err := env.Validate()
	if err == nil {
		t.Fatal("expected invalid JWT config to be rejected")
	}

	msg := err.Error()
	if !strings.Contains(msg, "jwt.secret is required") {
		t.Fatalf("expected missing secret error, got %q", msg)
	}
	if !strings.Contains(msg, "jwt.expired must be > 0") {
		t.Fatalf("expected invalid expiry error, got %q", msg)
	}
}
