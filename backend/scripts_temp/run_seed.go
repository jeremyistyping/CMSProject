package main

import (
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"log"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	log.Printf("Database URL: %s", cfg.DatabaseURL)

	// Connect to database
	db := database.ConnectDB()
	
	// Run seeding
	log.Println("🌱 Starting database seeding...")
	database.SeedData(db)
	log.Println("✅ Database seeding completed!")
}