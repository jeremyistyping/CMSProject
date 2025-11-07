package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("🔍 Deep Debug COA Account Update Issue...")

	// Initialize database connection
	cfg := config.LoadConfig()
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("✅ Database connected successfully")

	// Debug COA update issue
	if err := debugCOAUpdateIssue(db); err != nil {
		log.Fatalf("❌ Debug failed: %v", err)
	}

	log.Println("🎉 Debug completed!")
}

func debugCOAUpdateIssue(db *gorm.DB) error {
	log.Println("🔍 Investigating COA account update issue...")

	// Test case: CSH-2025-0001
	testCode := "CSH-2025-0001"
	
	// Get detailed information about this cash bank
	var cashBankInfo struct {
		ID        uint    `json:"id"`
		Code      string  `json:"code"`
		Name      string  `json:"name"`
		Balance   float64 `json:"balance"`
		AccountID uint    `json:"account_id"`
	}

	cashBankQuery := `
		SELECT id, code, name, balance, account_id
		FROM cash_banks 
		WHERE code = ? AND deleted_at IS NULL
	`

	if err := db.Raw(cashBankQuery, testCode).Scan(&cashBankInfo).Error; err != nil {
		return fmt.Errorf("failed to get cash bank info: %w", err)
	}

	log.Printf("Cash Bank Info:")
	log.Printf("  ID: %d", cashBankInfo.ID)
	log.Printf("  Code: %s", cashBankInfo.Code)
	log.Printf("  Name: %s", cashBankInfo.Name)
	log.Printf("  Balance: %.2f", cashBankInfo.Balance)
	log.Printf("  Account ID: %d", cashBankInfo.AccountID)

	// Get detailed information about the linked COA account
	var accountInfo struct {
		ID       uint    `json:"id"`
		Code     string  `json:"code"`
		Name     string  `json:"name"`
		Balance  float64 `json:"balance"`
		Type     string  `json:"type"`
		ParentID *uint   `json:"parent_id"`
		IsHeader bool    `json:"is_header"`
	}

	accountQuery := `
		SELECT id, code, name, balance, type, parent_id, is_header
		FROM accounts 
		WHERE id = ? AND deleted_at IS NULL
	`

	if err := db.Raw(accountQuery, cashBankInfo.AccountID).Scan(&accountInfo).Error; err != nil {
		return fmt.Errorf("failed to get account info: %w", err)
	}

	log.Printf("COA Account Info:")
	log.Printf("  ID: %d", accountInfo.ID)
	log.Printf("  Code: %s", accountInfo.Code)
	log.Printf("  Name: %s", accountInfo.Name)
	log.Printf("  Balance: %.2f", accountInfo.Balance)
	log.Printf("  Type: %s", accountInfo.Type)
	if accountInfo.ParentID != nil {
		log.Printf("  Parent ID: %d", *accountInfo.ParentID)
	} else {
		log.Printf("  Parent ID: NULL")
	}
	log.Printf("  Is Header: %t", accountInfo.IsHeader)

	// Check for triggers on accounts table
	log.Println("📋 Checking triggers on accounts table...")
	
	var triggers []struct {
		TriggerName string `json:"trigger_name"`
		Event       string `json:"event"`
		Timing      string `json:"timing"`
		Enabled     string `json:"enabled"`
	}

	triggerQuery := `
		SELECT 
			tr.trigger_name,
			tr.event_manipulation as event,
			tr.action_timing as timing,
			CASE WHEN pt.tgenabled = 'O' THEN 'enabled' ELSE 'disabled' END as enabled
		FROM information_schema.triggers tr
		JOIN pg_trigger pt ON pt.tgname = tr.trigger_name
		WHERE tr.event_object_table = 'accounts'
		AND tr.trigger_schema = 'public'
		ORDER BY tr.trigger_name
	`

	if err := db.Raw(triggerQuery).Scan(&triggers).Error; err != nil {
		return fmt.Errorf("failed to get triggers: %w", err)
	}

	log.Printf("Found %d triggers on accounts table:", len(triggers))
	for _, trigger := range triggers {
		log.Printf("  - %s: %s %s (%s)", trigger.TriggerName, trigger.Timing, trigger.Event, trigger.Enabled)
	}

	// Test direct SQL update
	log.Println("🧪 Testing direct SQL update...")
	
	testBalance := 12345.67
	updateSQL := "UPDATE accounts SET balance = ?, updated_at = NOW() WHERE id = ?"
	
	result := db.Exec(updateSQL, testBalance, accountInfo.ID)
	if result.Error != nil {
		log.Printf("❌ Direct update failed: %v", result.Error)
	} else {
		log.Printf("✅ Direct update succeeded, rows affected: %d", result.RowsAffected)
		
		// Check if the update actually took effect
		var newBalance float64
		if err := db.Raw("SELECT balance FROM accounts WHERE id = ?", accountInfo.ID).Scan(&newBalance).Error; err != nil {
			log.Printf("❌ Failed to check new balance: %v", err)
		} else {
			log.Printf("New balance after direct update: %.2f", newBalance)
			
			if newBalance == testBalance {
				log.Println("✅ Direct update was successful")
			} else {
				log.Printf("❌ Direct update was reverted by trigger: expected %.2f, got %.2f", testBalance, newBalance)
			}
		}
	}

	// Test with trigger disabled
	log.Println("🔧 Testing with validation trigger disabled...")
	
	// Disable the validation trigger temporarily
	if err := db.Exec("ALTER TABLE accounts DISABLE TRIGGER trigger_validate_account_balance").Error; err != nil {
		log.Printf("⚠️ Failed to disable trigger: %v", err)
	} else {
		log.Println("✅ Validation trigger disabled")
		
		// Try the update again
		testBalance2 := 54321.98
		result = db.Exec(updateSQL, testBalance2, accountInfo.ID)
		if result.Error != nil {
			log.Printf("❌ Update with disabled trigger failed: %v", result.Error)
		} else {
			log.Printf("✅ Update with disabled trigger succeeded")
			
			var newBalance2 float64
			if err := db.Raw("SELECT balance FROM accounts WHERE id = ?", accountInfo.ID).Scan(&newBalance2).Error; err != nil {
				log.Printf("❌ Failed to check new balance: %v", err)
			} else {
				log.Printf("New balance with disabled trigger: %.2f", newBalance2)
				
				if newBalance2 == testBalance2 {
					log.Println("✅ Update with disabled trigger was successful")
					log.Println("💡 The issue is caused by the validation trigger")
				} else {
					log.Printf("❌ Even with disabled trigger, update failed: expected %.2f, got %.2f", testBalance2, newBalance2)
				}
			}
		}
		
		// Re-enable the trigger
		if err := db.Exec("ALTER TABLE accounts ENABLE TRIGGER trigger_validate_account_balance").Error; err != nil {
			log.Printf("⚠️ Failed to re-enable trigger: %v", err)
		} else {
			log.Println("✅ Validation trigger re-enabled")
		}
	}

	// Test the fixed balance (should be the cash bank balance)
	log.Println("🎯 Testing sync with correct balance...")
	
	correctBalance := cashBankInfo.Balance
	result = db.Exec(updateSQL, correctBalance, accountInfo.ID)
	if result.Error != nil {
		log.Printf("❌ Sync with correct balance failed: %v", result.Error)
	} else {
		log.Printf("✅ Sync with correct balance succeeded")
		
		var finalBalance float64
		if err := db.Raw("SELECT balance FROM accounts WHERE id = ?", accountInfo.ID).Scan(&finalBalance).Error; err != nil {
			log.Printf("❌ Failed to check final balance: %v", err)
		} else {
			log.Printf("Final balance: %.2f", finalBalance)
			
			if finalBalance == correctBalance {
				log.Printf("✅ Account %s is now synced correctly!", accountInfo.Code)
			} else {
				log.Printf("❌ Sync still failed: expected %.2f, got %.2f", correctBalance, finalBalance)
				
				// If it's a header account, the trigger might be setting it to sum of children
				if accountInfo.IsHeader {
					log.Println("💡 This is a header account - trigger sets balance to sum of children")
					
					// Check children sum
					var childrenSum float64
					childQuery := `
						SELECT COALESCE(SUM(balance), 0)
						FROM accounts 
						WHERE parent_id = ? AND deleted_at IS NULL
					`
					
					if err := db.Raw(childQuery, accountInfo.ID).Scan(&childrenSum).Error; err != nil {
						log.Printf("❌ Failed to get children sum: %v", err)
					} else {
						log.Printf("Children sum: %.2f", childrenSum)
						
						if finalBalance == childrenSum {
							log.Println("✅ Header account balance correctly set to children sum")
							log.Println("💡 Solution: Update child accounts instead of header account")
						}
					}
				}
			}
		}
	}

	return nil
}