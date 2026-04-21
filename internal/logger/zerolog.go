package logger

import (
	"fmt"
	"os"

	"lockbox/internal/config"

	"github.com/rs/zerolog"
)

type Logger interface {
	Info(msg string, fields map[string]any)
	Error(msg string, err error, fields map[string]any)
}

type ZeroLogger struct {
	log zerolog.Logger
}

func NewLogger(cfg *config.AppConfig) (Logger, error) {
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		return &ZeroLogger{}, fmt.Errorf("invalid log level: %w", err)
	}

	log := zerolog.New(os.Stdout).Level(level).With().Timestamp().Logger()

	if cfg.Env == "development" {
		log = log.Output(zerolog.ConsoleWriter{
			Out: os.Stdout,
		})
	}

	return &ZeroLogger{
		log: log,
	}, nil
}

func (l *ZeroLogger) Info(msg string, fields map[string]any) {
	event := l.log.Info()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

func (l *ZeroLogger) Error(msg string, err error, fields map[string]any) {
	event := l.log.Error().Err(err)
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}
