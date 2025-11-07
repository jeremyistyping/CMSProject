package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
	"github.com/shopspring/decimal"
)

func main() {
	// Database connection
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntans_test port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("=== MANUAL FIX UNTUK ENTRY 1 ===")
	
	// Check current state of Entry 1
	var lines []models.SSOTJournalLine
	db.Where("journal_id = ?", 1).Find(&lines)
	
	fmt.Printf("Entry 1 current lines:\n")
	totalDebit := float64(0)
	totalCredit := float64(0)
	
	for _, line := range lines {
		debit := line.DebitAmount.InexactFloat64()
		credit := line.CreditAmount.InexactFloat64()
		totalDebit += debit
		totalCredit += credit
		
		var account models.Account
		db.First(&account, line.AccountID)
		
		fmt.Printf("   Line %d: %s - Debit=%.0f, Credit=%.0f\n", 
			line.LineNumber, account.Code, debit, credit)
	}
	
	fmt.Printf("Total: Debit=%.0f, Credit=%.0f (difference: %.0f)\n", 
		totalDebit, totalCredit, totalDebit - totalCredit)
	
	// Check if we already have line number 3 (from previous attempt)
	var existingLine3 models.SSOTJournalLine
	result := db.Where("journal_id = ? AND line_number = ?", 1, 3).First(&existingLine3)
	
	if result.Error == gorm.ErrRecordNotFound {
		// Add the missing revenue line with correct line number
		var revenueAccount models.Account
		db.Where("code = ?", "4101").First(&revenueAccount)
		
		// Find the next available line number
		var maxLineNum int
		db.Model(&models.SSOTJournalLine{}).
			Where("journal_id = ?", 1).
			Select("COALESCE(MAX(line_number), 0)").
			Scan(&maxLineNum)
		
		newLineNum := maxLineNum + 1
		
		newLine := models.SSOTJournalLine{
			JournalID:    1,
			AccountID:    uint64(revenueAccount.ID),
			LineNumber:   newLineNum,
			Description:  "Sales Revenue - Invoice INV/2025/09/0002 [MANUAL FIX]",
			DebitAmount:  decimal.NewFromFloat(0),
			CreditAmount: decimal.NewFromFloat(5000000), // Missing credit amount
		}
		
		if err := db.Create(&newLine).Error; err != nil {
			log.Fatal("Failed to create new line:", err)
		}
		
		fmt.Printf("✅ Added revenue line %d: Credit 5,000,000 to %s\n", newLineNum, revenueAccount.Code)
		
	} else if result.Error == nil {
		fmt.Printf("⚠️  Line 3 already exists, checking if it's correct...\n")
		
		var account models.Account
		db.First(&account, existingLine3.AccountID)
		
		credit := existingLine3.CreditAmount.InexactFloat64()
		fmt.Printf("   Line 3: %s - Credit=%.0f\n", account.Code, credit)
		
		// Check if the amount is correct (should be 5,000,000)
		if credit != 5000000 {
			// Update the amount
			db.Model(&existingLine3).Update("credit_amount", decimal.NewFromFloat(5000000))
			fmt.Printf("✅ Updated line 3 credit amount to 5,000,000\n")
		} else {
			fmt.Printf("✅ Line 3 already has correct amount\n")
		}
	} else {
		log.Fatal("Error checking existing line:", result.Error)
	}
	
	// Verify the fix
	fmt.Printf("\n=== VERIFICATION ===\n")
	
	var verifyLines []models.SSOTJournalLine
	db.Where("journal_id = ?", 1).Find(&verifyLines)
	
	verifyDebit := float64(0)
	verifyCredit := float64(0)
	
	fmt.Printf("Entry 1 after fix:\n")
	for _, line := range verifyLines {
		debit := line.DebitAmount.InexactFloat64()
		credit := line.CreditAmount.InexactFloat64()
		verifyDebit += debit
		verifyCredit += credit
		
		var account models.Account
		db.First(&account, line.AccountID)
		
		fmt.Printf("   Line %d: %s - Debit=%.0f, Credit=%.0f - %s\n", 
			line.LineNumber, account.Code, debit, credit, line.Description)
	}
	
	fmt.Printf("TOTAL: Debit=%.0f, Credit=%.0f", verifyDebit, verifyCredit)
	
	if verifyDebit == verifyCredit {
		fmt.Printf(" ✅ BALANCED!\n")
	} else {
		fmt.Printf(" ❌ STILL UNBALANCED (diff: %.0f)\n", verifyDebit - verifyCredit)
	}
}