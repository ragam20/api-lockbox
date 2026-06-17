package server

import (
	"time"

	"lockbox/internal/config"
	"lockbox/internal/http/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewGinEngine(
	requestTimeout *middleware.RequestTimeoutMiddleware,
	errorHandling *middleware.ErrorMiddleware,
	recovery *middleware.RecoveryMiddleware,
	appConfig *config.AppConfig,
) *gin.Engine {
	corsMiddleware := cors.New(cors.Config{
		AllowOrigins:     allowedOrigins(appConfig),
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "Timezone"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})

	r := gin.New()
	r.Use(
		gin.Logger(),
		recovery.Handle(),
		corsMiddleware,
		requestTimeout.Handle(appConfig.GlobalRequestTimeout),
		errorHandling.Handle(),
	)

	return r
}

func allowedOrigins(cfg *config.AppConfig) []string {
	if len(cfg.AllowedOrigins) > 0 {
		return cfg.AllowedOrigins
	}
	return []string{"http://localhost:5173"}
}
