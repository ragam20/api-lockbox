package middleware

import (
	"net/http"
	"runtime/debug"
	"strings"

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
				stack := FilterStack(debug.Stack())

				m.log.Error("panic recovered", nil, map[string]any{
					"panic":  r,
					"method": c.Request.Method,
					"path":   c.Request.URL.Path,
					"stack":  stack,
				})

				c.AbortWithStatusJSON(http.StatusInternalServerError, response.InternalServerError(""))
			}
		}()

		c.Next()
	}
}

func FilterStack(stack []byte) string {
	lines := strings.Split(string(stack), "\n")

	var result []string

	for _, line := range lines {
		if strings.Contains(line, "/internal/") {
			result = append(result, strings.TrimSpace(line))
		}
	}

	return strings.Join(result, "\n")
}
