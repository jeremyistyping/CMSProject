package main

import (
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

func main() {
	log.Println("🔧 Starting cleanup of duplicate sales journals...")
	
	// Connect to database
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}
	
	log.Println("✅ Connected to database")
	
	// Find and cleanup duplicate sales journals
	if err := cleanupDuplicateSalesJournals(db); err != nil {
		log.Fatalf("❌ Cleanup failed: %v", err)
	}
	
	// Find and cleanup duplicate payment journals
	if err := cleanupDuplicatePaymentJournals(db); err != nil {
		log.Fatalf("❌ Cleanup failed: %v", err)
	}
	
	// Recalculate COA balances after cleanup
	if err := recalculateCOABalances(db); err != nil {
		log.Fatalf("❌ Recalculation failed: %v", err)
	}
	
	log.Println("✅ Cleanup completed successfully!")
}

func cleanupDuplicateSalesJournals(db *gorm.DB) error {
	log.Println("\n📋 Checking for duplicate SALES journals...")
	
	// Find sales with multiple journal entries
	type DuplicateInfo struct {
		TransactionID uint
		Count         int64
	}
	
	var duplicates []DuplicateInfo
	err := db.Raw(`
		SELECT transaction_id, COUNT(*) as count
		FROM simple_ssot_journals
		WHERE transaction_type = 'SALES'
		AND deleted_at IS NULL
		GROUP BY transaction_id
		HAVING COUNT(*) > 1
		ORDER BY count DESC
	`).Scan(&duplicates).Error
	
	if err != nil {
		return fmt.Errorf("failed to find duplicates: %v", err)
	}
	
	if len(duplicates) == 0 {
		log.Println("✅ No duplicate SALES journals found")
		return nil
	}
	
	log.Printf("⚠️ Found %d sales with duplicate journals", len(duplicates))
	
	// For each duplicate, keep only the oldest entry and delete the rest
	for _, dup := range duplicates {
		log.Printf("\n🔍 Processing Sale ID %d (%d duplicate entries)...", dup.TransactionID, dup.Count)
		
		// Get all journals for this sale, ordered by creation time
		var journals []models.SimpleSSOTJournal
		err := db.Where("transaction_type = ? AND transaction_id = ?", "SALES", dup.TransactionID).
			Order("created_at ASC").
			Find(&journals).Error
		
		if err != nil {
			log.Printf("❌ Error loading journals for sale %d: %v", dup.TransactionID, err)
			continue
		}
		
		if len(journals) <= 1 {
			continue // No duplicates
		}
		
		// Keep the first (oldest) journal
		keepJournal := journals[0]
		log.Printf("  ✅ Keeping journal ID %d (created: %s)", keepJournal.ID, keepJournal.CreatedAt.Format(time.RFC3339))
		
		// Delete all journal items for duplicates
		for i := 1; i < len(journals); i++ {
			journal := journals[i]
			log.Printf("  🗑️ Deleting duplicate journal ID %d (created: %s)", journal.ID, journal.CreatedAt.Format(time.RFC3339))
			
			// Delete journal items first
			if err := db.Where("journal_id = ?", journal.ID).Delete(&models.SimpleSSOTJournalItem{}).Error; err != nil {
				log.Printf("    ❌ Error deleting journal items: %v", err)
				continue
			}
			
			// Delete journal entry
			if err := db.Delete(&journal).Error; err != nil {
				log.Printf("    ❌ Error deleting journal: %v", err)
				continue
			}
			
			log.Printf("    ✅ Deleted journal ID %d and its items", journal.ID)
		}
	}
	
	log.Printf("\n✅ Cleaned up %d duplicate SALES journals", len(duplicates))
	return nil
}

func cleanupDuplicatePaymentJournals(db *gorm.DB) error {
	log.Println("\n📋 Checking for duplicate SALES_PAYMENT journals...")
	
	type DuplicateInfo struct {
		TransactionID uint
		Count         int64
	}
	
	var duplicates []DuplicateInfo
	err := db.Raw(`
		SELECT transaction_id, COUNT(*) as count
		FROM simple_ssot_journals
		WHERE transaction_type = 'SALES_PAYMENT'
		AND deleted_at IS NULL
		GROUP BY transaction_id
		HAVING COUNT(*) > 1
		ORDER BY count DESC
	`).Scan(&duplicates).Error
	
	if err != nil {
		return fmt.Errorf("failed to find duplicates: %v", err)
	}
	
	if len(duplicates) == 0 {
		log.Println("✅ No duplicate SALES_PAYMENT journals found")
		return nil
	}
	
	log.Printf("⚠️ Found %d payments with duplicate journals", len(duplicates))
	
	// For each duplicate, keep only the oldest entry and delete the rest
	for _, dup := range duplicates {
		log.Printf("\n🔍 Processing Payment ID %d (%d duplicate entries)...", dup.TransactionID, dup.Count)
		
		var journals []models.SimpleSSOTJournal
		err := db.Where("transaction_type = ? AND transaction_id = ?", "SALES_PAYMENT", dup.TransactionID).
			Order("created_at ASC").
			Find(&journals).Error
		
		if err != nil {
			log.Printf("❌ Error loading journals for payment %d: %v", dup.TransactionID, err)
			continue
		}
		
		if len(journals) <= 1 {
			continue
		}
		
		// Keep the first (oldest) journal
		keepJournal := journals[0]
		log.Printf("  ✅ Keeping journal ID %d (created: %s)", keepJournal.ID, keepJournal.CreatedAt.Format(time.RFC3339))
		
		// Delete duplicates
		for i := 1; i < len(journals); i++ {
			journal := journals[i]
			log.Printf("  🗑️ Deleting duplicate journal ID %d (created: %s)", journal.ID, journal.CreatedAt.Format(time.RFC3339))
			
			// Delete journal items first
			if err := db.Where("journal_id = ?", journal.ID).Delete(&models.SimpleSSOTJournalItem{}).Error; err != nil {
				log.Printf("    ❌ Error deleting journal items: %v", err)
				continue
			}
			
			// Delete journal entry
			if err := db.Delete(&journal).Error; err != nil {
				log.Printf("    ❌ Error deleting journal: %v", err)
				continue
			}
			
			log.Printf("    ✅ Deleted journal ID %d and its items", journal.ID)
		}
	}
	
	log.Printf("\n✅ Cleaned up %d duplicate SALES_PAYMENT journals", len(duplicates))
	return nil
}

func recalculateCOABalances(db *gorm.DB) error {
	log.Println("\n📊 Recalculating COA balances from journal entries...")
	
	// Disable triggers temporarily to avoid conflicts
	log.Println("  ⏸️ Disabling triggers temporarily...")
	db.Exec("SET session_replication_role = replica;")
	defer db.Exec("SET session_replication_role = DEFAULT;")
	
	// Reset all account balances to 0
	log.Println("  🔄 Resetting all account balances to 0...")
	if err := db.Exec("UPDATE accounts SET balance = 0 WHERE deleted_at IS NULL").Error; err != nil {
		return fmt.Errorf("failed to reset balances: %v", err)
	}
	
	// Get all non-deleted journal items
	var journalItems []models.SimpleSSOTJournalItem
	err := db.Joins("JOIN simple_ssot_journals ON simple_ssot_journals.id = simple_ssot_journal_items.journal_id").
		Where("simple_ssot_journals.status = ? AND simple_ssot_journals.deleted_at IS NULL", "POSTED").
		Find(&journalItems).Error
	
	if err != nil {
		return fmt.Errorf("failed to load journal items: %v", err)
	}
	
	log.Printf("  📝 Processing %d journal items...", len(journalItems))
	
	// Calculate balance changes per account
	balanceChanges := make(map[uint]float64)
	
	for _, item := range journalItems {
		// Get account type
		var account models.Account
		if err := db.First(&account, item.AccountID).Error; err != nil {
			log.Printf("    ⚠️ Warning: Account %d not found for journal item", item.AccountID)
			continue
		}
		
		// Calculate net change based on account type
		var netChange float64
		switch account.Type {
		case "ASSET", "EXPENSE":
			// Assets and Expenses increase with debit
			netChange = item.Debit - item.Credit
		case "LIABILITY", "EQUITY", "REVENUE":
			// Liabilities, Equity, and Revenue increase with credit (stored as negative)
			netChange = item.Credit - item.Debit
		default:
			log.Printf("    ⚠️ Unknown account type: %s for account %d", account.Type, account.ID)
			continue
		}
		
		balanceChanges[item.AccountID] += netChange
	}
	
	// Update account balances
	log.Printf("  💾 Updating balances for %d accounts...", len(balanceChanges))
	for accountID, balance := range balanceChanges {
		if err := db.Model(&models.Account{}).Where("id = ?", accountID).Update("balance", balance).Error; err != nil {
			log.Printf("    ❌ Error updating balance for account %d: %v", accountID, err)
		}
	}
	
	// Update parent account balances recursively
	log.Println("  🔗 Updating parent account balances...")
	if err := updateAllParentBalances(db); err != nil {
		return fmt.Errorf("failed to update parent balances: %v", err)
	}
	
	log.Println("✅ COA balances recalculated successfully")
	return nil
}

func updateAllParentBalances(db *gorm.DB) error {
	// Get all parent accounts (accounts that have children)
	var parentAccounts []models.Account
	err := db.Raw(`
		SELECT DISTINCT a.*
		FROM accounts a
		WHERE EXISTS (
			SELECT 1 FROM accounts c 
			WHERE c.parent_id = a.id 
			AND c.deleted_at IS NULL
		)
		AND a.deleted_at IS NULL
		ORDER BY a.level DESC
	`).Scan(&parentAccounts).Error
	
	if err != nil {
		return fmt.Errorf("failed to get parent accounts: %v", err)
	}
	
	// Update each parent's balance as sum of children
	for _, parent := range parentAccounts {
		var childrenSum float64
		err := db.Model(&models.Account{}).
			Where("parent_id = ? AND deleted_at IS NULL", parent.ID).
			Select("COALESCE(SUM(balance), 0)").
			Scan(&childrenSum).Error
		
		if err != nil {
			log.Printf("    ⚠️ Error calculating children sum for account %d: %v", parent.ID, err)
			continue
		}
		
		// Update parent balance
		if err := db.Model(&models.Account{}).Where("id = ?", parent.ID).Update("balance", childrenSum).Error; err != nil {
			log.Printf("    ⚠️ Error updating parent balance for account %d: %v", parent.ID, err)
			continue
		}
		
		log.Printf("    ✅ Updated parent account %s (%s) balance to %.2f", parent.Code, parent.Name, childrenSum)
	}
	
	return nil
}

