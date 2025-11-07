package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
)

type CashBankWithCOA struct {
	ID          uint
	Code        string
	Name        string
	Balance     float64
	AccountID   uint
	AccountCode string
	AccountName string
}

type AccountUsage struct {
	AccountID   uint
	AccountCode string
	AccountName string
	UsageCount  int
	CashBanks   string // Comma-separated list of cash bank names
}

func main() {
	// Load configuration
	_ = config.LoadConfig()

	// Connect to database
	db := database.ConnectDB()

	fmt.Println("🔧 COMPREHENSIVE FIX: SHARED COA & RECALCULATE BALANCES")
	fmt.Println("=" + string(make([]byte, 70)))

	// ==========================================
	// STEP 1: Find Shared COA Accounts
	// ==========================================
	fmt.Println("\n📊 STEP 1: Checking for shared COA accounts...")
	
	var sharedAccounts []AccountUsage
	query := `
		SELECT 
			a.id as account_id,
			a.code as account_code,
			a.name as account_name,
			COUNT(cb.id) as usage_count,
			STRING_AGG(cb.name, ', ') as cash_banks
		FROM accounts a
		JOIN cash_banks cb ON cb.account_id = a.id AND cb.deleted_at IS NULL
		WHERE a.deleted_at IS NULL
		GROUP BY a.id, a.code, a.name
		HAVING COUNT(cb.id) > 1
		ORDER BY usage_count DESC
	`
	
	if err := db.Raw(query).Scan(&sharedAccounts).Error; err != nil {
		log.Fatalf("❌ Failed to check shared accounts: %v", err)
	}

	if len(sharedAccounts) > 0 {
		fmt.Printf("⚠️  Found %d COA accounts shared by multiple Cash Banks:\n", len(sharedAccounts))
		for _, sa := range sharedAccounts {
			fmt.Printf("\n  Account [%s] %s - Used by %d Cash Banks:\n", sa.AccountCode, sa.AccountName, sa.UsageCount)
			fmt.Printf("    → %s\n", sa.CashBanks)
		}
		
		fmt.Println("\n⚠️  WARNING: Multiple cash banks sharing same COA account will cause balance conflicts!")
		fmt.Println("   Each cash bank should have its own dedicated COA account.")
	} else {
		fmt.Println("✅ No shared COA accounts found")
	}

	// ==========================================
	// STEP 2: Create dedicated COA accounts for shared cash banks
	// ==========================================
	fmt.Println("\n\n📊 STEP 2: Creating dedicated COA accounts...")
	
	if len(sharedAccounts) > 0 {
		fmt.Print("\n❓ Do you want to create dedicated COA accounts for each cash bank? (y/n): ")
		var createAccounts string
		fmt.Scanln(&createAccounts)
		
		if createAccounts == "y" || createAccounts == "Y" {
			tx := db.Begin()
			
			for _, sa := range sharedAccounts {
				// Get all cash banks using this account
				var cashBanks []CashBankWithCOA
				db.Raw(`
					SELECT id, code, name, balance, account_id
					FROM cash_banks
					WHERE account_id = ? AND deleted_at IS NULL
					ORDER BY id
				`, sa.AccountID).Scan(&cashBanks)
				
				// Keep first cash bank with original account, create new for others
				for i, cb := range cashBanks {
					if i == 0 {
						fmt.Printf("  ✅ [%s] %s - Keeping original account %s\n", cb.Code, cb.Name, sa.AccountCode)
						continue
					}
					
					// Find next available account code
					baseCode := sa.AccountCode
					var newCode string
					counter := 1
					for {
						newCode = fmt.Sprintf("%s-%03d", baseCode, counter)
						var exists int64
						tx.Raw("SELECT COUNT(*) FROM accounts WHERE code = ?", newCode).Scan(&exists)
						if exists == 0 {
							break
						}
						counter++
					}
					
					// Create new account
					newAccountName := fmt.Sprintf("%s - %s", sa.AccountName, cb.Name)
					fmt.Printf("  🔧 [%s] %s - Creating new account %s (%s)\n", cb.Code, cb.Name, newCode, newAccountName)
					
					var parentID *uint
					tx.Raw("SELECT parent_id FROM accounts WHERE id = ?", sa.AccountID).Scan(&parentID)
					
					var newAccountID uint
					err := tx.Exec(`
						INSERT INTO accounts (code, name, type, parent_id, balance, created_at, updated_at)
						SELECT ?, ?, type, parent_id, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
						FROM accounts WHERE id = ?
						RETURNING id
					`, newCode, newAccountName, sa.AccountID).Error
					
					if err != nil {
						tx.Rollback()
						log.Fatalf("❌ Failed to create account %s: %v", newCode, err)
					}
					
					// Get the new account ID
					tx.Raw("SELECT id FROM accounts WHERE code = ?", newCode).Scan(&newAccountID)
					
					// Update cash bank to use new account
					if err := tx.Exec(`
						UPDATE cash_banks 
						SET account_id = ?, updated_at = CURRENT_TIMESTAMP 
						WHERE id = ?
					`, newAccountID, cb.ID).Error; err != nil {
						tx.Rollback()
						log.Fatalf("❌ Failed to update cash bank %s: %v", cb.Code, err)
					}
					
					fmt.Printf("    → Linked to new account %s (ID: %d)\n", newCode, newAccountID)
				}
			}
			
			if err := tx.Commit().Error; err != nil {
				log.Fatalf("❌ Failed to commit: %v", err)
			}
			
			fmt.Println("\n✅ Dedicated COA accounts created successfully!")
		} else {
			fmt.Println("⏭️  Skipping COA account creation")
		}
	}

	// ==========================================
	// STEP 3: Recalculate Cash Bank Balances
	// ==========================================
	fmt.Println("\n\n📊 STEP 3: Recalculating Cash Bank balances from transactions...")
	
	var cashBanks []CashBankWithCOA
	db.Raw(`
		SELECT 
			cb.id, cb.code, cb.name, cb.balance, cb.account_id,
			a.code as account_code, a.name as account_name
		FROM cash_banks cb
		LEFT JOIN accounts a ON cb.account_id = a.id
		WHERE cb.deleted_at IS NULL
		ORDER BY cb.id
	`).Scan(&cashBanks)
	
	fmt.Print("\n❓ Recalculate Cash Bank balances from transactions? (y/n): ")
	var recalc string
	fmt.Scanln(&recalc)
	
	if recalc == "y" || recalc == "Y" {
		tx := db.Begin()
		
		for _, cb := range cashBanks {
			// Calculate correct balance from transactions
			var txSum float64
			tx.Raw(`
				SELECT COALESCE(SUM(amount), 0)
				FROM cash_bank_transactions
				WHERE cash_bank_id = ? AND deleted_at IS NULL
			`, cb.ID).Scan(&txSum)
			
			oldBalance := cb.Balance
			
			if oldBalance != txSum {
				fmt.Printf("  🔧 [%s] %s: %.2f → %.2f (diff: %.2f)\n", 
					cb.Code, cb.Name, oldBalance, txSum, txSum-oldBalance)
				
				// Update cash bank balance
				if err := tx.Exec(`
					UPDATE cash_banks 
					SET balance = ?, updated_at = CURRENT_TIMESTAMP 
					WHERE id = ?
				`, txSum, cb.ID).Error; err != nil {
					tx.Rollback()
					log.Fatalf("❌ Failed to update cash bank %s: %v", cb.Code, err)
				}
			} else {
				fmt.Printf("  ✅ [%s] %s: Balance correct (%.2f)\n", cb.Code, cb.Name, txSum)
			}
		}
		
		if err := tx.Commit().Error; err != nil {
			log.Fatalf("❌ Failed to commit: %v", err)
		}
		
		fmt.Println("\n✅ Cash Bank balances recalculated!")
	} else {
		fmt.Println("⏭️  Skipping Cash Bank balance recalculation")
	}

	// ==========================================
	// STEP 4: Sync COA Balances with Cash Banks
	// ==========================================
	fmt.Println("\n\n📊 STEP 4: Syncing COA balances with Cash Bank balances...")
	
	// Reload cash banks after potential updates
	db.Raw(`
		SELECT 
			cb.id, cb.code, cb.name, cb.balance, cb.account_id,
			a.code as account_code, a.name as account_name
		FROM cash_banks cb
		LEFT JOIN accounts a ON cb.account_id = a.id
		WHERE cb.deleted_at IS NULL
		ORDER BY cb.id
	`).Scan(&cashBanks)
	
	fmt.Print("\n❓ Sync COA account balances with Cash Bank balances? (y/n): ")
	var syncCOA string
	fmt.Scanln(&syncCOA)
	
	if syncCOA == "y" || syncCOA == "Y" {
		tx := db.Begin()
		
		for _, cb := range cashBanks {
			if cb.AccountID == 0 {
				fmt.Printf("  ⚠️  [%s] %s: No linked COA account, skipping\n", cb.Code, cb.Name)
				continue
			}
			
			// Get current COA balance
			var coaBalance float64
			tx.Raw("SELECT balance FROM accounts WHERE id = ?", cb.AccountID).Scan(&coaBalance)
			
			if coaBalance != cb.Balance {
				fmt.Printf("  🔧 [%s] %s → Account %s: %.2f → %.2f\n", 
					cb.Code, cb.Name, cb.AccountCode, coaBalance, cb.Balance)
				
				// Update COA balance
				if err := tx.Exec(`
					UPDATE accounts 
					SET balance = ?, updated_at = CURRENT_TIMESTAMP 
					WHERE id = ?
				`, cb.Balance, cb.AccountID).Error; err != nil {
					tx.Rollback()
					log.Fatalf("❌ Failed to update account %s: %v", cb.AccountCode, err)
				}
			} else {
				fmt.Printf("  ✅ [%s] %s → Account %s: Already synced (%.2f)\n", 
					cb.Code, cb.Name, cb.AccountCode, cb.Balance)
			}
		}
		
		if err := tx.Commit().Error; err != nil {
			log.Fatalf("❌ Failed to commit: %v", err)
		}
		
		fmt.Println("\n✅ COA balances synced!")
	} else {
		fmt.Println("⏭️  Skipping COA balance sync")
	}

	// ==========================================
	// STEP 5: Final Verification
	// ==========================================
	fmt.Println("\n\n📊 STEP 5: Final Verification...")
	fmt.Println("-" + string(make([]byte, 70)))
	
	var finalCheck []struct {
		Code           string
		Name           string
		CashBankBalance float64
		AccountCode    string
		COABalance     float64
		Difference     float64
	}
	
	db.Raw(`
		SELECT 
			cb.code,
			cb.name,
			cb.balance as cash_bank_balance,
			a.code as account_code,
			a.balance as coa_balance,
			(cb.balance - a.balance) as difference
		FROM cash_banks cb
		LEFT JOIN accounts a ON cb.account_id = a.id
		WHERE cb.deleted_at IS NULL
		ORDER BY cb.id
	`).Scan(&finalCheck)
	
	allCorrect := true
	for _, fc := range finalCheck {
		status := "✅"
		if fc.Difference != 0 {
			status = "❌"
			allCorrect = false
		}
		
		fmt.Printf("%s [%s] %s:\n", status, fc.Code, fc.Name)
		fmt.Printf("   Cash/Bank: %.2f | COA [%s]: %.2f | Diff: %.2f\n\n", 
			fc.CashBankBalance, fc.AccountCode, fc.COABalance, fc.Difference)
	}
	
	if allCorrect {
		fmt.Println("🎉 ALL BALANCES ARE NOW IN SYNC!")
	} else {
		fmt.Println("⚠️  Some balances still not synced. Manual investigation may be needed.")
	}
	
	fmt.Println("\n" + string(make([]byte, 70)))
	fmt.Println("✅ Fix process completed!")
}
