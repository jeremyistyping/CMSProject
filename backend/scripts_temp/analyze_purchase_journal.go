package main

import (
	"fmt"
	"log"
	"strings"
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

	fmt.Println("🔍 ANALYZING PURCHASE JOURNAL TRANSACTIONS & COA BALANCE")
	fmt.Println(strings.Repeat("=", 70))

	// 1. Get all purchases
	var purchases []models.Purchase
	db.Where("deleted_at IS NULL").Order("created_at").Find(&purchases)

	fmt.Printf("\n📊 PURCHASE SUMMARY:\n")
	fmt.Printf("Total Purchases: %d\n", len(purchases))

	totalValue := float64(0)
	totalApprovedValue := float64(0)
	totalPaidValue := float64(0)
	outstandingAmount := float64(0)

	for _, purchase := range purchases {
		fmt.Printf("\n🛒 Purchase: %s\n", purchase.Code)
		// Get vendor name
		var vendor models.Contact
		db.First(&vendor, purchase.VendorID)
		fmt.Printf("   Vendor: %s\n", vendor.Name)
		fmt.Printf("   Date: %s\n", purchase.Date.Format("02/01/2006"))
		fmt.Printf("   Status: %s | Approval: %s\n", purchase.Status, purchase.ApprovalStatus)
		fmt.Printf("   Subtotal: %.2f | PPN: %.2f | Total: %.2f\n", 
			purchase.SubtotalBeforeDiscount, purchase.PPNAmount, purchase.TotalAmount)
		fmt.Printf("   Paid: %.2f | Outstanding: %.2f\n", 
			purchase.PaidAmount, purchase.TotalAmount - purchase.PaidAmount)

		totalValue += purchase.TotalAmount
		if purchase.ApprovalStatus == "APPROVED" {
			totalApprovedValue += purchase.TotalAmount
		}
		totalPaidValue += purchase.PaidAmount
		outstandingAmount += (purchase.TotalAmount - purchase.PaidAmount)
	}

	fmt.Printf("\n💰 PURCHASE TOTALS:\n")
	fmt.Printf("   Total Value: Rp %.2f\n", totalValue)
	fmt.Printf("   Total Approved: Rp %.2f\n", totalApprovedValue)
	fmt.Printf("   Total Paid: Rp %.2f\n", totalPaidValue)
	fmt.Printf("   Outstanding Amount: Rp %.2f\n", outstandingAmount)

	// 2. Analyze Journal Entries for Purchases
	fmt.Printf("\n📋 JOURNAL ENTRIES ANALYSIS:\n")
	
	var journalEntries []models.SSOTJournalEntry
	db.Where("source_type = ? AND deleted_at IS NULL", models.SSOTSourceTypePurchase).
		Order("created_at").Find(&journalEntries)

	fmt.Printf("Total Journal Entries for Purchases: %d\n", len(journalEntries))

	for _, je := range journalEntries {
		fmt.Printf("\n📝 Journal Entry: %s\n", je.EntryNumber)
		fmt.Printf("   Source: %s\n", je.SourceCode)
		fmt.Printf("   Date: %s | Status: %s\n", 
			je.EntryDate.Format("02/01/2006"), je.Status)
		fmt.Printf("   Total Debit: %s | Total Credit: %s\n", 
			je.TotalDebit.String(), je.TotalCredit.String())

		// Get journal lines
		var lines []models.SSOTJournalLine
		db.Where("journal_id = ?", je.ID).Order("line_number").Find(&lines)

		fmt.Printf("   Journal Lines:\n")
		for _, line := range lines {
			var account models.Account
			db.First(&account, line.AccountID)
			
			fmt.Printf("     %d. %s (%s) - %s\n", 
				line.LineNumber, account.Name, account.Code, account.Type)
			fmt.Printf("        Debit: %s | Credit: %s\n", 
				line.DebitAmount.String(), line.CreditAmount.String())
			fmt.Printf("        Description: %s\n", line.Description)
		}
	}

	// 3. Calculate expected account balances from journal entries
	fmt.Printf("\n🧮 CALCULATING EXPECTED BALANCES FROM JOURNAL ENTRIES:\n")
	
	accountBalances := make(map[uint]float64)
	accountInfo := make(map[uint]models.Account)

	// Process all journal lines to calculate balances
	var allLines []models.SSOTJournalLine
	db.Joins("JOIN unified_journal_ledger ON unified_journal_lines.journal_id = unified_journal_ledger.id").
		Where("unified_journal_ledger.status = 'POSTED' AND unified_journal_ledger.deleted_at IS NULL").
		Find(&allLines)

	for _, line := range allLines {
		if _, exists := accountInfo[uint(line.AccountID)]; !exists {
			var account models.Account
			db.First(&account, line.AccountID)
			accountInfo[uint(line.AccountID)] = account
		}

		account := accountInfo[uint(line.AccountID)]
		debitFloat, _ := line.DebitAmount.Float64()
		creditFloat, _ := line.CreditAmount.Float64()

		// Apply accounting rules
		switch account.Type {
		case "ASSET", "EXPENSE":
			// Assets and Expenses: Debit increases, Credit decreases
			accountBalances[uint(line.AccountID)] += debitFloat - creditFloat
		case "LIABILITY", "EQUITY", "REVENUE":
			// Liabilities, Equity, Revenue: Credit increases, Debit decreases
			accountBalances[uint(line.AccountID)] += creditFloat - debitFloat
		}
	}

	// 4. Compare with current COA balances
	fmt.Printf("\n🏦 COA BALANCE VERIFICATION:\n")
	fmt.Printf("%-8s %-30s %-12s %-15s %-15s %-10s\n", 
		"CODE", "ACCOUNT NAME", "TYPE", "CURRENT BAL", "EXPECTED BAL", "STATUS")
	fmt.Println(strings.Repeat("-", 90))

	var accounts []models.Account
	db.Where("deleted_at IS NULL AND balance != 0").Order("code").Find(&accounts)

	balanceMismatch := false
	for _, account := range accounts {
		expectedBalance := accountBalances[account.ID]
		status := "✅ MATCH"
		
		if account.Balance != expectedBalance {
			status = "❌ MISMATCH"
			balanceMismatch = true
		}

		fmt.Printf("%-8s %-30s %-12s %-15.2f %-15.2f %s\n",
			account.Code, 
			account.Name,
			account.Type,
			account.Balance,
			expectedBalance,
			status)
	}

	// 5. Check accounts with expected balance but zero current balance
	fmt.Printf("\n🔍 ACCOUNTS WITH EXPECTED BALANCE BUT ZERO CURRENT:\n")
	for accountID, expectedBal := range accountBalances {
		if expectedBal != 0 {
			account := accountInfo[accountID]
			if account.Balance == 0 {
				fmt.Printf("%-8s %-30s %-12s %-15.2f %-15.2f %s\n",
					account.Code, 
					account.Name,
					account.Type,
					account.Balance,
					expectedBal,
					"⚠️ MISSING")
				balanceMismatch = true
			}
		}
	}

	// 6. Verify accounting equation
	fmt.Printf("\n📊 ACCOUNTING EQUATION VERIFICATION:\n")
	
	var assetsTotal, liabilitiesTotal, equityTotal float64
	db.Raw("SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'ASSET' AND deleted_at IS NULL").Scan(&assetsTotal)
	db.Raw("SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'LIABILITY' AND deleted_at IS NULL").Scan(&liabilitiesTotal)
	db.Raw("SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'EQUITY' AND deleted_at IS NULL").Scan(&equityTotal)
	
	fmt.Printf("Assets: Rp %.2f\n", assetsTotal)
	fmt.Printf("Liabilities: Rp %.2f\n", liabilitiesTotal)
	fmt.Printf("Equity: Rp %.2f\n", equityTotal)
	fmt.Printf("Assets = Liabilities + Equity? %t\n", assetsTotal == liabilitiesTotal + equityTotal)
	fmt.Printf("Difference: Rp %.2f\n", assetsTotal - (liabilitiesTotal + equityTotal))

	// 7. Summary and recommendations
	fmt.Printf("\n%s\n", strings.Repeat("=", 70))
	fmt.Printf("📋 ANALYSIS SUMMARY:\n")
	
	if !balanceMismatch {
		fmt.Printf("✅ All account balances match journal entries\n")
	} else {
		fmt.Printf("❌ Balance mismatches detected - COA needs adjustment\n")
		fmt.Printf("💡 Recommendation: Run balance recalculation script\n")
	}

	if assetsTotal == liabilitiesTotal + equityTotal {
		fmt.Printf("✅ Accounting equation is balanced\n")
	} else {
		fmt.Printf("❌ Accounting equation is NOT balanced\n")
		fmt.Printf("💡 Recommendation: Review journal entries and fix imbalances\n")
	}

	fmt.Printf("\n🎯 KEY PURCHASE ACCOUNT BALANCES:\n")
	keyAccounts := []string{"1301", "2101", "2102"} // Inventory, Payable, Tax
	for _, code := range keyAccounts {
		var account models.Account
		if err := db.Where("code = ?", code).First(&account).Error; err == nil {
			expectedBal := accountBalances[account.ID]
			fmt.Printf("   %s (%s): Current %.2f | Expected %.2f\n", 
				account.Name, account.Code, account.Balance, expectedBal)
		}
	}
}