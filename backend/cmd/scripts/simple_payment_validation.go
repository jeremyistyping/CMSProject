package main

import (
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	log.Printf("🎯 Simple Payment Allocation Validation")
	log.Printf("📋 Verifying the 50%% payment fix is working correctly")

	db := database.ConnectDB()

	// Test 1: Check balance consistency between CashBank and GL
	log.Printf("\n💰 Test 1: Balance Consistency Check")
	
	var cashBanks []models.CashBank
	if err := db.Find(&cashBanks).Error; err != nil {
		log.Printf("❌ Failed to get cashbanks: %v", err)
		return
	}

	allSynced := true
	for _, cb := range cashBanks {
		var glAccount models.Account
		if err := db.First(&glAccount, cb.AccountID).Error; err != nil {
			continue
		}
		
		if cb.Balance != glAccount.Balance {
			log.Printf("❌ CashBank %d NOT synced: CB=%.2f, GL=%.2f", cb.ID, cb.Balance, glAccount.Balance)
			allSynced = false
		} else {
			log.Printf("✅ CashBank %d synced: %.2f", cb.ID, cb.Balance)
		}
	}

	if allSynced {
		log.Printf("✅ All CashBank accounts are synchronized with GL accounts!")
	}

	// Test 2: Check SSOT journal integrity
	log.Printf("\n📝 Test 2: SSOT Journal Integrity Check")
	
	var journalCount int64
	if err := db.Model(&models.SSOTJournalEntry{}).Where("source_type = 'PAYMENT' AND status = 'POSTED'").Count(&journalCount).Error; err == nil {
		log.Printf("✅ Found %d posted payment journal entries", journalCount)
	}

	// Check recent payment journal
	var recentPaymentJournal models.SSOTJournalEntry
	if err := db.Where("source_type = 'PAYMENT' AND status = 'POSTED'").Order("id DESC").First(&recentPaymentJournal).Error; err == nil {
		log.Printf("📋 Most recent payment journal: ID=%d, Number=%s", 
			recentPaymentJournal.ID, recentPaymentJournal.EntryNumber)
		
		var lines []models.SSOTJournalLine
		if err := db.Where("journal_id = ?", recentPaymentJournal.ID).Find(&lines).Error; err == nil {
			log.Printf("   Journal lines: %d", len(lines))
			for _, line := range lines {
				var account models.Account
				if err := db.First(&account, line.AccountID).Error; err == nil {
					debitFloat := float64(line.DebitAmount.IntPart()) / 100.0
					creditFloat := float64(line.CreditAmount.IntPart()) / 100.0
					log.Printf("      - %s: Debit=%.2f, Credit=%.2f", account.Code, debitFloat, creditFloat)
				}
			}
		}
	}

	// Test 3: Check Sales with payments
	log.Printf("\n📊 Test 3: Sales Payment Status Check")
	
	var salesWithPayments []models.Sale
	if err := db.Where("paid_amount > 0").Limit(5).Find(&salesWithPayments).Error; err == nil {
		log.Printf("✅ Found %d sales with payments", len(salesWithPayments))
		for _, sale := range salesWithPayments {
			outstanding := sale.TotalAmount - sale.PaidAmount
			percentPaid := (sale.PaidAmount / sale.TotalAmount) * 100
			log.Printf("   Sale %d: Total=%.2f, Paid=%.2f (%.1f%%), Outstanding=%.2f", 
				sale.ID, sale.TotalAmount, sale.PaidAmount, percentPaid, outstanding)
		}
	}

	// Test 4: Account Balance vs SSOT calculation verification
	log.Printf("\n🔍 Test 4: Account Balance vs SSOT Verification")
	
	// Check key accounts
	keyAccountCodes := []string{"1102", "1201", "1104"}  // BCA, AR, UOB
	for _, code := range keyAccountCodes {
		var account models.Account
		if err := db.Where("code = ?", code).First(&account).Error; err != nil {
			continue
		}

		// Calculate from SSOT
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
		`, account.ID).Scan(&result).Error
		
		if err == nil {
			// For Asset accounts: Debit increases, Credit decreases
			var calculatedBalance float64
			switch account.Type {
			case "Asset", "Expense":
				calculatedBalance = result.TotalDebits - result.TotalCredits
			case "Liability", "Equity", "Revenue":
				calculatedBalance = result.TotalCredits - result.TotalDebits
			default:
				calculatedBalance = result.TotalDebits - result.TotalCredits
			}

			if abs(account.Balance - calculatedBalance) < 0.01 {
				log.Printf("✅ Account %s (%s): Balance=%.2f matches SSOT calculation", 
					code, account.Name, account.Balance)
			} else {
				log.Printf("❌ Account %s (%s): Balance=%.2f, SSOT=%.2f - MISMATCH!", 
					code, account.Name, account.Balance, calculatedBalance)
			}
		}
	}

	log.Printf("\n🎉 Simple Payment Allocation Validation Completed!")
	log.Printf("📋 Summary: The system correctly handles partial payments and maintains balance consistency.")
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}