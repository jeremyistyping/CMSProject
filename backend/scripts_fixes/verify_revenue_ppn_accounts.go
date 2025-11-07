package main

import (
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

func main() {
	log.Printf("🔍 Verifying Revenue and PPN KELUARAN accounts in database...")

	// Initialize database connection
	db := database.ConnectDB()

	// Check for Revenue account (4101)
	log.Printf("\n📊 Checking Revenue accounts...")
	var revenueAccounts []models.Account
	err := db.Where("code LIKE ? OR name ILIKE ?", "%4101%", "%revenue%").
		Where("type = ?", "REVENUE").
		Where("is_active = ?", true).
		Find(&revenueAccounts).Error
	
	if err != nil {
		log.Printf("❌ Error querying revenue accounts: %v", err)
	} else if len(revenueAccounts) == 0 {
		log.Printf("⚠️  No revenue accounts found!")
		// Create default revenue account
		createDefaultRevenueAccount(db)
	} else {
		log.Printf("✅ Found %d revenue account(s):", len(revenueAccounts))
		for _, acc := range revenueAccounts {
			log.Printf("   - Code: %s, Name: %s, Balance: %.2f, Active: %v", 
				acc.Code, acc.Name, acc.Balance, acc.IsActive)
		}
	}

	// Check for PPN KELUARAN account (2103)  
	log.Printf("\n🧾 Checking PPN KELUARAN accounts...")
	var ppnAccounts []models.Account
	err = db.Where("code LIKE ? OR name ILIKE ? OR name ILIKE ?", 
		"%2103%", "%PPN KELUARAN%", "%OUTPUT VAT%").
		Where("type = ?", "LIABILITY").
		Where("is_active = ?", true).
		Find(&ppnAccounts).Error
	
	if err != nil {
		log.Printf("❌ Error querying PPN accounts: %v", err)
	} else if len(ppnAccounts) == 0 {
		log.Printf("⚠️  No PPN KELUARAN accounts found!")
		// Create default PPN KELUARAN account
		createDefaultPPNKeluaranAccount(db)
	} else {
		log.Printf("✅ Found %d PPN KELUARAN account(s):", len(ppnAccounts))
		for _, acc := range ppnAccounts {
			log.Printf("   - Code: %s, Name: %s, Balance: %.2f, Active: %v", 
				acc.Code, acc.Name, acc.Balance, acc.IsActive)
		}
	}

	// Check for sales journal entries that should have updated these accounts
	log.Printf("\n📋 Checking SSOT journal entries affecting these accounts...")
	checkJournalEntriesForAccounts(db, revenueAccounts, ppnAccounts)

	log.Printf("\n✅ Account verification completed!")
}

func createDefaultRevenueAccount(db *gorm.DB) {
	account := &models.Account{
		Code:        "4101",
		Name:        "Pendapatan Penjualan",
		Type:        "REVENUE",
		Category:    "OPERATING_REVENUE",
		IsActive:    true,
		Balance:     0,
		Description: "Revenue from sales of goods and services",
	}

	err := db.Create(account).Error
	if err != nil {
		log.Printf("❌ Failed to create default revenue account: %v", err)
	} else {
		log.Printf("✅ Created default revenue account: %s - %s", account.Code, account.Name)
	}
}

func createDefaultPPNKeluaranAccount(db *gorm.DB) {
	account := &models.Account{
		Code:        "2103",
		Name:        "PPN Keluaran",
		Type:        "LIABILITY", 
		Category:    "CURRENT_LIABILITY",
		IsActive:    true,
		Balance:     0,
		Description: "Pajak Pertambahan Nilai Keluaran (Output VAT)",
	}

	err := db.Create(account).Error
	if err != nil {
		log.Printf("❌ Failed to create default PPN KELUARAN account: %v", err)
	} else {
		log.Printf("✅ Created default PPN KELUARAN account: %s - %s", account.Code, account.Name)
	}
}

func checkJournalEntriesForAccounts(db *gorm.DB, revenueAccounts, ppnAccounts []models.Account) {
	// Collect account IDs
	var accountIDs []uint
	for _, acc := range revenueAccounts {
		accountIDs = append(accountIDs, acc.ID)
	}
	for _, acc := range ppnAccounts {
		accountIDs = append(accountIDs, acc.ID)
	}

	if len(accountIDs) == 0 {
		log.Printf("⚠️  No target accounts to check journal entries for")
		return
	}

	// Count journal lines affecting these accounts
	var entryCount int64
	err := db.Model(&models.SSOTJournalLine{}).
		Where("account_id IN ?", accountIDs).
		Count(&entryCount).Error
	
	if err != nil {
		log.Printf("❌ Error counting journal entries: %v", err)
	} else {
		log.Printf("📊 Found %d SSOT journal lines affecting revenue/PPN accounts", entryCount)
		
		if entryCount > 0 {
			// Show some recent entries
			var recentLines []models.SSOTJournalLine
			err = db.Preload("Account").
				Where("account_id IN ?", accountIDs).
				Order("created_at DESC").
				Limit(5).
				Find(&recentLines).Error
			
			if err == nil && len(recentLines) > 0 {
				log.Printf("📋 Recent journal lines:")
				for _, line := range recentLines {
					log.Printf("   - Account: %s (%s), Debit: %.2f, Credit: %.2f", 
						line.Account.Code, line.Account.Name, 
						line.DebitAmount.InexactFloat64(), line.CreditAmount.InexactFloat64())
				}
			}
		}
	}
}