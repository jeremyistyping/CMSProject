package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	fmt.Println("🛠️ FIXING: Revenue and PPN Balance Synchronization")
	fmt.Println("================================================")

	// Load configuration and connect to database
	_ = config.LoadConfig()
	db := database.ConnectDB()

	fmt.Println("\n1. ANALYZING CURRENT STATE:")
	fmt.Println("---------------------------")

	// 1. Get current account balances
	var accounts []models.Account
	db.Where("code IN (?)", []string{"1201", "2103", "4101"}).Order("code").Find(&accounts)

	var arAccount, ppnAccount, revenueAccount models.Account
	for _, acc := range accounts {
		switch acc.Code {
		case "1201":
			arAccount = acc
		case "2103":
			ppnAccount = acc
		case "4101":
			revenueAccount = acc
		}
		fmt.Printf("Account %s (%s): Current Balance = Rp %.2f\n", 
			acc.Code, acc.Name, acc.Balance)
	}

	// 2. Calculate correct balances from journal entries
	fmt.Println("\n2. CALCULATING CORRECT BALANCES FROM JOURNAL ENTRIES:")
	fmt.Println("-----------------------------------------------------")

	type AccountBalance struct {
		AccountID uint64
		Code      string
		Name      string
		Debits    float64
		Credits   float64
		NetBalance float64
		AccountType string
	}

	balanceMap := make(map[uint64]*AccountBalance)

	// Initialize for our key accounts
	balanceMap[uint64(arAccount.ID)] = &AccountBalance{
		AccountID: uint64(arAccount.ID),
		Code: arAccount.Code,
		Name: arAccount.Name,
		AccountType: string(arAccount.Type),
	}
	balanceMap[uint64(ppnAccount.ID)] = &AccountBalance{
		AccountID: uint64(ppnAccount.ID),
		Code: ppnAccount.Code,
		Name: ppnAccount.Name,
		AccountType: string(ppnAccount.Type),
	}
	balanceMap[uint64(revenueAccount.ID)] = &AccountBalance{
		AccountID: uint64(revenueAccount.ID),
		Code: revenueAccount.Code,
		Name: revenueAccount.Name,
		AccountType: string(revenueAccount.Type),
	}

	// Get all journal lines for these accounts
	var journalLines []models.SSOTJournalLine
	accountIDs := []uint64{uint64(arAccount.ID), uint64(ppnAccount.ID), uint64(revenueAccount.ID)}
	
	err := db.Joins("JOIN ssot_journal_entries ON ssot_journal_lines.ssot_journal_entry_id = ssot_journal_entries.id").
		Where("ssot_journal_lines.account_id IN (?) AND ssot_journal_entries.status = ?", accountIDs, "POSTED").
		Find(&journalLines).Error
	
	if err != nil {
		log.Fatalf("Error querying journal lines: %v", err)
	}

	// Calculate totals
	for _, line := range journalLines {
		if balance, exists := balanceMap[line.AccountID]; exists {
			debit := line.DebitAmount.InexactFloat64()
			credit := line.CreditAmount.InexactFloat64()
			
			balance.Debits += debit
			balance.Credits += credit
		}
	}

	// Calculate net balances based on account type
	for _, balance := range balanceMap {
		switch balance.AccountType {
		case "ASSET", "EXPENSE":
			// Assets and Expenses: Debit increases, Credit decreases
			balance.NetBalance = balance.Debits - balance.Credits
		case "LIABILITY", "EQUITY", "REVENUE":
			// Liabilities, Equity, Revenue: Credit increases, Debit decreases
			balance.NetBalance = balance.Credits - balance.Debits
		}
		
		fmt.Printf("Account %s (%s):\n", balance.Code, balance.Name)
		fmt.Printf("  Journal Debits: Rp %.2f\n", balance.Debits)
		fmt.Printf("  Journal Credits: Rp %.2f\n", balance.Credits)
		fmt.Printf("  Calculated Balance: Rp %.2f\n", balance.NetBalance)
		fmt.Println()
	}

	// 3. Check for discrepancies and fix them
	fmt.Println("3. FIXING BALANCE DISCREPANCIES:")
	fmt.Println("--------------------------------")

	balanceFixed := false

	for _, balance := range balanceMap {
		var currentAccount models.Account
		db.First(&currentAccount, balance.AccountID)
		
		currentBalance := currentAccount.Balance
		calculatedBalance := balance.NetBalance
		difference := currentBalance - calculatedBalance

		fmt.Printf("Account %s (%s):\n", balance.Code, balance.Name)
		fmt.Printf("  Current DB Balance: Rp %.2f\n", currentBalance)
		fmt.Printf("  Journal-Calculated: Rp %.2f\n", calculatedBalance)
		fmt.Printf("  Difference: Rp %.2f\n", difference)

		if difference != 0 {
			fmt.Printf("  ❌ FIXING: Updating balance from %.2f to %.2f\n", currentBalance, calculatedBalance)
			
			err := db.Model(&currentAccount).Update("balance", calculatedBalance).Error
			if err != nil {
				log.Printf("Error updating account %s: %v", balance.Code, err)
			} else {
				fmt.Printf("  ✅ FIXED: Account %s balance synchronized\n", balance.Code)
				balanceFixed = true
			}
		} else {
			fmt.Printf("  ✅ OK: Balance is already correct\n")
		}
		fmt.Println()
	}

	// 4. Additional checks and fixes
	fmt.Println("4. ADDITIONAL SYSTEM CHECKS:")
	fmt.Println("-----------------------------")

	// Check if there are any pending sales that should be invoiced
	var draftSales []models.Sale
	db.Where("status = ?", models.SaleStatusDraft).Find(&draftSales)
	
	fmt.Printf("Found %d sales in DRAFT status:\n", len(draftSales))
	for _, sale := range draftSales {
		fmt.Printf("  Sale ID %d (%s): Total=Rp %.2f, PPN=Rp %.2f\n", 
			sale.ID, sale.Code, sale.TotalAmount, sale.PPN)
		
		// Check if this sale has journal entries (shouldn't have any for DRAFT)
		var journalCount int64
		db.Model(&models.SSOTJournalEntry{}).Where("source_type = ? AND source_code = ?", "SALES", sale.Code).Count(&journalCount)
		
		if journalCount > 0 {
			fmt.Printf("    ⚠️  WARNING: DRAFT sale has %d journal entries (should be 0)\n", journalCount)
		}
	}

	// 5. Verify PPN validation will now work
	fmt.Println("\n5. VERIFYING PPN VALIDATION:")
	fmt.Println("-----------------------------")

	// Get updated PPN account balance
	var updatedPPNAccount models.Account
	db.First(&updatedPPNAccount, ppnAccount.ID)
	
	fmt.Printf("Updated PPN Keluaran Balance: Rp %.2f\n", updatedPPNAccount.Balance)

	// Test if a new sale can be processed
	if len(draftSales) > 0 {
		testSale := draftSales[0]
		testPPNAmount := testSale.PPN
		projectedBalance := updatedPPNAccount.Balance + testPPNAmount
		
		fmt.Printf("If Sale ID %d is confirmed:\n", testSale.ID)
		fmt.Printf("  Current PPN Balance: Rp %.2f\n", updatedPPNAccount.Balance)
		fmt.Printf("  Sale PPN Amount: Rp %.2f\n", testPPNAmount)
		fmt.Printf("  Projected PPN Balance: Rp %.2f\n", projectedBalance)
		
		if projectedBalance >= 0 {
			fmt.Printf("  ✅ PPN validation should PASS\n")
		} else {
			fmt.Printf("  ❌ PPN validation will still FAIL (negative balance)\n")
		}
	}

	// 6. Summary
	fmt.Println("\n6. SUMMARY:")
	fmt.Println("-----------")
	
	if balanceFixed {
		fmt.Println("✅ Account balances have been synchronized with journal entries")
		fmt.Println("✅ Revenue discrepancy has been resolved")
		fmt.Println("✅ PPN balance validation should now work correctly")
		fmt.Println("")
		fmt.Println("🔄 NEXT STEPS:")
		fmt.Println("1. Restart your backend application")
		fmt.Println("2. Try confirming Sale ID 4 again")
		fmt.Println("3. Check that the frontend now displays correct revenue amounts")
	} else {
		fmt.Println("✅ All balances were already synchronized")
		fmt.Println("💡 The issue may be elsewhere - check the PPN validation logic")
	}

	fmt.Println("\n🛠️ FIX COMPLETED")
}