package main

import (
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	log.Printf("🔧 Comprehensive Balance Synchronization Fix")

	db := database.ConnectDB()

	// Step 1: Fix CashBank balances based on transactions
	log.Printf("\n💰 Step 1: Fixing CashBank Balances")
	
	var cashBanks []models.CashBank
	if err := db.Find(&cashBanks).Error; err != nil {
		log.Fatalf("❌ Failed to get cashbanks: %v", err)
	}

	for _, cashBank := range cashBanks {
		// Get latest transaction balance
		var latestTxn models.CashBankTransaction
		if err := db.Where("cash_bank_id = ?", cashBank.ID).Order("id DESC").First(&latestTxn).Error; err == nil {
			if cashBank.Balance != latestTxn.BalanceAfter {
				log.Printf("🔄 Fixing CashBank %d (%s): %.2f -> %.2f", 
					cashBank.ID, cashBank.Name, cashBank.Balance, latestTxn.BalanceAfter)
				
				// Update CashBank balance
				if err := db.Model(&cashBank).Update("balance", latestTxn.BalanceAfter).Error; err != nil {
					log.Printf("❌ Failed to update CashBank %d: %v", cashBank.ID, err)
				} else {
					log.Printf("✅ CashBank %d balance updated", cashBank.ID)
				}
			} else {
				log.Printf("✅ CashBank %d balance already correct: %.2f", cashBank.ID, cashBank.Balance)
			}
		} else {
			log.Printf("⚠️ No transactions found for CashBank %d", cashBank.ID)
		}
	}

	// Step 2: Fix GL Account balances based on SSOT journal entries
	log.Printf("\n📊 Step 2: Fixing GL Account Balances based on SSOT")

	// Get all accounts that have journal lines
	var accountsWithJournalLines []struct {
		AccountID uint64
	}
	
	err := db.Raw("SELECT DISTINCT account_id FROM unified_journal_lines").Scan(&accountsWithJournalLines).Error
	if err != nil {
		log.Printf("❌ Failed to get accounts with journal lines: %v", err)
		return
	}

	log.Printf("📋 Found %d accounts with journal entries", len(accountsWithJournalLines))

	for _, accountData := range accountsWithJournalLines {
		accountID := accountData.AccountID
		
		// Calculate correct balance using raw SQL (works correctly)
		var result struct {
			TotalDebits  float64
			TotalCredits float64
		}
		
		err := db.Raw(`
			SELECT 
				COALESCE(SUM(ujl.debit_amount), 0) as total_debits,
				COALESCE(SUM(ujl.credit_amount), 0) as total_credits
			FROM unified_journal_lines ujl
			JOIN unified_journal_ledger uj ON ujl.journal_id = uj.id
			WHERE ujl.account_id = ? AND uj.status = 'POSTED'
		`, accountID).Scan(&result).Error
		
		if err != nil {
			log.Printf("❌ Failed to calculate balance for account %d: %v", accountID, err)
			continue
		}

		// Get account details
		var account models.Account
		if err := db.First(&account, accountID).Error; err != nil {
			log.Printf("❌ Account %d not found: %v", accountID, err)
			continue
		}

		// Calculate expected balance based on account type
		var expectedBalance float64
		switch account.Type {
		case "Asset", "Expense":
			// Normal debit balance accounts: Debit increases, Credit decreases
			expectedBalance = result.TotalDebits - result.TotalCredits
		case "Liability", "Equity", "Revenue":
			// Normal credit balance accounts: Credit increases, Debit decreases
			expectedBalance = result.TotalCredits - result.TotalDebits
		default:
			// Default to debit balance behavior
			expectedBalance = result.TotalDebits - result.TotalCredits
		}

		if account.Balance != expectedBalance {
			log.Printf("🔄 Fixing Account %d (%s - %s): %.2f -> %.2f", 
				account.ID, account.Code, account.Name, account.Balance, expectedBalance)
			
			// Update account balance
			err := db.Model(&account).Updates(map[string]interface{}{
				"balance": expectedBalance,
				"updated_at": db.NowFunc(),
			}).Error
			
			if err != nil {
				log.Printf("❌ Failed to update Account %d: %v", account.ID, err)
			} else {
				log.Printf("✅ Account %d balance updated successfully", account.ID)
			}
		} else {
			log.Printf("✅ Account %d (%s) balance already correct: %.2f", 
				account.ID, account.Code, account.Balance)
		}
	}

	// Step 3: Sync CashBank GL Accounts with CashBank table balances
	log.Printf("\n🔄 Step 3: Syncing CashBank GL Accounts with CashBank Table")
	
	for _, cashBank := range cashBanks {
		// Re-fetch updated cashbank balance
		var updatedCashBank models.CashBank
		if err := db.First(&updatedCashBank, cashBank.ID).Error; err != nil {
			continue
		}

		// Get related GL account
		var glAccount models.Account
		if err := db.First(&glAccount, updatedCashBank.AccountID).Error; err != nil {
			log.Printf("❌ GL Account %d for CashBank %d not found: %v", 
				updatedCashBank.AccountID, updatedCashBank.ID, err)
			continue
		}

		if glAccount.Balance != updatedCashBank.Balance {
			log.Printf("🔄 Syncing GL Account %d (%s) with CashBank %d: %.2f -> %.2f",
				glAccount.ID, glAccount.Code, updatedCashBank.ID, glAccount.Balance, updatedCashBank.Balance)
			
			err := db.Model(&glAccount).Updates(map[string]interface{}{
				"balance": updatedCashBank.Balance,
				"updated_at": db.NowFunc(),
			}).Error
			
			if err != nil {
				log.Printf("❌ Failed to sync GL Account %d: %v", glAccount.ID, err)
			} else {
				log.Printf("✅ GL Account %d synced with CashBank successfully", glAccount.ID)
			}
		} else {
			log.Printf("✅ GL Account %d (%s) already synced with CashBank %d: %.2f", 
				glAccount.ID, glAccount.Code, updatedCashBank.ID, glAccount.Balance)
		}
	}

	// Step 4: Update header account balances
	log.Printf("\n📈 Step 4: Updating Header Account Balances")
	
	headerMappings := map[string]string{
		"1000": "1%",   // ASSETS - all 1xxx accounts
		"1100": "11%",  // CURRENT ASSETS - all 11xx accounts  
		"1200": "12%",  // ACCOUNTS RECEIVABLE - all 12xx accounts
		"2000": "2%",   // LIABILITIES - all 2xxx accounts
		"2100": "21%",  // CURRENT LIABILITIES - all 21xx accounts
		"3000": "3%",   // EQUITY - all 3xxx accounts
		"4000": "4%",   // REVENUE - all 4xxx accounts
		"5000": "5%",   // EXPENSES - all 5xxx accounts
	}
	
	for headerCode, childPattern := range headerMappings {
		// Get header account
		var headerAccount models.Account
		if err := db.Where("code = ? AND is_header = ?", headerCode, true).First(&headerAccount).Error; err != nil {
			log.Printf("⚠️ Header account %s not found", headerCode)
			continue
		}
		
		// Calculate sum of children
		var childrenSum float64
		err := db.Model(&models.Account{}).
			Where("code LIKE ? AND code != ? AND is_header = ? AND deleted_at IS NULL", 
				childPattern, headerCode, false).
			Select("COALESCE(SUM(balance), 0)").
			Scan(&childrenSum).Error
			
		if err != nil {
			log.Printf("❌ Failed to calculate children sum for %s: %v", headerCode, err)
			continue
		}
		
		if headerAccount.Balance != childrenSum {
			log.Printf("🔄 Updating header account %s: %.2f -> %.2f", 
				headerCode, headerAccount.Balance, childrenSum)
			
			err := db.Model(&headerAccount).Updates(map[string]interface{}{
				"balance": childrenSum,
				"updated_at": db.NowFunc(),
			}).Error
			
			if err != nil {
				log.Printf("❌ Failed to update header account %s: %v", headerCode, err)
			} else {
				log.Printf("✅ Header account %s updated successfully", headerCode)
			}
		} else {
			log.Printf("✅ Header account %s balance already correct: %.2f", headerCode, headerAccount.Balance)
		}
	}

	// Step 5: Final verification
	log.Printf("\n🔍 Step 5: Final Verification")
	
	// Check CashBank ID 2 specifically
	var finalCashBank models.CashBank
	var finalGLAccount models.Account
	
	if err := db.First(&finalCashBank, 2).Error; err == nil {
		if err := db.First(&finalGLAccount, finalCashBank.AccountID).Error; err == nil {
			log.Printf("🎯 Final Check - CashBank ID 2:")
			log.Printf("   CashBank Balance: %.2f", finalCashBank.Balance)
			log.Printf("   GL Account Balance: %.2f", finalGLAccount.Balance)
			
			if finalCashBank.Balance == finalGLAccount.Balance {
				log.Printf("✅ CashBank and GL Account are now synchronized!")
			} else {
				log.Printf("❌ Still not synchronized!")
			}
		}
	}

	log.Printf("\n🎉 Comprehensive Balance Synchronization Fix Completed!")
	log.Printf("📋 Summary:")
	log.Printf("   - Fixed CashBank balances based on transaction records")
	log.Printf("   - Fixed GL Account balances based on SSOT journal entries") 
	log.Printf("   - Synced CashBank GL accounts with CashBank table balances")
	log.Printf("   - Updated header account balances")
	log.Printf("   - All balances should now be consistent and accurate")
}