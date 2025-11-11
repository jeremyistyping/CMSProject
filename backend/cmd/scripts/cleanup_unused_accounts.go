package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"app-sistem-akuntansi/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	log.Println("🧹 Starting cleanup of unused accounts...")

	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("⚠️  Warning: .env not found, using environment variables")
	}

	// Database connection
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "postgres"),
		getEnv("DB_NAME", "db_unipro"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_SSLMODE", "disable"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	log.Println("✅ Connected to database")

	// List of accounts to delete
	accountsToDelete := []string{
		// Fixed Assets - not needed for project cost control
		"1500", "1501", "1502", "1503", "1509",
		// Old Liability Structure
		"2100", "2107", "2108",
		// Old Equity
		"3201",
		// Old Revenue
		"4900",
		// Old Expenses
		"5203", "5204", "5900",
	}

	log.Println("\n📋 Accounts to be deleted:")
	for _, code := range accountsToDelete {
		var account models.Account
		if err := db.Where("code = ? AND deleted_at IS NULL", code).First(&account).Error; err == nil {
			log.Printf("   - %s: %s (%s)", account.Code, account.Name, account.Type)
		}
	}

	// Confirm deletion
	log.Println("\n⚠️  This will soft-delete the above accounts.")
	log.Print("Continue? (yes/no): ")
	var confirmation string
	fmt.Scanln(&confirmation)

	if confirmation != "yes" {
		log.Println("❌ Cleanup cancelled by user")
		return
	}

	// Perform soft delete
	result := db.Model(&models.Account{}).
		Where("code IN ?", accountsToDelete).
		Where("deleted_at IS NULL").
		Update("deleted_at", time.Now())

	if result.Error != nil {
		log.Fatalf("❌ Failed to delete accounts: %v", result.Error)
	}

	log.Printf("✅ Successfully soft-deleted %d accounts\n", result.RowsAffected)

	// Show remaining accounts
	log.Println("\n📊 Remaining active accounts:")
	var activeAccounts []models.Account
	db.Where("deleted_at IS NULL").Order("code").Find(&activeAccounts)

	currentType := ""
	for _, acc := range activeAccounts {
		if string(acc.Type) != currentType {
			currentType = string(acc.Type)
			log.Printf("\n=== %s ===", currentType)
		}
		log.Printf("  %s - %s", acc.Code, acc.Name)
	}

	log.Println("\n✅ Cleanup completed successfully!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
