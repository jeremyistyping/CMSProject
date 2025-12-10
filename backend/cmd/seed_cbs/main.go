package main

import (
	"app-sistem-akuntansi/database"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("🌱 CBS Seeder for Padel Bandung")
	log.Println("================================")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using environment variables")
	}

	// Connect to database
	db := database.ConnectDB()

	// Run CBS seeder
	if err := database.SeedCBSPadelBandung(db); err != nil {
		log.Fatalf("❌ Failed to seed CBS data: %v", err)
	}

	log.Println("\n✅ CBS seeding completed successfully!")
	os.Exit(0)
}
