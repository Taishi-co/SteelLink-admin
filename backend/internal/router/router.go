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

			businessReadAdmin := protected.Group("")
			businessReadAdmin.Use(middleware.RequireRole(1, 2, 3))
			{
				businessReadAdmin.GET("/dashboard/stats", handler.GetDashboardStats)
				businessReadAdmin.GET("/files/private-url", handler.GetPrivateFileURL)
				businessReadAdmin.POST("/files/upload", handler.UploadFile)

				businessReadAdmin.GET("/users", handler.ListUsers)
				businessReadAdmin.GET("/users/:id", handler.GetUser)

				businessReadAdmin.GET("/enterprises", handler.ListEnterprises)
				businessReadAdmin.GET("/enterprises/:id", handler.GetEnterprise)
				businessReadAdmin.GET("/enterprises/:id/services", handler.ListEnterpriseServicePermissions)
				businessReadAdmin.GET("/enterprise-service-applications", handler.ListEnterpriseServiceApplications)

				businessReadAdmin.GET("/products", handler.ListProducts)
				businessReadAdmin.GET("/products/:id", handler.GetProduct)
				businessReadAdmin.GET("/supply-demands", handler.ListSupplyDemandPosts)
				businessReadAdmin.GET("/warehouse-services", handler.ListWarehouseServiceOffers)
				businessReadAdmin.GET("/logistics-services", handler.ListLogisticsServiceOffers)
				businessReadAdmin.GET("/finance-services", handler.ListFinanceServiceOffers)
				businessReadAdmin.GET("/orders", handler.ListOrders)
				businessReadAdmin.GET("/orders/:id", handler.GetOrder)
				businessReadAdmin.GET("/product-categories", handler.ListProductCategories)
				businessReadAdmin.GET("/operation-logs", handler.ListOperationLogs)
			}

			auditAdmin := protected.Group("")
			auditAdmin.Use(middleware.RequireRole(1, 2, 3))
			{
				auditAdmin.PUT("/users/:id/identity", handler.AuditUserIdentity)
				auditAdmin.PUT("/enterprises/:id/audit", handler.AuditEnterprise)
				auditAdmin.PUT("/enterprise-service-applications/:id/audit", handler.AuditEnterpriseServiceApplication)
				auditAdmin.PUT("/products/:id/audit", handler.AuditProduct)
				auditAdmin.PUT("/supply-demands/:id/audit", handler.AuditSupplyDemandPost)
				auditAdmin.PUT("/warehouse-services/:id/audit", handler.AuditWarehouseServiceOffer)
				auditAdmin.PUT("/logistics-services/:id/audit", handler.AuditLogisticsServiceOffer)
				auditAdmin.PUT("/finance-services/:id/audit", handler.AuditFinanceServiceOffer)
			}

			operatorAdmin := protected.Group("")
			operatorAdmin.Use(middleware.RequireRole(1, 2))
			{
				operatorAdmin.PUT("/users/:id/status", handler.UpdateUserStatus)
				operatorAdmin.PUT("/enterprises/:id/services/:service/status", handler.UpdateEnterpriseServicePermission)
				operatorAdmin.PUT("/products/:id", handler.UpdateProduct)
				operatorAdmin.DELETE("/products/:id", handler.DeleteProduct)
				operatorAdmin.POST("/orders/:id/freeze", handler.FreezeOrder)
				operatorAdmin.POST("/orders/:id/unfreeze", handler.UnfreezeOrder)
				operatorAdmin.POST("/orders/breach-records/:id/mediate", handler.MediateBreach)
				operatorAdmin.POST("/orders/breach-records/:id/resolve", handler.ResolveBreach)
				operatorAdmin.PUT("/orders/litigation-records/:id", handler.UpdateLitigationStatus)
			}

			notificationAdmin := protected.Group("")
			notificationAdmin.Use(middleware.RequireRole(1, 2))
			{
				notificationAdmin.POST("/admin/notifications", handler.SendNotification)
			}

			aiProcurementMonitorAdmin := protected.Group("")
			aiProcurementMonitorAdmin.Use(middleware.RequireRole(1, 2))
			{
				aiProcurementMonitorAdmin.GET("/ai-procurement/monitor", handler.GetAIProcurementMonitor)
			}

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
