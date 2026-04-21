package middleware

import (
	"net/http"

	"lockbox/internal/http/response"
	"lockbox/internal/logger"

	"github.com/gin-gonic/gin"
)

type RecoveryMiddleware struct {
	log logger.Logger
}

func NewRecoveryMiddleware(log logger.Logger) *RecoveryMiddleware {
	return &RecoveryMiddleware{log: log}
}

func (m *RecoveryMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				m.log.Error("panic recovered", nil, map[string]any{
					"panic": r,
					"path":  c.Request.URL.Path,
				})

				c.AbortWithStatusJSON(http.StatusInternalServerError, response.InternalServerError(""))
			}
		}()

		c.Next()
	}
}
