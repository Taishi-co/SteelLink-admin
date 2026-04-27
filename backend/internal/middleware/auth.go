package middleware

import (
	"net/http"
	"strings"

	"github.com/Taishi-co/SteelLink-admin/backend/internal/auth"
	"github.com/gin-gonic/gin"
)

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
		c.Set("admin_role_id", claims.RoleID)
		c.Next()
	}
}

func RequireRole(roleIDs ...int) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleID, exists := c.Get("admin_role_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "unauthorized",
			})
			c.Abort()
			return
		}

		adminRoleID := roleID.(int)
		for _, id := range roleIDs {
			if adminRoleID == id {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "forbidden: insufficient role permissions",
		})
		c.Abort()
	}
}

func IsSuperAdmin(c *gin.Context) bool {
	roleID, exists := c.Get("admin_role_id")
	if !exists {
		return false
	}
	return roleID.(int) == 1
}
