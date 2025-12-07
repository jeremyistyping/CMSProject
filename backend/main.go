package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/routes"
	"app-sistem-akuntansi/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Embed migration files into binary for auto-migration
//
//go:embed migrations/*.sql
var migrationsFS embed.FS

// @title Unipro Project Management API
// @version 1.0
// @description API untuk aplikasi Project Management System dengan fitur lengkap manajemen proyek, daily updates, milestones, dan cost control.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @schemes http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Set Gin mode based on configuration
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Connect to database
	db := database.ConnectDB()

	// AUTO-MIGRATION: Run golang-migrate migrations
	log.Println("🔄 Running database migrations...")
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("⚠️ Warning: Could not get SQL DB for migrations: %v", err)
	} else {
		migrationService := services.NewMigrationService(sqlDB)
		if err := migrationService.RunMigrations(migrationsFS, "migrations"); err != nil {
			log.Printf("⚠️ Warning: Auto-migration encountered an issue: %v", err)
			log.Println("📝 Application will continue, but some schema changes may not be applied")
		} else {
			version, dirty, _ := migrationService.GetCurrentVersion(migrationsFS, "migrations")
			if dirty {
				log.Printf("⚠️ Database is in dirty state at version %d", version)
			} else {
				log.Printf("✅ Database migrations complete - version: %d", version)
			}
		}
	}

	// Auto migrate models (GORM AutoMigrate for models)
	database.AutoMigrate(db)

	// Run auto SQL migrations
	if err := database.RunAutoMigrations(db); err != nil {
		log.Printf("Warning: Auto-migration failed: %v", err)
	}

	// Migrate permissions table
	if err := database.MigratePermissions(db); err != nil {
		log.Printf("Error migrating permissions: %v", err)
	}

	// Seed database with initial data
	database.SeedData(db)

	// Run startup tasks
	startupService := services.NewStartupService(db)
	startupService.RunStartupTasks()

	// Run database optimization for better performance
	log.Println("⚡ Starting database performance optimization...")
	if err := database.RunFullDatabaseOptimization(db); err != nil {
		log.Printf("⚠️ Warning: Database optimization failed: %v", err)
	} else {
		log.Println("✅ Database optimization completed successfully!")
	}

	// Initialize Gin router without default middleware
	r := gin.New()

	// Add recovery middleware
	r.Use(gin.Recovery())

	// Add custom logger middleware only in development
	if cfg.Environment != "production" {
		r.Use(gin.Logger())
	}

	// Configure trusted proxies for security
	r.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	// CORS middleware with dynamic origins
	allowedOrigins := config.GetAllowedOrigins(cfg)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "Cache-Control"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Update Swagger docs with dynamic configuration
	if cfg.Environment == "development" || os.Getenv("ENABLE_SWAGGER") == "true" {
		config.UpdateSwaggerDocs()
		config.PrintSwaggerInfo()

		log.Println("🎆 Setting up Enhanced Swagger with authentication support...")
		config.SetupEnhancedSwaggerRoutes(r)
		log.Println("✅ Enhanced Swagger routes configured successfully!")
	}

	// Serve static files for uploads with proper CORS headers
	uploadDir := filepath.Join(".", "uploads")
	var uploadPath string
	if absPath, err := filepath.Abs(uploadDir); err == nil {
		log.Printf("📁 Serving static files from: %s", absPath)
		uploadPath = absPath
	} else {
		log.Printf("⚠️ Warning: Could not get absolute path for uploads, using relative path")
		uploadPath = "./uploads"
	}

	// Static file handler with CORS headers for images
	log.Println("🖼️  Registering /uploads route handler...")

	r.GET("/uploads/*filepath", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Header("Cache-Control", "public, max-age=31536000")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		filePath := c.Param("filepath")
		if len(filePath) > 0 && filePath[0] == '/' {
			filePath = filePath[1:]
		}
		fullPath := filepath.Join(uploadPath, filePath)

		log.Printf("📝 Serving file: %s -> %s", c.Request.URL.Path, fullPath)

		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			log.Printf("❌ File not found: %s", fullPath)
			c.JSON(404, gin.H{"error": "File not found", "path": filePath, "full_path": fullPath})
			return
		}

		c.File(fullPath)
		log.Printf("✅ File served successfully: %s", fullPath)
	})
	log.Println("✅ /uploads route handler registered!")

	// Setup routes
	routes.SetupRoutes(r, db, startupService)

	// Start server
	port := cfg.ServerPort
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Unipro Project Management Server starting on port %s", port)
	log.Fatal(r.Run(":" + port))
}
