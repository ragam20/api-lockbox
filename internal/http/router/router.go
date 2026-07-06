package router

import (
	"net/http"
	"time"

	apphttp "lockbox/internal/http"
	"lockbox/internal/http/handler"
	"lockbox/internal/http/middleware"
	"lockbox/internal/http/response"

	"github.com/gin-gonic/gin"
)

func NewRoutes(
	g *gin.Engine,
	authHandler *handler.AuthHandler,
	vaultHandler *handler.VaultHandler,
	userHandler *handler.UserHandler,
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
	authAPI.Use(middleware.NewRequestTimeoutMiddleware().Handle(12 * time.Second))
	{
		authAPI.POST("/logout", verifyJWT.Handle(), authHandler.Logout)
		authAPI.POST("/request-otp", authHandler.RequestOTP)
		authAPI.POST("/verify-otp", authHandler.VerifyOTP)
		authAPI.POST("/unlock", verifyJWT.Handle(), authHandler.Unlock)
		authAPI.POST("/refresh-token", authHandler.RefreshToken)
		authAPI.POST("/register", authHandler.Register)
		authAPI.POST("/register/request-verification-link", authHandler.SendVerifyLink)
		authAPI.GET("/register/verify-email", authHandler.VerifyEmail)
		authAPI.POST("/register/session", authHandler.GetRegisterSession)
	}

	vaultAPI := api.Group("/vault")
	vaultAPI.Use(requestTimeout.Handle(5*time.Second), verifyJWT.Handle())
	{
		vaultAPI.GET("", vaultHandler.GetVault)
		vaultAPI.POST("/items", vaultHandler.CreateVaultItem)
		vaultAPI.GET("/items", vaultHandler.GetVaultItems)
		vaultAPI.GET("/items/favorites", vaultHandler.GetFavorites)
		vaultAPI.GET("/items/:id", vaultHandler.GetVaultItemByID)
		vaultAPI.PUT("/items/:id", vaultHandler.UpdateVaultItem)
		vaultAPI.POST("/items/:id/archive", vaultHandler.ArchiveVaultItem)
		vaultAPI.POST("/items/:id/unarchive", vaultHandler.UnarchiveVaultItem)
		vaultAPI.DELETE("/items/:id", vaultHandler.DeleteVaultItem)
		vaultAPI.PUT("/items/:id/favorite", vaultHandler.SetFavorite)
		vaultAPI.DELETE("/items/:id/favorite", vaultHandler.DeleteFavorite)
		vaultAPI.GET("/argon-config", vaultHandler.GetArgonConfig)
		vaultAPI.GET("/items/protection-key", vaultHandler.GetProtectionKeys)
		vaultAPI.PUT("/argon-config", vaultHandler.UpdateArgonConfig)
	}

	userAPI := api.Group("/user",
		middleware.NewRequestTimeoutMiddleware().Handle(5*time.Second),
		verifyJWT.Handle())
	{
		userAPI.GET("", userHandler.GetUser)
		userAPI.PUT("", userHandler.UpdateProfile)
		userAPI.GET("/avatar/upload-permission", userHandler.UploadAvatarPermission)
		userAPI.PATCH("/avatar", userHandler.SaveAvatarURL)
		userAPI.DELETE("/avatar", userHandler.RemoveAvatar)
	}

	g.NoRoute(func(c *gin.Context) {
		appErr := response.NotFoundError(apphttp.ErrTransport.Error(), "route not found")
		c.JSON(appErr.Status, response.Error(appErr.Status, appErr.Code, appErr.Message, appErr.Timestamp))
	})

	g.NoMethod(func(c *gin.Context) {
		appErr := response.NewAppErr(http.StatusMethodNotAllowed, apphttp.ErrTransport.Error(), "method not allowed")
		c.JSON(appErr.Status, response.Error(appErr.Status, appErr.Code, appErr.Message, appErr.Timestamp))
	})
}
