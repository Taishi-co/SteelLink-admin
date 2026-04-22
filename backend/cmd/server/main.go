package main

import (
	"log"
	"os"

	"github.com/Taishi-co/SteelLink-admin/backend/internal/store"
	"github.com/Taishi-co/SteelLink-admin/backend/internal/router"
	"github.com/gin-gonic/gin"
)

func main() {
	// Set Gin mode based on environment
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.DebugMode)
	}

	// Create Gin engine
	r := gin.Default()

	if err := store.InitDB(); err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// Register routes
	router.RegisterRoutes(r)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
