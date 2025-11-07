package main

import (
	"app-sistem-akuntansi/models"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	fmt.Println("🔍 Checking Purchase Status")
	fmt.Println("===========================")

	// Database connection
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost/sistem_akuntansi?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Printf("❌ Database connection failed: %v", err)
		return
	}

	fmt.Println("✅ Database connected successfully\n")

	// Query purchases
	var purchases []models.Purchase
	err = db.Find(&purchases).Error
	if err != nil {
		log.Printf("❌ Failed to query purchases: %v", err)
		return
	}

	fmt.Printf("📊 Found %d purchases:\n", len(purchases))
	for i, purchase := range purchases {
		fmt.Printf("  %d. ID: %d, Code: %s, Status: %s, Approval: %s, Amount: Rp %.2f\n", 
			i+1, purchase.ID, purchase.Code, purchase.Status, purchase.ApprovalStatus, purchase.TotalAmount)
	}

	// Check SSOT journal entries
	var journalEntries []models.SSOTJournalEntry
	err = db.Find(&journalEntries).Error
	if err != nil {
		log.Printf("❌ Failed to query journal entries: %v", err)
		return
	}

	fmt.Printf("\n📔 Found %d SSOT journal entries:\n", len(journalEntries))
	for i, entry := range journalEntries {
		fmt.Printf("  %d. ID: %d, Entry#: %s, Source: %s, SourceID: %v, Status: %s\n", 
			i+1, entry.ID, entry.EntryNumber, entry.SourceType, entry.SourceID, entry.Status)
	}

	// Check accounts for purchase COA
	purchaseAccounts := []string{"1301", "1240", "2101", "2111", "2112", "1103"}
	fmt.Printf("\n💰 Purchase-related account balances:\n")
	for _, code := range purchaseAccounts {
		var account models.Account
		err = db.Where("code = ?", code).First(&account).Error
		if err != nil {
			fmt.Printf("  ❌ Account %s: Not found\n", code)
		} else {
			fmt.Printf("  ✅ Account %s (%s): Rp %.2f\n", code, account.Name, account.Balance)
		}
	}
}