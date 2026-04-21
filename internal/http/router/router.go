package router

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func NewRoutes(g *gin.Engine) {
	api := g.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":    http.StatusOK,
				"message":   "pong",
				"timestamp": time.Now().UTC(),
			})
		})
	}

	g.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	g.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"status":    http.StatusNotFound,
			"message":   "route not found",
			"timestamp": time.Now().UTC(),
		})
	})

	g.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"status":    http.StatusMethodNotAllowed,
			"message":   "method not allowed",
			"timestamp": time.Now().UTC(),
		})
	})
}
