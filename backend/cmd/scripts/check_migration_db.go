package main

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type MigrationLog struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	MigrationName string `json:"migration_name"`
	ExecutedAt    string `json:"executed_at"`
	Description   string `json:"description"`
	Status        string `json:"status"`
}

func main() {
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntans_test port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("Connection failed: %v\n", err)
		return
	}

	var migrations []MigrationLog
	result := db.Where("migration_name LIKE ?", "%purchase%").Order("executed_at DESC").Find(&migrations)
	if result.Error != nil {
		fmt.Printf("Query error: %v\n", result.Error)
		return
	}

	fmt.Printf("Found %d purchase-related migrations:\n", len(migrations))
	for _, m := range migrations {
		fmt.Printf("- %s (Status: %s, Date: %s)\n", m.MigrationName, m.Status, m.ExecutedAt)
	}

	// Also check specifically for our migrations
	var count int64
	db.Model(&MigrationLog{}).Where("migration_name IN (?)", []string{
		"021_install_purchase_balance_validation",
		"022_purchase_balance_validation_postgresql", 
		"023_purchase_balance_validation_go_compatible",
		"024_purchase_balance_simple",
		"025_purchase_balance_no_dollar_quotes",
		"026_purchase_balance_minimal",
	}).Count(&count)
	
	fmt.Printf("\nTarget migrations found: %d/6\n", count)
}