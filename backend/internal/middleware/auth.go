package middleware

import (
	"net/http"
	"strings"

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
		// TODO: Validate JWT token
		// For now, just check if token is not empty
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "unauthorized: invalid token",
			})
			c.Abort()
			return
		}

		// Set admin info in context
		c.Set("admin_id", "1")
		c.Next()
	}
}
