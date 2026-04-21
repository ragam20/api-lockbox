package middleware

import (
	"context"
	"net/http"

	"lockbox/internal/config"
	"lockbox/internal/http/response"

	"github.com/gin-gonic/gin"
)

type RequestTimeoutMiddleware struct {
	cfg *config.AppConfig
}

func NewRequestTimeoutMiddleware(cfg *config.AppConfig) *RequestTimeoutMiddleware {
	return &RequestTimeoutMiddleware{cfg: cfg}
}

func (m *RequestTimeoutMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), m.cfg.RequestTimeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		if ctx.Err() == context.DeadlineExceeded {
			c.AbortWithStatusJSON(http.StatusBadGateway, response.NewAppErr(http.StatusBadGateway, "request timeout"))
		}
	}
}
