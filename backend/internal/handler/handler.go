package handler

import (
	"net/http"
	"os"

	"github.com/Taishi-co/SteelLink-admin/backend/internal/auth"
	"github.com/gin-gonic/gin"
)

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func defaultAdminUsername() string {
	if value := os.Getenv("ADMIN_DEFAULT_USERNAME"); value != "" {
		return value
	}
	return "admin"
}

func defaultAdminPassword() string {
	if value := os.Getenv("ADMIN_DEFAULT_PASSWORD"); value != "" {
		return value
	}
	return "admin123"
}

func defaultAdminRealName() string {
	if value := os.Getenv("ADMIN_DEFAULT_REAL_NAME"); value != "" {
		return value
	}
	return "Platform Admin"
}

// HealthCheck returns server health status
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "SteelLink Admin API is running",
	})
}

// Login handles admin login
func Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid login payload",
		})
		return
	}

	if req.Username != defaultAdminUsername() || req.Password != defaultAdminPassword() {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "invalid username or password",
		})
		return
	}

	token, err := auth.GenerateAdminToken(1, defaultAdminUsername(), defaultAdminRealName())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"token": token,
			"adminInfo": gin.H{
				"id":       1,
				"username": defaultAdminUsername(),
				"realName": defaultAdminRealName(),
				"roleId":   1,
			},
		},
	})
}

// Logout clears current session on client side
func Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    true,
	})
}

// GetAdminInfo returns current admin info from token claims
func GetAdminInfo(c *gin.Context) {
	adminID, _ := c.Get("admin_id")
	username, _ := c.Get("admin_username")
	realName, _ := c.Get("admin_real_name")

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"id":       adminID,
			"username": username,
			"realName": realName,
			"roleId":   1,
		},
	})
}

// ListAdmins returns list of admins
func ListAdmins(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":  []interface{}{},
			"total": 0,
		},
	})
}

// GetAdmin returns a single admin
func GetAdmin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    nil,
	})
}

// CreateAdmin creates a new admin
func CreateAdmin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// UpdateAdmin updates an admin
func UpdateAdmin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// DeleteAdmin deletes an admin
func DeleteAdmin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// ListRoles returns list of roles
func ListRoles(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":  []interface{}{},
			"total": 0,
		},
	})
}

// GetRole returns a single role
func GetRole(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    nil,
	})
}

// CreateRole creates a new role
func CreateRole(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// UpdateRole updates a role
func UpdateRole(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// DeleteRole deletes a role
func DeleteRole(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// ListMenus returns list of menus
func ListMenus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":  []interface{}{},
			"total": 0,
		},
	})
}

// GetMenu returns a single menu
func GetMenu(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    nil,
	})
}

// CreateMenu creates a new menu
func CreateMenu(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// UpdateMenu updates a menu
func UpdateMenu(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// DeleteMenu deletes a menu
func DeleteMenu(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// ListUsers returns list of users
func ListUsers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":  []interface{}{},
			"total": 0,
		},
	})
}

// GetUser returns a single user
func GetUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    nil,
	})
}

// UpdateUserStatus updates user status
func UpdateUserStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// ListEnterprises returns list of enterprises
func ListEnterprises(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":  []interface{}{},
			"total": 0,
		},
	})
}

// GetEnterprise returns a single enterprise
func GetEnterprise(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    nil,
	})
}

// AuditEnterprise audits an enterprise
func AuditEnterprise(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// ListProducts returns list of products
func ListProducts(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":  []interface{}{},
			"total": 0,
		},
	})
}

// GetProduct returns a single product
func GetProduct(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    nil,
	})
}

// AuditProduct audits a product
func AuditProduct(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// ListProductCategories returns list of product categories
func ListProductCategories(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":  []interface{}{},
			"total": 0,
		},
	})
}

// CreateProductCategory creates a new product category
func CreateProductCategory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// UpdateProductCategory updates a product category
func UpdateProductCategory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// DeleteProductCategory deletes a product category
func DeleteProductCategory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}
