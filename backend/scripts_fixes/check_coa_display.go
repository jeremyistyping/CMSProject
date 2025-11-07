package main

import (
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

func main() {
	log.Printf("🔍 Checking if Revenue and PPN Keluaran will display in COA...")

	// Initialize database connection
	db := database.ConnectDB()

	log.Printf("\n📊 Current COA Status:")
	checkCOADisplay(db)

	log.Printf("\n🎯 Specific Revenue and PPN Accounts:")
	checkSpecificAccounts(db)

	log.Printf("\n📋 Recent SSOT Journal Activity:")
	checkJournalActivity(db)

	log.Printf("\n✅ COA Display verification completed!")
}

func checkCOADisplay(db *gorm.DB) {
	// Get all accounts that would appear in COA (active accounts with balance != 0)
	var activeAccounts []models.Account
	db.Where("is_active = ? AND balance != 0", true).
		Order("code ASC").
		Find(&activeAccounts)

	log.Printf("📈 Active accounts with non-zero balances (will display in COA):")
	
	revenueFound := false
	ppnFound := false
	
	for _, acc := range activeAccounts {
		log.Printf("   %s - %s: Balance %.2f (%s)", 
			acc.Code, acc.Name, acc.Balance, acc.Type)
			
		// Check if our target accounts are included
		if acc.Code == "4101" || acc.Name == "Pendapatan Penjualan" {
			revenueFound = true
		}
		if acc.Code == "2103" || acc.Name == "PPN Keluaran" {
			ppnFound = true
		}
	}
	
	log.Printf("\n🎯 Target Account Status:")
	if revenueFound {
		log.Printf("   ✅ Revenue (4101) AKAN TAMPIL di COA")
	} else {
		log.Printf("   ❌ Revenue (4101) TIDAK AKAN TAMPIL di COA")
	}
	
	if ppnFound {
		log.Printf("   ✅ PPN Keluaran (2103) AKAN TAMPIL di COA")
	} else {
		log.Printf("   ❌ PPN Keluaran (2103) TIDAK AKAN TAMPIL di COA")
	}
}

func checkSpecificAccounts(db *gorm.DB) {
	targetCodes := []string{"4101", "2103"}
	
	for _, code := range targetCodes {
		var account models.Account
		err := db.Where("code = ?", code).First(&account).Error
		
		if err != nil {
			log.Printf("❌ Account %s: NOT FOUND", code)
			continue
		}
		
		log.Printf("📋 Account %s (%s):", code, account.Name)
		log.Printf("   - Balance: %.2f", account.Balance)
		log.Printf("   - Type: %s", account.Type)
		log.Printf("   - Active: %v", account.IsActive)
		log.Printf("   - Will display in COA: %v", account.IsActive && account.Balance != 0)
		
		if account.Balance != 0 {
			log.Printf("   ✅ HAS BALANCE - will appear in COA")
		} else {
			log.Printf("   ⚠️  ZERO BALANCE - will NOT appear in COA (normal for unused accounts)")
		}
	}
}

func checkJournalActivity(db *gorm.DB) {
	// Check recent SSOT journal entries affecting revenue/PPN accounts
	var entries []struct {
		EntryID     uint64 `gorm:"column:id"`
		Description string
		AccountCode string `gorm:"column:code"`
		AccountName string `gorm:"column:name"`
		DebitAmount float64
		CreditAmount float64
	}
	
	query := `
		SELECT 
			je.id,
			je.description,
			a.code,
			a.name,
			jl.debit_amount::float as debit_amount,
			jl.credit_amount::float as credit_amount
		FROM ssot_journal_entries je
		JOIN ssot_journal_lines jl ON je.id = jl.journal_entry_id  
		JOIN accounts a ON jl.account_id = a.id
		WHERE a.code IN ('4101', '2103')
		AND je.created_at > NOW() - INTERVAL '1 hour'
		ORDER BY je.created_at DESC
		LIMIT 10
	`
	
	db.Raw(query).Scan(&entries)
	
	if len(entries) == 0 {
		log.Printf("ℹ️  No recent journal activity for Revenue/PPN accounts")
		log.Printf("   This is normal - accounts only get activity when sales are INVOICED")
	} else {
		log.Printf("📊 Recent journal activity (%d entries):", len(entries))
		for _, entry := range entries {
			log.Printf("   Entry %d: %s", entry.EntryID, entry.Description)
			log.Printf("     → Account %s (%s): Debit %.2f, Credit %.2f", 
				entry.AccountCode, entry.AccountName, 
				entry.DebitAmount, entry.CreditAmount)
		}
	}
}