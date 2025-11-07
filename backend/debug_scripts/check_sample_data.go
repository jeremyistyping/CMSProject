package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	// Connect to database
	db := database.ConnectDB()
	if db == nil {
		log.Fatalf("Failed to connect to database")
	}

	fmt.Println("📊 Checking Sample Data Counts")
	fmt.Println("==============================")

	// Check sales
	var salesCount int64
	db.Model(&models.Sale{}).Count(&salesCount)
	fmt.Printf("💼 Sales: %d records\n", salesCount)

	// Check purchases
	var purchasesCount int64
	db.Model(&models.Purchase{}).Count(&purchasesCount)
	fmt.Printf("🛒 Purchases: %d records\n", purchasesCount)

	// Check journal entries
	var journalEntriesCount int64
	db.Model(&models.JournalEntry{}).Count(&journalEntriesCount)
	fmt.Printf("📝 Journal Entries: %d records\n", journalEntriesCount)

	// Check journal lines
	var journalLinesCount int64
	db.Model(&models.JournalLine{}).Count(&journalLinesCount)
	fmt.Printf("📋 Journal Lines: %d records\n", journalLinesCount)

	// Check posted journal entries
	var postedJournalCount int64
	db.Model(&models.JournalEntry{}).Where("status = ?", models.JournalStatusPosted).Count(&postedJournalCount)
	fmt.Printf("✅ Posted Journal Entries: %d records\n", postedJournalCount)

	fmt.Println("\n🎯 Sample data is ready for testing reports!")
}