package response

import (
	"time"

	"github.com/gin-gonic/gin"
)

type Meta map[string]any

type AppResponse struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data"`
	Meta      *Meta     `json:"meta,omitempty"`
}

func SuccessWithMeta(c *gin.Context, status int, message string, data any, meta Meta) {
	if data == nil {
		data = gin.H{}
	}

	c.JSON(status, AppResponse{
		Success:   true,
		Message:   message,
		Timestamp: time.Now().UTC(),
		Data:      data,
		Meta:      &meta,
	})
}

func Success(c *gin.Context, status int, message string, data any) {
	if data == nil {
		data = gin.H{}
	}

	c.JSON(status, AppResponse{
		Success:   true,
		Message:   message,
		Timestamp: time.Now().UTC(),
		Data:      data,
	})
}

func Error(status int, code, message string, timestamp time.Time) gin.H {
	return gin.H{
		"success":   false,
		"code":      code,
		"message":   message,
		"timestamp": timestamp,
	}
}
