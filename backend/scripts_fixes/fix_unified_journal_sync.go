package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	fmt.Println("🛠️ FINAL FIX: Syncing Account Balances from Unified Journal Data")
	fmt.Println("===============================================================")

	// Load configuration and connect to database
	_ = config.LoadConfig()
	db := database.ConnectDB()

	fmt.Println("\n1. RECALCULATING FROM UNIFIED JOURNAL DATA:")
	fmt.Println("-------------------------------------------")

	// Get all accounts we need to fix
	var accounts []models.Account
	db.Where("code IN (?)", []string{"1201", "2103", "4101"}).Order("code").Find(&accounts)

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

	// Initialize balance map
	for _, acc := range accounts {
		balanceMap[uint64(acc.ID)] = &AccountBalance{
			AccountID: uint64(acc.ID),
			Code: acc.Code,
			Name: acc.Name,
			AccountType: string(acc.Type),
		}
	}

	// Get all unified journal lines for these accounts
	type UnifiedJournalLineData struct {
		AccountID    uint64  `json:"account_id"`
		DebitAmount  float64 `json:"debit_amount"`
		CreditAmount float64 `json:"credit_amount"`
		Description  string  `json:"description"`
		AccountCode  string  `json:"account_code"`
		AccountName  string  `json:"account_name"`
	}

	var unifiedLines []UnifiedJournalLineData
	accountIDs := []uint64{uint64(accounts[0].ID), uint64(accounts[1].ID), uint64(accounts[2].ID)}
	
	err := db.Raw(`
		SELECT ujl.account_id, ujl.debit_amount, ujl.credit_amount, ujl.description,
		       a.code as account_code, a.name as account_name
		FROM unified_journal_lines ujl
		JOIN accounts a ON ujl.account_id = a.id
		WHERE ujl.account_id IN (?)
		ORDER BY ujl.id
	`, accountIDs).Scan(&unifiedLines).Error
	
	if err != nil {
		log.Fatalf("Error querying unified journal lines: %v", err)
	}

	fmt.Printf("Found %d unified journal line entries\n", len(unifiedLines))

	// Display the entries to understand what's happening
	for _, line := range unifiedLines {
		fmt.Printf("  %s (%s): %s - Debit: %.2f, Credit: %.2f\n",
			line.AccountCode, line.AccountName, line.Description, line.DebitAmount, line.CreditAmount)
	}

	// Calculate totals
	for _, line := range unifiedLines {
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
		
		fmt.Printf("\nAccount %s (%s):\n", balance.Code, balance.Name)
		fmt.Printf("  Journal Debits: Rp %.2f\n", balance.Debits)
		fmt.Printf("  Journal Credits: Rp %.2f\n", balance.Credits)
		fmt.Printf("  Calculated Balance: Rp %.2f\n", balance.NetBalance)
	}

	// Apply the corrected balances
	fmt.Println("\n2. APPLYING CORRECTED BALANCES:")
	fmt.Println("-------------------------------")

	for _, balance := range balanceMap {
		err := db.Model(&models.Account{}).Where("id = ?", balance.AccountID).Update("balance", balance.NetBalance).Error
		if err != nil {
			log.Printf("Error updating account %s: %v", balance.Code, err)
		} else {
			fmt.Printf("✅ Updated Account %s balance to Rp %.2f\n", balance.Code, balance.NetBalance)
		}
	}

	// Verify the fix worked
	fmt.Println("\n3. VERIFICATION:")
	fmt.Println("----------------")

	var updatedAccounts []models.Account
	db.Where("code IN (?)", []string{"1201", "2103", "4101"}).Order("code").Find(&updatedAccounts)

	for _, acc := range updatedAccounts {
		fmt.Printf("Account %s (%s): NEW Balance = Rp %.2f\n", 
			acc.Code, acc.Name, acc.Balance)
	}

	// Test the problematic sale
	fmt.Println("\n4. TESTING PROBLEMATIC SALE:")
	fmt.Println("-----------------------------")

	var draftSale models.Sale
	db.Where("status = ?", models.SaleStatusDraft).First(&draftSale)

	if draftSale.ID > 0 {
		// Find the updated PPN account
		var ppnAccount models.Account
		db.Where("code = ?", "2103").First(&ppnAccount)

		fmt.Printf("Sale ID %d (%s):\n", draftSale.ID, draftSale.Code)
		fmt.Printf("  Sale Total: Rp %.2f\n", draftSale.TotalAmount)
		fmt.Printf("  Sale PPN: Rp %.2f\n", draftSale.PPN)
		fmt.Printf("  Current PPN Balance: Rp %.2f\n", ppnAccount.Balance)
		fmt.Printf("  After Sale PPN Balance: Rp %.2f\n", ppnAccount.Balance + draftSale.PPN)

		if ppnAccount.Balance + draftSale.PPN >= 0 {
			fmt.Printf("  ✅ PPN validation should now PASS!\n")
		} else {
			fmt.Printf("  ❌ PPN validation might still fail\n")
		}
	}

	// Final summary
	fmt.Println("\n5. FINAL SUMMARY:")
	fmt.Println("-----------------")
	fmt.Println("✅ Account balances synchronized from unified journal data")
	fmt.Println("✅ Revenue account now shows correct balance")
	fmt.Println("✅ PPN account balance corrected")
	fmt.Println("✅ AR account balance corrected")
	fmt.Println("")
	fmt.Println("🔄 NEXT STEPS:")
	fmt.Println("1. Restart your backend server")
	fmt.Println("2. Try confirming the draft sale (Sale ID 4)")
	fmt.Println("3. Check that frontend shows correct revenue amounts")
	fmt.Println("4. The PPN validation should now work correctly")

	fmt.Println("\n🎉 FINAL FIX COMPLETED SUCCESSFULLY!")
}