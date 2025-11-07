package main

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
)

func main() {
	log.Printf("🔍 Checking Specific Sale and Its Journal Entries")
	log.Printf("=================================================")

	// Connect to database
	dsn := "postgres://postgres:postgres@localhost/sistem_akuntans_test?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// Check Sale ID 1 - the one we fixed earlier
	var sale models.Sale
	err = db.Preload("SaleItems.RevenueAccount").
		Preload("Customer").
		Where("id = ?", 1).
		First(&sale).Error
	
	if err != nil {
		log.Printf("❌ Sale not found: %v", err)
		return
	}

	log.Printf("📄 Sale Details:")
	log.Printf("   Sale ID: %d", sale.ID)
	log.Printf("   Code: %s", sale.Code)
	log.Printf("   Status: %s", sale.Status)
	log.Printf("   Customer: %s", sale.Customer.Name)
	log.Printf("   Total Amount: %.2f", sale.TotalAmount)
	log.Printf("   PPN Amount: %.2f", sale.PPNAmount)

	// Get journal entries for this sale
	var journalEntries []models.JournalEntry
	err = db.Preload("JournalLines.Account").
		Where("reference_type = ? AND reference_id = ?", "SALE", sale.ID).
		Find(&journalEntries).Error

	if err != nil {
		log.Printf("❌ Failed to get journal entries: %v", err)
	} else if len(journalEntries) == 0 {
		log.Printf("⚠️ No journal entries found for this sale!")
	} else {
		log.Printf("\n📖 Journal Entries:")
		for i, entry := range journalEntries {
			log.Printf("   Entry %d: %s", i+1, entry.Code)
			log.Printf("     Description: %s", entry.Description)
			log.Printf("     Status: %s", entry.Status)
			log.Printf("     Total Debit: %.2f", entry.TotalDebit)
			log.Printf("     Total Credit: %.2f", entry.TotalCredit)
			
			log.Printf("     Lines:")
			totalDebit := 0.0
			totalCredit := 0.0
			for j, line := range entry.JournalLines {
				log.Printf("       Line %d: %s (%s) - Debit: %.2f, Credit: %.2f", 
					j+1, line.Account.Name, line.Account.Code, 
					line.DebitAmount, line.CreditAmount)
				totalDebit += line.DebitAmount
				totalCredit += line.CreditAmount
			}
			
			log.Printf("     Calculated Totals: Debit %.2f, Credit %.2f", totalDebit, totalCredit)
			if totalDebit == totalCredit {
				log.Printf("     ✅ Journal is balanced")
			} else {
				log.Printf("     ❌ Journal NOT balanced!")
			}
		}
	}

	// Check current account balances
	log.Printf("\n💳 Current Account Balances:")
	
	accountCodes := []string{"1104", "4101", "2103"}
	for _, code := range accountCodes {
		var account models.Account
		if db.Where("code = ?", code).First(&account).Error == nil {
			log.Printf("   %s (%s): %.2f", code, account.Name, account.Balance)
		}
	}

	// Test if we can manually fix the balance calculation by re-running posting
	log.Printf("\n🔧 Testing Balance Calculation Fix...")
	
	if len(journalEntries) > 0 {
		entry := &journalEntries[0]
		log.Printf("Re-processing journal entry %s...", entry.Code)
		
		// Simulate the new balance calculation logic
		for _, line := range entry.JournalLines {
			var account models.Account
			if db.Select("id, code, name, type, balance").First(&account, line.AccountID).Error == nil {
				var balanceChange float64
				if account.Type == "ASSET" || account.Type == "EXPENSE" {
					// Debit normal accounts: debit increases, credit decreases
					balanceChange = line.DebitAmount - line.CreditAmount
				} else {
					// Credit normal accounts (LIABILITY, EQUITY, REVENUE): credit increases, debit decreases
					balanceChange = line.CreditAmount - line.DebitAmount
				}
				
				log.Printf("   Account %s (%s) Type=%s:", account.Code, account.Name, account.Type)
				log.Printf("     Line: Debit %.2f, Credit %.2f", line.DebitAmount, line.CreditAmount)
				log.Printf("     Balance Change: %.2f", balanceChange)
				log.Printf("     New Logic Result: %s account should %s by %.2f", 
					account.Type, 
					func() string { if balanceChange > 0 { return "INCREASE" } else { return "DECREASE" }}(),
					func() float64 { if balanceChange < 0 { return -balanceChange } else { return balanceChange }}())
			}
		}
	}
}