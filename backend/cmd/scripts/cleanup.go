package main

import (
	"fmt"
	"log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"app-sistem-akuntansi/models"
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
	
	// Check for accounts with code '1008' including soft deleted ones
	var accounts []models.Account
	result := db.Unscoped().Where("code = ?", "1008").Find(&accounts)
	if result.Error != nil {
		log.Fatalf("Error querying accounts: %v", result.Error)
	}

	fmt.Printf("Found %d accounts with code '1008':\n", len(accounts))
	for _, acc := range accounts {
		fmt.Printf("ID: %d, Code: %s, Name: %s, DeletedAt: %v\n", 
			acc.ID, acc.Code, acc.Name, acc.DeletedAt)
	}

	// Delete soft deleted records permanently
	result = db.Unscoped().Where("code = ? AND deleted_at IS NOT NULL", "1008").Delete(&models.Account{})
	if result.Error != nil {
		log.Printf("Error permanently deleting soft deleted accounts: %v", result.Error)
	} else {
		fmt.Printf("Permanently deleted %d soft deleted accounts with code '1008'\n", result.RowsAffected)
	}

	// Check remaining records
	var remainingAccounts []models.Account
	result = db.Unscoped().Where("code = ?", "1008").Find(&remainingAccounts)
	if result.Error != nil {
		log.Fatalf("Error querying remaining accounts: %v", result.Error)
	}

	fmt.Printf("Remaining accounts with code '1008': %d\n", len(remainingAccounts))
	for _, acc := range remainingAccounts {
		fmt.Printf("ID: %d, Code: %s, Name: %s, DeletedAt: %v\n", 
			acc.ID, acc.Code, acc.Name, acc.DeletedAt)
	}

	// Check for duplicate codes across all accounts
	type CodeCount struct {
		Code  string
		Count int64
	}
	
	var duplicates []CodeCount
	result = db.Model(&models.Account{}).
		Select("code, COUNT(*) as count").
		Group("code").
		Having("COUNT(*) > 1").
		Find(&duplicates)
	
	if result.Error != nil {
		log.Printf("Error checking for duplicates: %v", result.Error)
	} else {
		fmt.Printf("Found %d duplicate codes:\n", len(duplicates))
		for _, dup := range duplicates {
			fmt.Printf("Code: %s, Count: %d\n", dup.Code, dup.Count)
		}
	}

	fmt.Println("Database cleanup completed")
}
