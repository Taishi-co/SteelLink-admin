package middleware

import (
	"net/http"
	"strings"

	"github.com/Taishi-co/SteelLink-admin/backend/internal/auth"
	"github.com/gin-gonic/gin"
)

// Auth is a JWT authentication middleware
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "unauthorized: missing authorization header",
			})
			c.Abort()
			return
		}

		// Extract token from Bearer header
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "unauthorized: invalid authorization format",
			})
			c.Abort()
			return
		}

		token := parts[1]
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "unauthorized: invalid token",
			})
			c.Abort()
			return
		}

		claims, err := auth.ParseAdminToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "unauthorized: token validation failed",
			})
			c.Abort()
			return
		}

		c.Set("admin_id", claims.AdminID)
		c.Set("admin_username", claims.Username)
		c.Set("admin_real_name", claims.RealName)
		c.Next()
	}
}
