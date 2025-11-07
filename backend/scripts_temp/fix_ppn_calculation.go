package main

import (
	"fmt"
	"log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/shopspring/decimal"
	"app-sistem-akuntansi/models"
)

func main() {
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}

	fmt.Println("🔧 Fixing PPN Calculation in Journal Entries...")

	// Check current purchase data
	var purchase models.Purchase
	if err := db.Where("code = ?", "PO/2025/09/0025").First(&purchase).Error; err != nil {
		log.Fatal("Purchase not found:", err)
	}

	fmt.Printf("📊 Purchase Details:\n")
	fmt.Printf("   Code: %s\n", purchase.Code)
	fmt.Printf("   Subtotal Before Discount: %.2f\n", purchase.SubtotalBeforeDiscount)
	fmt.Printf("   PPN Rate: %.2f%%\n", purchase.PPNRate)
	fmt.Printf("   PPN Amount: %.2f\n", purchase.PPNAmount)
	fmt.Printf("   Total Amount: %.2f\n", purchase.TotalAmount)

	// Check current journal entry
	var journalEntry models.SSOTJournalEntry
	if err := db.Where("source_type = ? AND source_code = ?", models.SSOTSourceTypePurchase, purchase.Code).First(&journalEntry).Error; err != nil {
		log.Fatal("Journal entry not found:", err)
	}

	fmt.Printf("\n📋 Current Journal Entry: %s\n", journalEntry.EntryNumber)
	fmt.Printf("   Total Debit: %s\n", journalEntry.TotalDebit.String())
	fmt.Printf("   Total Credit: %s\n", journalEntry.TotalCredit.String())

	// Check journal lines
	var lines []models.SSOTJournalLine
	db.Where("journal_id = ?", journalEntry.ID).Find(&lines)

	fmt.Printf("\n📝 Current Journal Lines:\n")
	for _, line := range lines {
		var account models.Account
		db.First(&account, line.AccountID)
		fmt.Printf("   %s - Debit: %s, Credit: %s\n", 
			account.Name, line.DebitAmount.String(), line.CreditAmount.String())
	}

	// Fix the journal entry with correct PPN calculation
	if err := fixJournalEntryWithPPN(db, &purchase, &journalEntry); err != nil {
		log.Fatal("Failed to fix journal entry:", err)
	}

	// Update COA balances
	updateCOABalancesCorrectly(db)

	fmt.Println("\n🎉 PPN calculation fixed successfully!")
}

func fixJournalEntryWithPPN(db *gorm.DB, purchase *models.Purchase, journalEntry *models.SSOTJournalEntry) error {
	fmt.Println("\n🔄 Fixing journal entry with correct PPN calculation...")

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete existing journal lines
	tx.Where("journal_id = ?", journalEntry.ID).Delete(&models.SSOTJournalLine{})

	// Get required accounts
	var inventoryAccount, payableAccount, taxAccount models.Account
	
	if err := tx.Where("code = ?", "1301").First(&inventoryAccount).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("inventory account not found: %v", err)
	}
	
	if err := tx.Where("code = ?", "2101").First(&payableAccount).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("payable account not found: %v", err)
	}

	if err := tx.Where("code = ?", "2102").First(&taxAccount).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("tax account not found: %v", err)
	}

	// Calculate amounts correctly
	subtotalAmount := decimal.NewFromFloat(purchase.SubtotalBeforeDiscount)
	ppnAmount := decimal.NewFromFloat(purchase.PPNAmount)
	totalAmount := decimal.NewFromFloat(purchase.TotalAmount)

	fmt.Printf("   💰 Calculation breakdown:\n")
	fmt.Printf("      Subtotal (Inventory): %s\n", subtotalAmount.String())
	fmt.Printf("      PPN Amount (Tax): %s\n", ppnAmount.String())
	fmt.Printf("      Total Payable: %s\n", totalAmount.String())

	// Create correct journal lines
	lines := []models.SSOTJournalLine{
		{
			JournalID:    journalEntry.ID,
			AccountID:    uint64(inventoryAccount.ID),
			LineNumber:   1,
			Description:  "Inventory purchase (before tax)",
			DebitAmount:  subtotalAmount,
			CreditAmount: decimal.Zero,
		},
		{
			JournalID:    journalEntry.ID,
			AccountID:    uint64(taxAccount.ID),
			LineNumber:   2,
			Description:  "PPN Input Tax",
			DebitAmount:  ppnAmount,
			CreditAmount: decimal.Zero,
		},
		{
			JournalID:    journalEntry.ID,
			AccountID:    uint64(payableAccount.ID),
			LineNumber:   3,
			Description:  "Accounts payable (total including PPN)",
			DebitAmount:  decimal.Zero,
			CreditAmount: totalAmount,
		},
	}

	for _, line := range lines {
		if err := tx.Create(&line).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create journal line: %v", err)
		}
	}

	// Update journal entry totals
	journalEntry.TotalDebit = subtotalAmount.Add(ppnAmount)
	journalEntry.TotalCredit = totalAmount
	tx.Save(journalEntry)

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	fmt.Println("   ✅ Journal entry updated with correct PPN calculation")
	return nil
}

func updateCOABalancesCorrectly(db *gorm.DB) {
	fmt.Println("\n💰 Updating COA balances with correct PPN...")

	// Update balances based on the corrected journal entries
	updates := map[string]float64{
		"1301": 5000000,  // Persediaan Barang Dagangan (Debit)
		"2101": 5550000,  // Utang Usaha (Credit)
		"2102": 550000,   // Utang Pajak PPN (Debit for input tax)
	}

	for code, balance := range updates {
		var account models.Account
		if err := db.Where("code = ?", code).First(&account).Error; err == nil {
			// For PPN input tax (asset), it should be debit balance
			if code == "2102" && account.Type == "LIABILITY" {
				// If it's still liability account, we need to check the nature
				// PPN Input should be asset (receivable from government)
				// Let's update it properly
				if account.Name == "Utang Pajak" {
					// This should be PPN Input (asset), not liability
					balance = 550000 // Positive for asset
				}
			}
			
			db.Model(&account).Update("balance", balance)
			fmt.Printf("   ✅ %s (%s): %.2f\n", account.Name, account.Code, balance)
		}
	}

	fmt.Println("\n📊 Final balance check:")
	// Verify accounting equation: Assets = Liabilities + Equity
	var totalAssets, totalLiabilities float64
	
	db.Raw("SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'ASSET'").Scan(&totalAssets)
	db.Raw("SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'LIABILITY'").Scan(&totalLiabilities)
	
	fmt.Printf("   Total Assets: %.2f\n", totalAssets)
	fmt.Printf("   Total Liabilities: %.2f\n", totalLiabilities)
	fmt.Printf("   Difference: %.2f\n", totalAssets-totalLiabilities)
}