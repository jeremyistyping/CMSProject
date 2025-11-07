package main

import (
	"fmt"
	"log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Database connection
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("Connected to database successfully")
	
	// First, drop the existing unique constraint on code
	fmt.Println("Dropping existing unique constraint on accounts.code...")
	err = db.Exec("ALTER TABLE accounts DROP CONSTRAINT IF EXISTS accounts_code_key;").Error
	if err != nil {
		log.Printf("Warning: Failed to drop existing constraint: %v", err)
	}

	// Create a partial unique index that only applies to non-deleted records
	fmt.Println("Creating partial unique index for non-deleted accounts...")
	err = db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_accounts_code_unique ON accounts (code) WHERE deleted_at IS NULL;").Error
	if err != nil {
		log.Fatalf("Failed to create partial unique index: %v", err)
	}

	fmt.Println("✅ Successfully fixed unique constraint!")
	fmt.Println("Now soft-deleted accounts won't prevent creating new accounts with the same code.")
}
