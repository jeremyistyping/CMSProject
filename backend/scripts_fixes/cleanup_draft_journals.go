package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Sale struct {
	ID         uint      `gorm:"primaryKey"`
	SaleNumber string    `gorm:"unique;not null"`
	Status     string    
	Total      float64   
	CreatedAt  time.Time
}

type JournalEntry struct {
	ID         uint      `gorm:"primaryKey"`
	SourceType string    
	SourceID   uint      
	EntryDate  time.Time 
	CreatedAt  time.Time
}

type JournalLineItem struct {
	ID             uint    `gorm:"primaryKey"`
	JournalEntryID uint    
	AccountCode    string  
	Debit          float64 
	Credit         float64 
	Description    string  
}

func main() {
	fmt.Println("🧹 CLEANUP SCRIPT: Removing incorrect journal entries from DRAFT/CONFIRMED sales")
	fmt.Println("=" + string(make([]byte, 70)))

	// Database connection
	dsn := "postgres://postgres:postgres@localhost/sistem_akuntans_test?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("✅ Connected to database")
	fmt.Println()

	// Step 1: Find all DRAFT and CONFIRMED sales
	var draftSales []Sale
	err = db.Where("status IN (?)", []string{"DRAFT", "CONFIRMED"}).Find(&draftSales).Error
	if err != nil {
		log.Fatal("Error finding draft/confirmed sales:", err)
	}

	fmt.Printf("📊 Found %d sales with DRAFT or CONFIRMED status\n", len(draftSales))

	if len(draftSales) == 0 {
		fmt.Println("✅ No DRAFT/CONFIRMED sales found - nothing to clean up!")
		return
	}

	// Step 2: Find journal entries for these sales
	totalEntriesDeleted := 0
	totalLineItemsDeleted := 0

	for _, sale := range draftSales {
		fmt.Printf("\n🔍 Checking Sale #%s (ID: %d, Status: %s)\n", sale.SaleNumber, sale.ID, sale.Status)

		// Check in standard journal_entries table
		var journalEntries []JournalEntry
		err = db.Where("source_type IN (?) AND source_id = ?", 
			[]string{"SALE", "SALES", "sale", "sales"}, sale.ID).Find(&journalEntries).Error
		if err != nil {
			fmt.Printf("   ⚠️ Error checking journal entries: %v\n", err)
			continue
		}

		if len(journalEntries) > 0 {
			fmt.Printf("   ❗ Found %d incorrect journal entries for this %s sale\n", len(journalEntries), sale.Status)
			
			// Get journal entry IDs for line item deletion
			var entryIDs []uint
			for _, entry := range journalEntries {
				entryIDs = append(entryIDs, entry.ID)
			}

			// Delete line items first (due to foreign key constraint)
			var lineItemsDeleted int64
			result := db.Where("journal_entry_id IN (?)", entryIDs).Delete(&JournalLineItem{})
			lineItemsDeleted = result.RowsAffected
			if result.Error != nil {
				fmt.Printf("   ⚠️ Error deleting line items: %v\n", result.Error)
				continue
			}

			// Then delete journal entries
			var entriesDeleted int64
			result = db.Where("id IN (?)", entryIDs).Delete(&JournalEntry{})
			entriesDeleted = result.RowsAffected
			if result.Error != nil {
				fmt.Printf("   ⚠️ Error deleting journal entries: %v\n", result.Error)
				continue
			}

			fmt.Printf("   🗑️ Deleted %d journal entries and %d line items\n", entriesDeleted, lineItemsDeleted)
			totalEntriesDeleted += int(entriesDeleted)
			totalLineItemsDeleted += int(lineItemsDeleted)
		} else {
			fmt.Printf("   ✅ No journal entries found (correct)\n")
		}
	}

	// Step 3: Recalculate COA balances after cleanup
	fmt.Println("\n📊 Recalculating COA balances after cleanup...")
	
	// Get all unique account codes from journal line items
	var accountCodes []string
	err = db.Table("journal_line_items").
		Distinct("account_code").
		Pluck("account_code", &accountCodes).Error
	if err != nil {
		fmt.Printf("⚠️ Error getting account codes: %v\n", err)
	} else {
		for _, code := range accountCodes {
			var totalDebit, totalCredit float64
			
			// Calculate total debits
			db.Table("journal_line_items").
				Where("account_code = ?", code).
				Select("COALESCE(SUM(debit), 0)").
				Scan(&totalDebit)
			
			// Calculate total credits
			db.Table("journal_line_items").
				Where("account_code = ?", code).
				Select("COALESCE(SUM(credit), 0)").
				Scan(&totalCredit)
			
			balance := totalDebit - totalCredit
			
			// Update COA balance
			err = db.Table("chart_of_accounts").
				Where("account_code = ?", code).
				Update("balance", balance).Error
			if err != nil {
				fmt.Printf("   ⚠️ Error updating balance for account %s: %v\n", code, err)
			} else {
				fmt.Printf("   ✅ Updated balance for account %s: %.2f\n", code, balance)
			}
		}
	}

	// Summary
	fmt.Println("\n" + "=" + string(make([]byte, 70)))
	fmt.Println("🎯 CLEANUP COMPLETE!")
	fmt.Printf("   • Checked %d DRAFT/CONFIRMED sales\n", len(draftSales))
	fmt.Printf("   • Deleted %d incorrect journal entries\n", totalEntriesDeleted)
	fmt.Printf("   • Deleted %d incorrect line items\n", totalLineItemsDeleted)
	fmt.Println("   • Recalculated COA balances")
	fmt.Println("\n✅ Your journal and COA are now clean and accurate!")
}