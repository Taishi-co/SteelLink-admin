package router

import (
	"github.com/Taishi-co/SteelLink-admin/backend/internal/handler"
	"github.com/Taishi-co/SteelLink-admin/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all application routes
func RegisterRoutes(r *gin.Engine) {
	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health check endpoint
	r.GET("/health", handler.HealthCheck)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Public routes
		v1.POST("/login", handler.Login)

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.Auth())
		{
			protected.POST("/logout", handler.Logout)
			protected.GET("/admin/info", handler.GetAdminInfo)

			// Admin routes
			protected.GET("/admins", handler.ListAdmins)
			protected.GET("/admins/:id", handler.GetAdmin)
			protected.POST("/admins", handler.CreateAdmin)
			protected.PUT("/admins/:id", handler.UpdateAdmin)
			protected.DELETE("/admins/:id", handler.DeleteAdmin)

			// Role routes
			protected.GET("/roles", handler.ListRoles)
			protected.GET("/roles/:id", handler.GetRole)
			protected.POST("/roles", handler.CreateRole)
			protected.PUT("/roles/:id", handler.UpdateRole)
			protected.DELETE("/roles/:id", handler.DeleteRole)

			// Menu routes
			protected.GET("/menus", handler.ListMenus)
			protected.GET("/menus/:id", handler.GetMenu)
			protected.POST("/menus", handler.CreateMenu)
			protected.PUT("/menus/:id", handler.UpdateMenu)
			protected.DELETE("/menus/:id", handler.DeleteMenu)

			// User management routes
			protected.GET("/users", handler.ListUsers)
			protected.GET("/users/:id", handler.GetUser)
			protected.PUT("/users/:id/status", handler.UpdateUserStatus)

			// Enterprise management routes
			protected.GET("/enterprises", handler.ListEnterprises)
			protected.GET("/enterprises/:id", handler.GetEnterprise)
			protected.PUT("/enterprises/:id/audit", handler.AuditEnterprise)

			// Product management routes
			protected.GET("/products", handler.ListProducts)
			protected.GET("/products/:id", handler.GetProduct)
			protected.PUT("/products/:id/audit", handler.AuditProduct)
			protected.GET("/product-categories", handler.ListProductCategories)
			protected.POST("/product-categories", handler.CreateProductCategory)
			protected.PUT("/product-categories/:id", handler.UpdateProductCategory)
			protected.DELETE("/product-categories/:id", handler.DeleteProductCategory)
		}
	}
}
