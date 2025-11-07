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

	fmt.Println("=== DEBUG CASH & BANK BALANCE DISCREPANCY ===")
	
	// 1. Get all cash bank records
	var cashBanks []models.CashBank
	if err := db.Where("deleted_at IS NULL").Find(&cashBanks).Error; err != nil {
		log.Fatal("Failed to get cash banks:", err)
	}

	fmt.Printf("\n1. CASH BANK RECORDS (%d found):\n", len(cashBanks))
	fmt.Println("ID\tCode\t\tName\t\t\tBalance\t\tAccount ID\tActive")
	fmt.Println("--\t----\t\t----\t\t\t-------\t\t----------\t------")
	
	for _, cb := range cashBanks {
	fmt.Printf("%d\t%s\t\t%-15s\t%.2f\t\t%v\t\t%v\n", 
			cb.ID, cb.Code, cb.Name, cb.Balance, cb.AccountID, cb.IsActive)
	}
	
	// 2. Check related accounts in COA for each cash bank
	fmt.Println("\n2. RELATED COA ACCOUNTS:")
	fmt.Println("CashBank\t\tCOA Code\tCOA Name\t\tCOA Balance\tMatch?")
	fmt.Println("--------\t\t--------\t--------\t\t-----------\t------")
	
	for _, cb := range cashBanks {
		if cb.AccountID != 0 {
			var account models.Account
			if err := db.Where("id = ? AND deleted_at IS NULL", cb.AccountID).First(&account).Error; err == nil {
				match := "❌ NO"
				if cb.Balance == account.Balance {
					match = "✅ YES"
				}
				fmt.Printf("%-15s\t%s\t\t%-15s\t%.2f\t\t%s\n", 
					cb.Name, account.Code, account.Name, account.Balance, match)
			} else {
				fmt.Printf("%-15s\tN/A\t\tACCOUNT NOT FOUND\t\t\t❌ ERROR\n", cb.Name)
			}
		} else {
			fmt.Printf("%-15s\tN/A\t\tNO GL ACCOUNT LINKED\t\t\t❌ ERROR\n", cb.Name)
		}
	}

	// 3. Check seed data - look for cashbank seed files
	fmt.Println("\n3. CHECKING FOR SEED DATA...")
	
	// Check if there are any cash bank transactions that might have created this balance
	type CashBankTransaction struct {
		ID       uint    `json:"id"`
		Amount   float64 `json:"amount"`
		Type     string  `json:"type"`
		CashBankID uint  `json:"cash_bank_id"`
	}
	
	var transactions []CashBankTransaction
	db.Table("cash_bank_transactions").
		Where("deleted_at IS NULL").
		Find(&transactions)
	
	fmt.Printf("Found %d cash bank transactions:\n", len(transactions))
	for _, tx := range transactions {
		fmt.Printf("  ID: %d, Amount: %.2f, Type: %s, CashBank: %d\n", 
			tx.ID, tx.Amount, tx.Type, tx.CashBankID)
	}

	// 4. Specific focus on Bank BRI (likely code 1105 based on COA)
	fmt.Println("\n4. BANK BRI ANALYSIS:")
	
	var briBankAccount models.Account
	if err := db.Where("code = '1105' AND deleted_at IS NULL").First(&briBankAccount).Error; err == nil {
		fmt.Printf("COA Account 1105 (Bank BRI): Balance = %.2f\n", briBankAccount.Balance)
	} else {
		fmt.Println("COA Account 1105 (Bank BRI): NOT FOUND")
	}
	
	// Find cash bank record that might be linked to BRI
	var briCashBank models.CashBank
	found := false
	for _, cb := range cashBanks {
		if cb.AccountID != 0 && cb.AccountID == briBankAccount.ID {
			briCashBank = cb
			found = true
			break
		}
		// Also check by name containing "BRI"
		if !found && (cb.Name == "Bank BRI" || cb.Code == "BNK-2025-0003") {
			briCashBank = cb
			found = true
		}
	}
	
	if found {
		fmt.Printf("Cash Bank BRI: Balance = %.2f\n", briCashBank.Balance)
		fmt.Printf("Discrepancy: %.2f (CashBank) vs %.2f (COA)\n", 
			briCashBank.Balance, briBankAccount.Balance)
			
		// Check journal entries for this account
		var journalCount int64
		db.Table("journal_entries").
			Joins("JOIN journals ON journal_entries.journal_id = journals.id").
			Where("journal_entries.account_id = ? AND journals.deleted_at IS NULL", briBankAccount.ID).
			Count(&journalCount)
		
		fmt.Printf("Journal entries for Bank BRI account: %d\n", journalCount)
		
		if journalCount == 0 && briCashBank.Balance > 0 {
			fmt.Println("🚨 ISSUE FOUND: CashBank has balance but no journal entries - likely from seed data!")
		}
	} else {
		fmt.Println("Cash Bank BRI: NOT FOUND")
	}

	// 5. Check all accounts with balance != 0 from accounts table
	fmt.Println("\n5. ALL ACCOUNTS WITH NON-ZERO BALANCE:")
	
	var accountsWithBalance []models.Account
	db.Where("balance != 0 AND deleted_at IS NULL").Find(&accountsWithBalance)
	
	fmt.Printf("Found %d accounts with non-zero balance:\n", len(accountsWithBalance))
	for _, acc := range accountsWithBalance {
		fmt.Printf("  %s (%s): %.2f - Header: %v\n", 
			acc.Code, acc.Name, acc.Balance, acc.IsHeader)
	}

	fmt.Println("\n=== SUMMARY ===")
	fmt.Println("If CashBank has balance but COA account has 0 balance with no journal entries,")
	fmt.Println("this indicates the balance came from seed data and should be removed.")
}
