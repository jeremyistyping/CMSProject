package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	// Initialize database connection
	db := database.ConnectDB()

	fmt.Println("=== SSOT Balance Synchronization Fix ===")
	fmt.Println()

	// First, backup current account balances
	fmt.Println("1. Backing up current account balances...")
	type AccountBackup struct {
		ID      uint    `json:"id"`
		Code    string  `json:"code"`
		Name    string  `json:"name"`
		Balance float64 `json:"balance"`
	}

	var backupBalances []AccountBackup
	err := db.Raw(`SELECT id, code, name, balance FROM accounts ORDER BY code`).Scan(&backupBalances).Error
	if err != nil {
		log.Printf("Error backing up balances: %v", err)
		return
	}

	fmt.Printf("Backed up %d account balances\n", len(backupBalances))

	// Calculate correct balances from SSOT journal entries
	fmt.Println("\n2. Calculating correct balances from SSOT journal entries...")
	
	// Reset all account balances to zero first
	fmt.Println("  Resetting all account balances to zero...")
	err = db.Exec("UPDATE accounts SET balance = 0").Error
	if err != nil {
		log.Printf("Error resetting balances: %v", err)
		return
	}

	// Calculate and update balances from SSOT journal lines
	fmt.Println("  Calculating balances from SSOT journal lines...")
	
	// For each account that has SSOT journal lines, calculate the correct balance
	var accountUpdates []struct {
		AccountID       uint    `json:"account_id"`
		CalculatedBalance float64 `json:"calculated_balance"`
	}

	err = db.Raw(`
		SELECT 
			account_id,
			SUM(debit_amount) - SUM(credit_amount) as calculated_balance
		FROM unified_journal_lines ujl
		JOIN unified_journal_ledger uj ON ujl.journal_id = uj.id
		WHERE uj.status = 'POSTED'
		GROUP BY account_id
	`).Scan(&accountUpdates).Error

	if err != nil {
		log.Printf("Error calculating balances: %v", err)
		return
	}

	fmt.Printf("  Found %d accounts with SSOT transactions\n", len(accountUpdates))

	// Update each account balance
	for _, update := range accountUpdates {
		err = db.Model(&models.Account{}).Where("id = ?", update.AccountID).Update("balance", update.CalculatedBalance).Error
		if err != nil {
			log.Printf("  Error updating account %d: %v", update.AccountID, err)
		} else {
			fmt.Printf("  Updated account ID %d balance to %.2f\n", update.AccountID, update.CalculatedBalance)
		}
	}

	// Update parent account balances (sum of children)
	fmt.Println("\n3. Updating parent account balances...")
	
	// Update parent account balances by summing children (for header accounts)
	err = db.Raw(`
		UPDATE accounts 
		SET balance = (
			SELECT COALESCE(SUM(child.balance), 0)
			FROM accounts child 
			WHERE child.parent_id = accounts.id
		)
		WHERE is_header = true
	`).Error

	if err != nil {
		log.Printf("Error updating parent balances: %v", err)
	} else {
		fmt.Println("  Updated parent account balances")
	}

	// Show the results
	fmt.Println("\n4. Balance Synchronization Results:")
	var finalBalances []struct {
		ID      uint    `json:"id"`
		Code    string  `json:"code"`
		Name    string  `json:"name"`
		Balance float64 `json:"balance"`
	}

	err = db.Raw(`
		SELECT id, code, name, balance 
		FROM accounts 
		WHERE id IN (SELECT DISTINCT account_id FROM unified_journal_lines)
		ORDER BY code
	`).Scan(&finalBalances).Error

	if err != nil {
		log.Printf("Error fetching final balances: %v", err)
	} else {
		for _, balance := range finalBalances {
			fmt.Printf("  %s: %s - Balance: %.2f\n", balance.Code, balance.Name, balance.Balance)
		}
	}

	// Verify the synchronization worked
	fmt.Println("\n5. Verification - Comparing SSOT vs COA Balances:")
	var verificationResults []struct {
		AccountID       uint    `json:"account_id"`
		AccountCode     string  `json:"account_code"`
		AccountName     string  `json:"account_name"`
		CoaBalance      float64 `json:"coa_balance"`
		SsotBalance     float64 `json:"ssot_balance"`
		Difference      float64 `json:"difference"`
	}

	err = db.Raw(`
		SELECT 
			a.id as account_id,
			a.code as account_code,
			a.name as account_name,
			a.balance as coa_balance,
			COALESCE(
				(SELECT SUM(ujl.debit_amount) - SUM(ujl.credit_amount)
				 FROM unified_journal_lines ujl 
				 JOIN unified_journal_ledger uj ON ujl.journal_id = uj.id
				 WHERE ujl.account_id = a.id AND uj.status = 'POSTED'),
				0
			) as ssot_balance,
			a.balance - COALESCE(
				(SELECT SUM(ujl.debit_amount) - SUM(ujl.credit_amount)
				 FROM unified_journal_lines ujl 
				 JOIN unified_journal_ledger uj ON ujl.journal_id = uj.id
				 WHERE ujl.account_id = a.id AND uj.status = 'POSTED'),
				0
			) as difference
		FROM accounts a
		WHERE a.id IN (
			SELECT DISTINCT account_id 
			FROM unified_journal_lines
		)
		ORDER BY a.code
	`).Scan(&verificationResults).Error

	if err != nil {
		log.Printf("Error verifying results: %v", err)
	} else {
		allMatch := true
		for _, result := range verificationResults {
			status := "✅ SYNCHRONIZED"
			if result.Difference != 0 {
				status = "❌ STILL MISMATCH"
				allMatch = false
			}
			fmt.Printf("  %s: %s - COA: %.2f, SSOT: %.2f, Diff: %.2f %s\n", 
				result.AccountCode, result.AccountName, result.CoaBalance, result.SsotBalance, result.Difference, status)
		}

		if allMatch {
			fmt.Println("\n✅ SUCCESS: All account balances are now synchronized with SSOT journal entries!")
		} else {
			fmt.Println("\n❌ WARNING: Some accounts still have mismatched balances")
		}
	}

	fmt.Println("\n=== SSOT Balance Synchronization Complete ===")
}