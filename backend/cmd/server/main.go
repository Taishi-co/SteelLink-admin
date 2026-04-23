package main

import (
	"log"
	"os"

	"github.com/Taishi-co/SteelLink-admin/backend/internal/database"
	"github.com/Taishi-co/SteelLink-admin/backend/internal/router"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load optional local env file for backend development.
	_ = godotenv.Load()

	// Set Gin mode based on environment
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.DebugMode)
	}

	if err := database.InitMySQL(); err != nil {
		log.Fatalf("Failed to connect MySQL: %v", err)
	}
	defer func() {
		if err := database.CloseMySQL(); err != nil {
			log.Printf("Failed to close MySQL: %v", err)
		}
	}()

	// Create Gin engine
	r := gin.Default()

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
