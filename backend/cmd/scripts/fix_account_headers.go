package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"app-sistem-akuntansi/repositories"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("🔧 Starting Account Header Status Fix...")

	// Get database URL from environment or use default
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		// Default for local development
		databaseURL = "host=localhost user=postgres password=postgres dbname=accounting_system port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("📊 Database connected successfully")

	// Initialize account repository
	accountRepo := repositories.NewAccountRepository(db)

	// Call the fix function
	ctx := context.Background()
	if err := accountRepo.(*repositories.AccountRepo).FixAccountHeaderStatus(ctx); err != nil {
		log.Fatalf("Failed to fix account header status: %v", err)
	}

	log.Println("✅ Account Header Status Fix completed successfully!")
	log.Println("🔍 All accounts that have children are now marked as headers")
	log.Println("🔍 All accounts that don't have children are no longer marked as headers")
	fmt.Println("")
	log.Println("💡 You can now test the account hierarchy at http://localhost:3000/accounts")
}
