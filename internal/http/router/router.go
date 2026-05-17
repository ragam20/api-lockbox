package router

import (
	"net/http"
	"time"

	"lockbox/internal/http/handler"
	"lockbox/internal/http/middleware"
	"lockbox/internal/http/response"

	"github.com/gin-gonic/gin"
)

func NewRoutes(
	g *gin.Engine,
	authHandler *handler.AuthHandler,
	vaultHandler *handler.VaultHandler,
	verifyJWT *middleware.VerifyJWTMiddleware,
	requestTimeout *middleware.RequestTimeoutMiddleware,
) {
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
	authAPI.Use(middleware.NewRequestTimeoutMiddleware().Handle(5 * time.Second))
	{
		authAPI.POST("/login", authHandler.Login)
		authAPI.POST("/refresh-token", authHandler.RefreshToken)
		authAPI.POST("/register", authHandler.Register)
		authAPI.POST("/register/send-verification-link", authHandler.SendVerifyLink)
		authAPI.GET("/register/verify-email", authHandler.VerifyEmail)
	}

	vaultAPI := api.Group("/vault")
	vaultAPI.Use(verifyJWT.Handle(), requestTimeout.Handle(5*time.Second))
	{
		vaultAPI.POST("/items", vaultHandler.CreateVaultItem)
		vaultAPI.GET("/items", vaultHandler.GetVaultItem)
		vaultAPI.GET("/items/:id", vaultHandler.GetVaultItemByID)
		vaultAPI.PUT("/items/:id", vaultHandler.UpdateVaultItem)
		vaultAPI.POST("/items/:id/archive", vaultHandler.ArchiveVaultItem)
		vaultAPI.POST("/items/:id/unarchive", vaultHandler.UnarchiveVaultItem)
		vaultAPI.DELETE("/items/:id", vaultHandler.DeleteVaultItem)
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
