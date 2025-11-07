package main

import (
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"log"
	"os"
)

func main() {
	// Load configuration (for potential future use)
	_ = config.LoadConfig()

	// Connect to database
	db := database.ConnectDB()

	log.Println("Running shares fields migration...")

	// Run the migration
	if err := database.RunSharesFieldsMigration(db); err != nil {
		log.Fatalf("Failed to run shares fields migration: %v", err)
		os.Exit(1)
	}

	log.Println("Shares fields migration completed successfully!")
}
