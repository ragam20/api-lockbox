package router

import (
	"net/http"

	"lockbox/internal/http/handler"
	"lockbox/internal/http/response"

	"github.com/gin-gonic/gin"
)

func NewRoutes(g *gin.Engine, authHandler *handler.AuthHandler) {
	api := g.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			response.Success(c, http.StatusOK, "pong", nil)
		})
	}

	g.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	authAPI := api.Group("/auth")
	{
		authAPI.POST("/register/send-verification-link", authHandler.SendVerifyLink)
		authAPI.GET("/register/verify-email", authHandler.VerifyEmail)
		authAPI.POST("/register", authHandler.Register)
	}

	g.NoRoute(func(c *gin.Context) {
		appErr := response.NotFoundError("route not found")
		c.JSON(appErr.Status, response.Error(appErr.Status, appErr.Message, appErr.Timestamp))
	})

	g.NoMethod(func(c *gin.Context) {
		appErr := response.NewAppErr(http.StatusMethodNotAllowed, "method not allowed")
		c.JSON(appErr.Status, response.Error(appErr.Status, appErr.Message, appErr.Timestamp))
	})
}
