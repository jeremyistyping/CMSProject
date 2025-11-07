package main

import (
	"log"

	"app-sistem-akuntansi/database"
)

func main() {
	log.Println("🔧 Disabling auto-sync triggers and fixing balances...")
	
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}
	
	// Step 1: Disable triggers
	log.Println("\n📍 Step 1: Disabling auto-sync triggers...")
	
	triggers := []struct{
		name  string
		table string
	}{
		{"trigger_sync_cashbank_coa", "cash_banks"},
		{"trigger_recalc_cashbank_balance_insert", "cash_bank_transactions"},
		{"trigger_recalc_cashbank_balance_update", "cash_bank_transactions"},
		{"trigger_recalc_cashbank_balance_delete", "cash_bank_transactions"},
		{"trigger_validate_account_balance", "accounts"}, // This also causes conflicts
	}
	
	for _, trigger := range triggers {
		sql := "DROP TRIGGER IF EXISTS " + trigger.name + " ON " + trigger.table
		if err := db.Exec(sql).Error; err != nil {
			log.Printf("⚠️ Warning: Failed to drop trigger %s: %v", trigger.name, err)
		} else {
			log.Printf("✅ Dropped trigger: %s on %s", trigger.name, trigger.table)
		}
	}
	
	log.Println("\n✅ All auto-sync triggers disabled!")
	log.Println("   These triggers were causing DOUBLE POSTING:")
	log.Println("   - Journal system updates COA")
	log.Println("   - Trigger updates COA AGAIN")
	log.Println("   - Result: COA balance = 2x the correct amount!")
	
	// Step 2: Reset ALL account balances to 0
	log.Println("\n📍 Step 2: Resetting all account balances...")
	
	// Disable ALL triggers temporarily to avoid conflicts
	db.Exec("SET session_replication_role = replica;")
	defer db.Exec("SET session_replication_role = DEFAULT;")
	
	if err := db.Exec("UPDATE accounts SET balance = 0 WHERE deleted_at IS NULL").Error; err != nil {
		log.Fatalf("❌ Failed to reset balances: %v", err)
	}
	log.Println("✅ All account balances reset to 0")
	
	// Step 3: Recalculate from journal entries
	log.Println("\n📍 Step 3: Recalculating balances from journal entries...")
	
	type JournalItem struct {
		AccountID uint
		Debit     float64
		Credit    float64
	}
	
	var items []JournalItem
	err := db.Raw(`
		SELECT 
			ji.account_id,
			ji.debit,
			ji.credit
		FROM simple_ssot_journal_items ji
		JOIN simple_ssot_journals j ON j.id = ji.journal_id
		WHERE j.status = 'POSTED'
		AND j.deleted_at IS NULL
	`).Scan(&items).Error
	
	if err != nil {
		log.Fatalf("❌ Failed to load journal items: %v", err)
	}
	
	log.Printf("📝 Processing %d journal items...", len(items))
	
	// Calculate balance changes per account
	balanceChanges := make(map[uint]float64)
	
	for _, item := range items {
		// Get account type
		type AccountInfo struct {
			ID   uint
			Type string
		}
		var acc AccountInfo
		if err := db.Raw("SELECT id, type FROM accounts WHERE id = ? AND deleted_at IS NULL", item.AccountID).Scan(&acc).Error; err != nil {
			log.Printf("⚠️ Warning: Account %d not found", item.AccountID)
			continue
		}
		
		// Calculate net change based on account type
		var netChange float64
		switch acc.Type {
		case "ASSET", "EXPENSE":
			// Assets and Expenses increase with debit
			netChange = item.Debit - item.Credit
		case "LIABILITY", "EQUITY", "REVENUE":
			// Liabilities, Equity, and Revenue increase with credit
			netChange = item.Credit - item.Debit
		default:
			log.Printf("⚠️ Unknown account type: %s for account %d", acc.Type, acc.ID)
			continue
		}
		
		balanceChanges[item.AccountID] += netChange
	}
	
	// Update account balances
	log.Printf("💾 Updating balances for %d accounts...", len(balanceChanges))
	for accountID, balance := range balanceChanges {
		if err := db.Exec("UPDATE accounts SET balance = ?, updated_at = NOW() WHERE id = ?", balance, accountID).Error; err != nil {
			log.Printf("❌ Error updating balance for account %d: %v", accountID, err)
		}
	}
	
	// Step 4: Update parent account balances
	log.Println("\n📍 Step 4: Updating parent account balances...")
	
	type ParentAccount struct {
		ID   uint
		Code string
		Name string
	}
	
	var parents []ParentAccount
	db.Raw(`
		SELECT DISTINCT a.id, a.code, a.name
		FROM accounts a
		WHERE EXISTS (
			SELECT 1 FROM accounts c 
			WHERE c.parent_id = a.id 
			AND c.deleted_at IS NULL
		)
		AND a.deleted_at IS NULL
		ORDER BY a.level DESC
	`).Scan(&parents)
	
	for _, parent := range parents {
		var childrenSum float64
		db.Raw(`
			SELECT COALESCE(SUM(balance), 0)
			FROM accounts 
			WHERE parent_id = ? 
			AND deleted_at IS NULL
		`, parent.ID).Scan(&childrenSum)
		
		if err := db.Exec("UPDATE accounts SET balance = ?, updated_at = NOW() WHERE id = ?", childrenSum, parent.ID).Error; err != nil {
			log.Printf("❌ Error updating parent %d: %v", parent.ID, err)
		} else {
			log.Printf("✅ Updated parent %s (%s) = %.2f", parent.Code, parent.Name, childrenSum)
		}
	}
	
	// Step 5: Show final balances
	log.Println("\n📊 Final Account Balances:")
	
	type Account struct {
		Code    string
		Name    string
		Balance float64
	}
	
	var accounts []Account
	db.Raw(`
		SELECT code, name, balance
		FROM accounts
		WHERE code IN ('1102', '1201', '4101', '2103')
		AND deleted_at IS NULL
		ORDER BY code
	`).Scan(&accounts)
	
	log.Println("┌──────┬──────────────────────┬──────────────┐")
	log.Println("│ Code │ Name                 │ Balance      │")
	log.Println("├──────┼──────────────────────┼──────────────┤")
	for _, acc := range accounts {
		log.Printf("│ %-4s │ %-20s │ %12.2f │", acc.Code, acc.Name, acc.Balance)
	}
	log.Println("└──────┴──────────────────────┴──────────────┘")
	
	log.Println("\n✅ ALL DONE!")
	log.Println("   - Triggers disabled")
	log.Println("   - Balances recalculated from journals")
	log.Println("   - Parent balances updated")
	log.Println("\n🚀 Restart backend to ensure changes take effect!")
}

