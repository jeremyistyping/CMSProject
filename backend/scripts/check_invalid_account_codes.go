package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Account struct {
	ID   uint   `gorm:"column:id"`
	Code string `gorm:"column:code"`
	Name string `gorm:"column:name"`
}

func main() {
	// Setup database connection
	dsn := "host=localhost user=postgres password=postgres dbname=accounting_db port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("🔍 Checking for invalid account codes...")

	var invalidAccounts []Account
	err = db.Raw(`
		SELECT id, code, name 
		FROM accounts 
		WHERE code !~ '^[0-9]{4}$' 
		  AND code !~ '^[0-9]{4}\.[0-9]+$' 
		  AND deleted_at IS NULL 
		ORDER BY code
		LIMIT 50
	`).Scan(&invalidAccounts).Error

	if err != nil {
		log.Fatalf("❌ Query failed: %v", err)
	}

	if len(invalidAccounts) == 0 {
		log.Println("✅ No invalid account codes found!")
		log.Println("✅ Safe to apply check constraint")
		os.Exit(0)
	}

	log.Printf("⚠️  Found %d accounts with invalid codes:\n", len(invalidAccounts))
	log.Println("=====================================")
	for _, acc := range invalidAccounts {
		log.Printf("ID: %d | Code: '%s' | Name: %s", acc.ID, acc.Code, acc.Name)
	}
	log.Println("=====================================")
	log.Println("\n📝 Action needed:")
	log.Println("1. Fix these account codes manually in the database")
	log.Println("2. Or modify the constraint to allow these formats")
	log.Println("\nValid formats:")
	log.Println("  - 4 digits: 1234")
	log.Println("  - 4 digits + dot + numbers: 1234.01, 1234.123")
}
