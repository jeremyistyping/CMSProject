package main

import (
	"fmt"
	"log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
)

func main() {
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}

	fmt.Println("🔍 Verifying PPN Fix Results...")

	// Verify journal entry for purchase PO/2025/09/0025
	var journalEntry models.SSOTJournalEntry
	if err := db.Where("source_type = ? AND source_code = ?", models.SSOTSourceTypePurchase, "PO/2025/09/0025").First(&journalEntry).Error; err != nil {
		log.Fatal("Journal entry not found:", err)
	}

	fmt.Printf("\n📋 Journal Entry: %s\n", journalEntry.EntryNumber)
	fmt.Printf("   Source: %s - %s\n", journalEntry.SourceType, journalEntry.SourceCode)
	fmt.Printf("   Total Debit: %s\n", journalEntry.TotalDebit.String())
	fmt.Printf("   Total Credit: %s\n", journalEntry.TotalCredit.String())
	fmt.Printf("   Status: %s\n", journalEntry.Status)

	// Get journal lines with account details
	var lines []models.SSOTJournalLine
	db.Where("journal_id = ?", journalEntry.ID).Order("line_number").Find(&lines)

	fmt.Printf("\n📝 Journal Lines:\n")
	totalDebit := float64(0)
	totalCredit := float64(0)
	
	for _, line := range lines {
		var account models.Account
		db.First(&account, line.AccountID)
		
		debitFloat, _ := line.DebitAmount.Float64()
		creditFloat, _ := line.CreditAmount.Float64()
		totalDebit += debitFloat
		totalCredit += creditFloat
		
		fmt.Printf("   Line %d: %s (%s)\n", line.LineNumber, account.Name, account.Code)
		fmt.Printf("           %s\n", line.Description)
		fmt.Printf("           Debit: %s, Credit: %s\n", line.DebitAmount.String(), line.CreditAmount.String())
		fmt.Printf("           Account Type: %s, Balance: %.2f\n\n", account.Type, account.Balance)
	}

	fmt.Printf("💰 Totals Verification:\n")
	fmt.Printf("   Total Debit: %.2f\n", totalDebit)
	fmt.Printf("   Total Credit: %.2f\n", totalCredit)
	fmt.Printf("   Balance: %.2f\n", totalDebit - totalCredit)

	// Verify the accounting equation
	fmt.Printf("\n📊 Accounting Equation Check:\n")
	
	var assetsTotal, liabilitiesTotal, equityTotal float64
	db.Raw("SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'ASSET' AND deleted_at IS NULL").Scan(&assetsTotal)
	db.Raw("SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'LIABILITY' AND deleted_at IS NULL").Scan(&liabilitiesTotal)
	db.Raw("SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'EQUITY' AND deleted_at IS NULL").Scan(&equityTotal)
	
	fmt.Printf("   Assets: %.2f\n", assetsTotal)
	fmt.Printf("   Liabilities: %.2f\n", liabilitiesTotal)
	fmt.Printf("   Equity: %.2f\n", equityTotal)
	fmt.Printf("   Assets = Liabilities + Equity? %t (%.2f)\n", 
		assetsTotal == liabilitiesTotal + equityTotal, 
		assetsTotal - (liabilitiesTotal + equityTotal))

	// Check specific accounts affected by the purchase
	fmt.Printf("\n🏦 Key Account Balances:\n")
	checkAccounts := []string{"1301", "2101", "2102"}
	accountNames := map[string]string{
		"1301": "Persediaan Barang Dagangan",
		"2101": "Utang Usaha", 
		"2102": "Utang Pajak",
	}
	
	for _, code := range checkAccounts {
		var account models.Account
		if err := db.Where("code = ?", code).First(&account).Error; err == nil {
			fmt.Printf("   %s (%s): %.2f [%s]\n", 
				accountNames[code], account.Code, account.Balance, account.Type)
		}
	}

	// Verify PPN calculation is correct
	fmt.Printf("\n🧮 PPN Calculation Verification:\n")
	var purchase models.Purchase
	if err := db.Where("code = ?", "PO/2025/09/0025").First(&purchase).Error; err == nil {
		expectedPPN := purchase.SubtotalBeforeDiscount * purchase.PPNRate / 100
		fmt.Printf("   Subtotal: %.2f\n", purchase.SubtotalBeforeDiscount)
		fmt.Printf("   PPN Rate: %.2f%%\n", purchase.PPNRate)
		fmt.Printf("   Expected PPN: %.2f\n", expectedPPN)
		fmt.Printf("   Actual PPN: %.2f\n", purchase.PPNAmount)
		fmt.Printf("   PPN Calculation Correct: %t\n", expectedPPN == purchase.PPNAmount)
		fmt.Printf("   Total Amount: %.2f\n", purchase.TotalAmount)
		fmt.Printf("   Total Calculation Correct: %t\n", 
			purchase.SubtotalBeforeDiscount + purchase.PPNAmount == purchase.TotalAmount)
	}

	fmt.Println("\n✅ PPN verification completed!")
}