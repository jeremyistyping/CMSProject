package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	db, err := connectDB()
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Simple count
	var count int64
	db.Table("unified_journal_lines").Count(&count)
	fmt.Printf("Total rows in unified_journal_lines: %d\n", count)

	// Check table structure
	var columns []struct {
		ColumnName string `gorm:"column:column_name"`
	}
	db.Raw("SELECT column_name FROM information_schema.columns WHERE table_name = 'unified_journal_lines' ORDER BY ordinal_position").Scan(&columns)
	
	fmt.Println("\nTable structure:")
	for _, col := range columns {
		fmt.Printf("  - %s\n", col.ColumnName)
	}

	// If empty, check why
	if count == 0 {
		fmt.Println("\n⚠️  WARNING: unified_journal_lines table is EMPTY!")
		fmt.Println("This explains why General Ledger shows no data.")
		fmt.Println("\nJournal entries exist but their line items (debit/credit details) are missing.")
	}
}

func connectDB() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	}
	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
}

