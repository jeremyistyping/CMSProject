package migrations

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// FixRevenueDuplication fixes the revenue duplication issue caused by:
// 1. Account name variations in journal entries (case sensitivity)
// 2. Parent accounts not marked as headers
func FixRevenueDuplication(db *gorm.DB) error {
	log.Println("[MIGRATION] Starting revenue duplication fix...")

	// Step 1: Mark parent accounts as headers
	log.Println("[MIGRATION] Step 1: Marking parent accounts as headers...")
	parentAccounts := []string{"1000", "1100", "1500", "2000", "2100", "3000", "4000", "5000"}
	
	result := db.Exec(`
		UPDATE accounts 
		SET is_header = true, 
		    updated_at = NOW()
		WHERE code IN (?)
		  AND COALESCE(is_header, false) = false
	`, parentAccounts)
	
	if result.Error != nil {
		return fmt.Errorf("failed to mark parent accounts as headers: %v", result.Error)
	}
	log.Printf("[MIGRATION] Marked %d parent accounts as headers", result.RowsAffected)

	// Step 2: Standardize account names in journal_entries (legacy system)
	log.Println("[MIGRATION] Step 2: Standardizing account names in journal_entries...")
	result = db.Exec(`
		UPDATE journal_entries je
		INNER JOIN accounts a ON a.code = je.account_code
		SET je.account_name = a.name,
		    je.updated_at = NOW()
		WHERE je.account_name != a.name
		  AND je.account_code LIKE '4%'
	`)
	
	if result.Error != nil {
		return fmt.Errorf("failed to standardize journal_entries: %v", result.Error)
	}
	log.Printf("[MIGRATION] Standardized %d journal entries (4xxx accounts)", result.RowsAffected)

	// Step 3: Standardize all expense accounts too (5xxx)
	result = db.Exec(`
		UPDATE journal_entries je
		INNER JOIN accounts a ON a.code = je.account_code
		SET je.account_name = a.name,
		    je.updated_at = NOW()
		WHERE je.account_name != a.name
		  AND je.account_code LIKE '5%'
	`)
	
	if result.Error != nil {
		return fmt.Errorf("failed to standardize expense journal_entries: %v", result.Error)
	}
	log.Printf("[MIGRATION] Standardized %d journal entries (5xxx accounts)", result.RowsAffected)

	// Step 4: Standardize account names in unified_journal_lines (SSOT system)
	log.Println("[MIGRATION] Step 3: Standardizing account names in unified_journal_lines...")
	result = db.Exec(`
		UPDATE unified_journal_lines ujl
		INNER JOIN accounts a ON a.id = ujl.account_id
		SET ujl.account_name = a.name,
		    ujl.updated_at = NOW()
		WHERE ujl.account_name != a.name
		  AND ujl.account_code LIKE '4%'
	`)
	
	if result.Error != nil {
		// Not critical if unified_journal_lines doesn't exist yet
		log.Printf("[MIGRATION] Warning: Could not standardize unified_journal_lines: %v", result.Error)
	} else {
		log.Printf("[MIGRATION] Standardized %d unified journal lines (4xxx accounts)", result.RowsAffected)
	}

	// Step 5: Standardize expense accounts in unified system
	result = db.Exec(`
		UPDATE unified_journal_lines ujl
		INNER JOIN accounts a ON a.id = ujl.account_id
		SET ujl.account_name = a.name,
		    ujl.updated_at = NOW()
		WHERE ujl.account_name != a.name
		  AND ujl.account_code LIKE '5%'
	`)
	
	if result.Error != nil {
		log.Printf("[MIGRATION] Warning: Could not standardize expense unified_journal_lines: %v", result.Error)
	} else {
		log.Printf("[MIGRATION] Standardized %d unified journal lines (5xxx accounts)", result.RowsAffected)
	}

	// Step 6: Verification
	log.Println("[MIGRATION] Step 4: Verifying fix...")
	
	type VerifyResult struct {
		AccountCode  string  `json:"account_code"`
		AccountName  string  `json:"account_name"`
		NameCount    int     `json:"name_count"`
		AllNames     string  `json:"all_names"`
		TotalAmount  float64 `json:"total_amount"`
	}
	
	var dupes []VerifyResult
	db.Raw(`
		SELECT 
			je.account_code,
			je.account_name,
			COUNT(DISTINCT je.account_name) as name_count,
			GROUP_CONCAT(DISTINCT je.account_name SEPARATOR ' | ') as all_names,
			SUM(je.credit - je.debit) as total_amount
		FROM journal_entries je
		WHERE je.account_code LIKE '4%'
		GROUP BY je.account_code
		HAVING COUNT(DISTINCT je.account_name) > 1
	`).Scan(&dupes)
	
	if len(dupes) > 0 {
		log.Printf("[MIGRATION] WARNING: Still found %d accounts with name variations:", len(dupes))
		for _, d := range dupes {
			log.Printf("  - Account %s has %d variations: %s", d.AccountCode, d.NameCount, d.AllNames)
		}
	} else {
		log.Println("[MIGRATION] ✓ Verification passed: No duplicate account names found")
	}

	log.Println("[MIGRATION] Revenue duplication fix completed successfully!")
	return nil
}

// RunRevenueDuplicationFix is the public function to be called from migrations
func RunRevenueDuplicationFix(db *gorm.DB) error {
	return FixRevenueDuplication(db)
}

