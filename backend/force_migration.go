package main

import (
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"log"
)

func main() {
	// Load configuration
	config.LoadConfig()

	// Connect to database
	db := database.ConnectDB()

	log.Println("🔌 Connected to database. Running AutoMigrate on Inventory model...")

	// Use GORM AutoMigrate to add missing columns
	err := db.AutoMigrate(&models.Inventory{})
	if err != nil {
		log.Fatalf("❌ Failed to auto-migrate: %v", err)
	}

	log.Println("✅ Database schema update executed successfully.")
	log.Println("✅ Inventory table now has project_id column if it was missing.")
}
