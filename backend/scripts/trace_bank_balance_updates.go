package main

import (
	"log"

	"app-sistem-akuntansi/database"
)

func main() {
	log.Println("🔍 Tracing all Bank (1102) balance updates...")
	
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}
	
	// Get Bank account info
	type AccountInfo struct {
		ID      uint
		Code    string
		Name    string
		Balance float64
	}
	
	var bankAccount AccountInfo
	if err := db.Raw("SELECT id, code, name, balance FROM accounts WHERE code = '1102' AND deleted_at IS NULL").Scan(&bankAccount).Error; err != nil {
		log.Fatalf("Error loading bank account: %v", err)
	}
	
	log.Printf("\n💰 Bank Account (1102):")
	log.Printf("  ID: %d", bankAccount.ID)
	log.Printf("  Name: %s", bankAccount.Name)
	log.Printf("  Current Balance: %.2f", bankAccount.Balance)
	
	// Check all journal items affecting Bank (1102)
	type JournalItem struct {
		ItemID      uint
		JournalID   uint
		JournalType string
		TransID     uint
		EntryNumber string
		Debit       float64
		Credit      float64
		NetChange   float64
		CreatedAt   string
	}
	
	var items []JournalItem
	err := db.Raw(`
		SELECT 
			ji.id as item_id,
			j.id as journal_id,
			j.transaction_type as journal_type,
			j.transaction_id as trans_id,
			j.entry_number,
			ji.debit,
			ji.credit,
			(ji.debit - ji.credit) as net_change,
			j.created_at
		FROM simple_ssot_journal_items ji
		JOIN simple_ssot_journals j ON j.id = ji.journal_id
		WHERE ji.account_id = ?
		AND j.deleted_at IS NULL
		ORDER BY j.created_at ASC
	`, bankAccount.ID).Scan(&items).Error
	
	if err != nil {
		log.Fatalf("Error loading journal items: %v", err)
	}
	
	log.Printf("\n📊 All Journal Items Affecting Bank (1102):")
	log.Println("┌─────────┬────────────┬──────────────┬──────────────┬──────────────┬──────────────┬─────────────────────┐")
	log.Println("│ Item ID │ Journal ID │ Type         │ Trans ID     │ Debit        │ Credit       │ Net Change          │")
	log.Println("├─────────┼────────────┼──────────────┼──────────────┼──────────────┼──────────────┼─────────────────────┤")
	
	runningBalance := 0.0
	for _, item := range items {
		runningBalance += item.NetChange
		log.Printf("│ %-7d │ %-10d │ %-12s │ %-12d │ %12.2f │ %12.2f │ %12.2f → %7.2f │", 
			item.ItemID, item.JournalID, item.JournalType, item.TransID,
			item.Debit, item.Credit, item.NetChange, runningBalance)
	}
	log.Println("└─────────┴────────────┴──────────────┴──────────────┴──────────────┴──────────────┴─────────────────────┘")
	
	log.Printf("\n🧮 Calculation from Journal Items:")
	log.Printf("  Expected Balance (from journals): %.2f", runningBalance)
	log.Printf("  Actual Balance (in accounts table): %.2f", bankAccount.Balance)
	log.Printf("  Difference: %.2f", bankAccount.Balance - runningBalance)
	
	if bankAccount.Balance != runningBalance {
		log.Printf("\n❌ MISMATCH! Account balance doesn't match journal calculation!")
		log.Printf("   This means COA balance was updated OUTSIDE of journal system!")
		log.Printf("\n🔍 Possible causes:")
		log.Printf("   1. Manual COA balance update (via updateCOABalance)")
		log.Printf("   2. Cash/Bank transaction that updated COA directly")
		log.Printf("   3. Trigger or stored procedure")
		log.Printf("   4. Double update bug in payment processing")
	} else {
		log.Printf("\n✅ Balance matches journal calculation!")
	}
	
	// Check if there's a linked cash_bank record
	type CashBankInfo struct {
		ID         uint
		Name       string
		AccountID  uint
		Balance    float64
	}
	
	var cashBank CashBankInfo
	err = db.Raw("SELECT id, name, account_id, balance FROM cash_banks WHERE account_id = ? AND deleted_at IS NULL", bankAccount.ID).Scan(&cashBank).Error
	
	if err == nil && cashBank.ID > 0 {
		log.Printf("\n💳 Linked Cash/Bank Record:")
		log.Printf("  Cash Bank ID: %d", cashBank.ID)
		log.Printf("  Name: %s", cashBank.Name)
		log.Printf("  Balance: %.2f", cashBank.Balance)
		
		// Check cash_bank_transactions
		type CashBankTx struct {
			ID              uint
			TransactionType string
			Amount          float64
			Description     string
			CreatedAt       string
		}
		
		var cbTxs []CashBankTx
		db.Raw(`
			SELECT id, transaction_type, amount, description, created_at
			FROM cash_bank_transactions
			WHERE cash_bank_id = ?
			AND deleted_at IS NULL
			ORDER BY created_at ASC
		`, cashBank.ID).Scan(&cbTxs)
		
		if len(cbTxs) > 0 {
			log.Printf("\n⚠️ Found %d Cash/Bank Transactions:", len(cbTxs))
			for _, tx := range cbTxs {
				log.Printf("  - ID %d: %s | Amount: %.2f | %s", 
					tx.ID, tx.TransactionType, tx.Amount, tx.Description)
			}
			log.Printf("\n❗ These transactions might have updated COA balance directly!")
		}
	}
}

