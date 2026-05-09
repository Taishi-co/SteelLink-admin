package main

import (
	"log"
	"os"

	"github.com/Taishi-co/SteelLink-admin/backend/internal/database"
	"github.com/Taishi-co/SteelLink-admin/backend/internal/handler"
	"github.com/Taishi-co/SteelLink-admin/backend/internal/router"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func shouldAllowDegradedStartup() bool {
	return os.Getenv("GIN_MODE") != gin.ReleaseMode
}

func main() {
	// Load optional local env file for backend development.
	_ = godotenv.Load()

	// Set Gin mode based on environment
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.DebugMode)
	}

	if err := database.InitMySQL(); err != nil {
		if shouldAllowDegradedStartup() {
			log.Printf("Warning: failed to connect MySQL, starting in degraded mode: %v", err)
		} else {
			log.Fatalf("Failed to connect MySQL: %v", err)
		}
	}
	defer func() {
		if err := database.CloseMySQL(); err != nil {
			log.Printf("Failed to close MySQL: %v", err)
		}
	}()

	if err := database.InitRedis(); err != nil {
		if shouldAllowDegradedStartup() {
			log.Printf("Warning: failed to connect Redis, starting in degraded mode: %v", err)
		} else {
			log.Fatalf("Failed to connect Redis: %v", err)
		}
	}
	defer func() {
		if err := database.CloseRedis(); err != nil {
			log.Printf("Failed to close Redis: %v", err)
		}
	}()

	if database.GetDB() != nil {
		handler.AutoMigrate(database.GetDB())
		log.Println("Auto-migration completed")
	} else {
		log.Println("Skipping auto-migration because MySQL is unavailable")
	}

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
