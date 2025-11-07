package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	fmt.Println("🛠️ FIXING: Revenue and PPN Balance Synchronization (CORRECTED)")
	fmt.Println("===========================================================")

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

	// 2. Calculate correct balances from journal entries using the actual tables
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

	// Get all journal lines for these accounts from the actual tables
	type JournalLineData struct {
		AccountID    uint64  `json:"account_id"`
		DebitAmount  float64 `json:"debit_amount"`
		CreditAmount float64 `json:"credit_amount"`
	}

	var journalLines []JournalLineData
	accountIDs := []uint64{uint64(arAccount.ID), uint64(ppnAccount.ID), uint64(revenueAccount.ID)}
	
	// Query from the actual journal_lines table with proper joins
	err := db.Raw(`
		SELECT jl.account_id, jl.debit_amount, jl.credit_amount
		FROM journal_lines jl
		JOIN journal_entries je ON jl.journal_entry_id = je.id
		WHERE jl.account_id IN (?) AND je.status = 'POSTED'
	`, accountIDs).Scan(&journalLines).Error
	
	if err != nil {
		log.Printf("Warning: Error querying journal lines from journal_lines table: %v", err)
		
		// Try unified_journal_lines as fallback
		err = db.Raw(`
			SELECT account_id, debit_amount, credit_amount
			FROM unified_journal_lines
			WHERE account_id IN (?)
		`, accountIDs).Scan(&journalLines).Error
		
		if err != nil {
			log.Printf("Error querying from unified_journal_lines: %v", err)
		} else {
			fmt.Println("✅ Using unified_journal_lines table")
		}
	} else {
		fmt.Println("✅ Using journal_lines table")
	}

	fmt.Printf("Found %d journal line entries\n", len(journalLines))

	// Calculate totals
	for _, line := range journalLines {
		if balance, exists := balanceMap[line.AccountID]; exists {
			balance.Debits += line.DebitAmount
			balance.Credits += line.CreditAmount
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

	// 4. Check PPN validation issue specifically
	fmt.Println("4. INVESTIGATING PPN VALIDATION ISSUE:")
	fmt.Println("--------------------------------------")

	// Get updated PPN account balance after fix
	var updatedPPNAccount models.Account
	db.First(&updatedPPNAccount, ppnAccount.ID)
	
	fmt.Printf("Current PPN Keluaran Balance: Rp %.2f\n", updatedPPNAccount.Balance)

	// Check the pending sale that's causing the issue
	var draftSales []models.Sale
	db.Where("status = ?", models.SaleStatusDraft).Find(&draftSales)
	
	if len(draftSales) > 0 {
		problemSale := draftSales[0] // This should be the one failing
		testPPNAmount := problemSale.PPN
		projectedBalance := updatedPPNAccount.Balance + testPPNAmount
		
		fmt.Printf("\nTesting Sale ID %d (%s):\n", problemSale.ID, problemSale.Code)
		fmt.Printf("  Sale Total: Rp %.2f\n", problemSale.TotalAmount)
		fmt.Printf("  Sale PPN: Rp %.2f\n", testPPNAmount)
		fmt.Printf("  Current PPN Balance: Rp %.2f\n", updatedPPNAccount.Balance)
		fmt.Printf("  Projected PPN Balance After Sale: Rp %.2f\n", projectedBalance)
		
		if projectedBalance >= 0 {
			fmt.Printf("  ✅ PPN validation should now PASS\n")
		} else {
			fmt.Printf("  ❌ PPN validation will still FAIL (would result in negative balance)\n")
			fmt.Printf("     This suggests there's an issue with the PPN calculation logic.\n")
			
			// Let's check what's happening in more detail
			fmt.Printf("\n  🔍 DETAILED PPN ANALYSIS:\n")
			fmt.Printf("     The system is trying to prevent PPN account from going negative.\n")
			fmt.Printf("     But the calculation seems to be subtracting instead of adding.\n")
			fmt.Printf("     PPN Keluaran should INCREASE (credit) when making a sale.\n")
			
			// The issue might be in the enhanced service validation
			fmt.Printf("     💡 RECOMMENDATION: Check PPN validation logic in EnhancedSalesJournalService\n")
		}
	}

	// 5. Additional diagnostics
	fmt.Println("\n5. ADDITIONAL SYSTEM CHECKS:")
	fmt.Println("-----------------------------")

	// Check what journal system is actually being used
	var journalEntryCount int64
	db.Model(&models.JournalEntry{}).Count(&journalEntryCount)
	
	var unifiedJournalLineCount int64
	db.Raw("SELECT COUNT(*) FROM unified_journal_lines").Scan(&unifiedJournalLineCount)

	fmt.Printf("Traditional Journal Entries: %d\n", journalEntryCount)
	fmt.Printf("Unified Journal Lines: %d\n", unifiedJournalLineCount)

	if journalEntryCount > 0 {
		fmt.Printf("✅ System is using traditional journal_entries table\n")
	}
	if unifiedJournalLineCount > 0 {
		fmt.Printf("✅ System has unified_journal_lines data\n")
	}

	// 6. Summary and recommendations
	fmt.Println("\n6. SUMMARY & RECOMMENDATIONS:")
	fmt.Println("==============================")
	
	if balanceFixed {
		fmt.Println("✅ Account balances have been synchronized with journal entries")
		fmt.Println("✅ Revenue discrepancy has been resolved")
		
		if updatedPPNAccount.Balance >= 0 {
			fmt.Println("✅ PPN balance is non-negative")
			fmt.Println("")
			fmt.Println("🔄 NEXT STEPS:")
			fmt.Println("1. The issue might be in the PPN validation logic itself")
			fmt.Println("2. Check the EnhancedSalesJournalService validation")
			fmt.Println("3. The error suggests it's calculating a negative balance incorrectly")
		} else {
			fmt.Println("❌ PPN balance is still negative - more investigation needed")
		}
	} else {
		fmt.Println("ℹ️  All balances were already correct")
		fmt.Println("💡 The revenue/PPN issue must be in the validation logic, not the balances")
	}

	fmt.Println("\n🛠️ CORRECTED FIX COMPLETED")
}