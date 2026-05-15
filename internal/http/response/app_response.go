package response

import (
	"time"

	"github.com/gin-gonic/gin"
)

type AppResponse struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data"`
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

func Error(status int, message string, timestamp time.Time) gin.H {
	return gin.H{
		"success":   false,
		"message":   message,
		"timestamp": timestamp,
	}
}
