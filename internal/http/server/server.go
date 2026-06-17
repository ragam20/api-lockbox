package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"lockbox/internal/config"
	"lockbox/internal/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

func NewHTTPServer(
	lc fx.Lifecycle,
	cfg *config.AppConfig,
	ge *gin.Engine,
	log logger.Logger,
) *http.Server {
	port := os.Getenv("PORT")

	if port == "" && cfg.Port != 0 {
		port = strconv.Itoa(cfg.Port)
	}

	if port == "" {
		port = "8000"
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: ge,
	}

	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				log.Info("listening on", map[string]any{
					"port": port,
				})

				go func() {
					if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
						log.Error("http server failed", err, nil)
					}
				}()

				return nil
			},
			OnStop: func(ctx context.Context) error {
				return srv.Shutdown(ctx)
			},
		},
	)

	return srv
}
