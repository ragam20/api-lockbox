package middleware

import (
	"errors"
	"net/http"
	"time"

	"lockbox/internal/http/response"
	"lockbox/internal/logger"

	"github.com/gin-gonic/gin"
)

type ErrorMiddleware struct {
	log logger.Logger
}

func NewErrorMiddleware(log logger.Logger) *ErrorMiddleware {
	return &ErrorMiddleware{log: log}
}

func (m *ErrorMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if c.Writer.Written() || len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		status := http.StatusInternalServerError
		msg := "internal server error"
		timestamp := time.Now().UTC()

		var appErr *response.AppErrResponse
		if errors.As(err, &appErr) {
			status = appErr.Status
			timestamp = appErr.Timestamp

			if status >= 500 {
				m.log.Error(appErr.Message, err, map[string]any{
					"status":    appErr.Status,
					"timestamp": appErr.Timestamp,
					"path":      c.Request.URL.Path,
					"ipAddress": c.ClientIP(),
				})
			}

			if status >= 400 && status < 500 {
				msg = appErr.Message
			}
		}

		c.AbortWithStatusJSON(status, gin.H{
			"success":   false,
			"message":   msg,
			"timestamp": timestamp,
		})
	}
}
