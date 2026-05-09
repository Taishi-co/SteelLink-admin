package router

import (
	"github.com/Taishi-co/SteelLink-admin/backend/internal/handler"
	"github.com/Taishi-co/SteelLink-admin/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
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

	r.GET("/health", handler.HealthCheck)

	v1 := r.Group("/api/v1")
	{
		v1.POST("/login", handler.Login)
		v1.GET("/internal/ai-procurement/runtime-config", handler.GetAIProcurementRuntimeConfig)

		protected := v1.Group("")
		protected.Use(middleware.Auth())
		{
			protected.POST("/logout", handler.Logout)
			protected.GET("/admin/info", handler.GetAdminInfo)
			protected.GET("/admin/menus", handler.GetMyMenus)
			protected.GET("/dashboard/stats", handler.GetDashboardStats)
			protected.GET("/files/private-url", handler.GetPrivateFileURL)
			protected.POST("/files/upload", handler.UploadFile)

			protected.GET("/users", handler.ListUsers)
			protected.GET("/users/:id", handler.GetUser)
			protected.PUT("/users/:id/status", handler.UpdateUserStatus)
			protected.PUT("/users/:id/identity", handler.AuditUserIdentity)

			protected.GET("/enterprises", handler.ListEnterprises)
			protected.GET("/enterprises/:id", handler.GetEnterprise)
			protected.PUT("/enterprises/:id/audit", handler.AuditEnterprise)
			protected.GET("/enterprises/:id/services", handler.ListEnterpriseServicePermissions)
			protected.PUT("/enterprises/:id/services/:service/status", handler.UpdateEnterpriseServicePermission)
			protected.GET("/enterprise-service-applications", handler.ListEnterpriseServiceApplications)
			protected.PUT("/enterprise-service-applications/:id/audit", handler.AuditEnterpriseServiceApplication)

			protected.GET("/products", handler.ListProducts)
			protected.GET("/products/:id", handler.GetProduct)
			protected.PUT("/products/:id", handler.UpdateProduct)
			protected.PUT("/products/:id/audit", handler.AuditProduct)
			protected.DELETE("/products/:id", handler.DeleteProduct)
			protected.GET("/warehouse-services", handler.ListWarehouseServiceOffers)
			protected.PUT("/warehouse-services/:id/audit", handler.AuditWarehouseServiceOffer)
			protected.GET("/logistics-services", handler.ListLogisticsServiceOffers)
			protected.PUT("/logistics-services/:id/audit", handler.AuditLogisticsServiceOffer)
			protected.GET("/finance-services", handler.ListFinanceServiceOffers)
			protected.PUT("/finance-services/:id/audit", handler.AuditFinanceServiceOffer)
			protected.GET("/orders", handler.ListOrders)
			protected.GET("/product-categories", handler.ListProductCategories)
			protected.GET("/operation-logs", handler.ListOperationLogs)

			superOnly := protected.Group("")
			superOnly.Use(middleware.RequireRole(1))
			{
				superOnly.GET("/admins", handler.ListAdmins)
				superOnly.GET("/admins/:id", handler.GetAdmin)
				superOnly.POST("/admins", handler.CreateAdmin)
				superOnly.PUT("/admins/:id", handler.UpdateAdmin)
				superOnly.DELETE("/admins/:id", handler.DeleteAdmin)

				superOnly.GET("/roles", handler.ListRoles)
				superOnly.GET("/roles/:id", handler.GetRole)
				superOnly.POST("/roles", handler.CreateRole)
				superOnly.PUT("/roles/:id", handler.UpdateRole)
				superOnly.DELETE("/roles/:id", handler.DeleteRole)
				superOnly.GET("/roles/:id/menus", handler.GetRoleMenus)
				superOnly.PUT("/roles/:id/menus", handler.UpdateRoleMenus)

				superOnly.GET("/menus", handler.ListMenus)
				superOnly.GET("/menus/:id", handler.GetMenu)
				superOnly.POST("/menus", handler.CreateMenu)
				superOnly.PUT("/menus/:id", handler.UpdateMenu)
				superOnly.DELETE("/menus/:id", handler.DeleteMenu)

				superOnly.POST("/product-categories", handler.CreateProductCategory)
				superOnly.PUT("/product-categories/:id", handler.UpdateProductCategory)
				superOnly.DELETE("/product-categories/:id", handler.DeleteProductCategory)
				superOnly.GET("/ai-procurement/config", handler.GetAIProcurementConfig)
				superOnly.PUT("/ai-procurement/config", handler.UpdateAIProcurementConfig)
			}
		}
	}
}
