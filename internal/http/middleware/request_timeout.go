package middleware

import (
	"context"
	"net/http"
	"time"

	apphttp "lockbox/internal/http"
	"lockbox/internal/http/response"

	"github.com/gin-gonic/gin"
)

type RequestTimeoutMiddleware struct{}

func NewRequestTimeoutMiddleware() *RequestTimeoutMiddleware {
	return &RequestTimeoutMiddleware{}
}

func (m *RequestTimeoutMiddleware) Handle(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		if ctx.Err() == context.DeadlineExceeded {
			appErr := response.NewAppErr(http.StatusBadGateway, apphttp.ErrTransport.Error(), "request timeout")
			c.AbortWithStatusJSON(appErr.Status, response.Error(appErr.Status, appErr.Code, appErr.Message, appErr.Timestamp))
		}
	}
}
