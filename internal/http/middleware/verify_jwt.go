package middleware

import (
	"strings"

	"lockbox/internal/app/auth"
	"lockbox/internal/constant"
	"lockbox/internal/http/response"

	"github.com/gin-gonic/gin"
)

type VerifyJWTMiddleware struct {
	jwtManager auth.TokenManager
}

func NewVerifyJWTMiddleware(jm auth.TokenManager) *VerifyJWTMiddleware {
	return &VerifyJWTMiddleware{jwtManager: jm}
}

func (m *VerifyJWTMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, _ := c.Cookie("authSession")

		if token == "" {
			token = getTokenFromHeader(c)
			if token == "" {
				_ = c.Error(response.Unauthorized())
				c.Abort()
				return
			}
		}

		jwtUser, err := m.jwtManager.ValidateAccessToken(token)
		if err != nil {
			_ = c.Error(err)
			c.Abort()
			return
		}

		c.Set(constant.UserIDCtx, jwtUser.UserID)
		c.Next()
	}
}

func getTokenFromHeader(c *gin.Context) string {
	bearer := c.GetHeader(constant.Authorization)
	if bearer == "" {
		return ""
	}

	s := strings.Split(bearer, " ")
	if len(s) < 2 || strings.ToLower(s[0]) != "bearer" {
		return ""
	}

	return s[1]
}
