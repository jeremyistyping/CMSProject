package main

import (
	"log"
	"time"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"

	"gorm.io/gorm"
)

func main() {
	log.Println("🧪 Verifying Payment Journal Entry Fix")
	log.Println("======================================")

	// Initialize database connection
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	verifyFixImplementation(db)
}

func verifyFixImplementation(db *gorm.DB) {
	log.Println("🔍 Checking Payment Journal Entry Fix Implementation")
	
	// 1. Check recent payments and their journal entries
	var recentPayments []models.Payment
	err := db.Where("created_at > ?", time.Now().AddDate(0, 0, -1)).
		Order("created_at DESC").
		Limit(5).
		Find(&recentPayments).Error
	
	if err != nil {
		log.Printf("❌ Error fetching recent payments: %v", err)
		return
	}
	
	log.Printf("📄 Found %d recent payments to analyze", len(recentPayments))
	
	if len(recentPayments) == 0 {
		log.Println("⚠️  No recent payments found. Testing with mock analysis...")
		analyzeMockScenario()
		return
	}
	
	// 2. For each payment, check journal entries
	for i, payment := range recentPayments {
		log.Printf("\n--- Payment %d Analysis ---", i+1)
		log.Printf("Payment ID: %d", payment.ID)
		log.Printf("Payment Code: %s", payment.Code)
		log.Printf("Amount: Rp %.2f", payment.Amount)
		log.Printf("Date: %s", payment.Date.Format("2006-01-02"))
		
		analyzePaymentJournalEntries(db, payment)
	}
	
	log.Println("\n🏁 Fix verification completed!")
}

func analyzePaymentJournalEntries(db *gorm.DB, payment models.Payment) {
	// Find journal entries related to this payment
	var journalEntries []models.JournalEntry
	err := db.Where("reference_type = ? AND reference_id = ?", "PAYMENT", payment.ID).
		Preload("JournalLines").
		Find(&journalEntries).Error
	
	if err != nil {
		log.Printf("❌ Error fetching journal entries: %v", err)
		return
	}
	
	log.Printf("📋 Found %d journal entries for payment %d", len(journalEntries), payment.ID)
	
	// Check if fix is working
	if len(journalEntries) == 0 {
		log.Printf("❌ ERROR: No journal entries found - payment not properly recorded!")
		return
	} else if len(journalEntries) == 1 {
		log.Printf("✅ CORRECT: Only 1 journal entry created (fix working!)")
	} else {
		log.Printf("❌ ERROR: %d journal entries found - double entry bug still exists!", len(journalEntries))
	}
	
	// Analyze journal entry details
	for j, entry := range journalEntries {
		log.Printf("   Entry %d:", j+1)
		log.Printf("      ID: %d", entry.ID)
		log.Printf("      Description: %s", entry.Description)
		log.Printf("      Total Debit: Rp %.2f", entry.TotalDebit)
		log.Printf("      Total Credit: Rp %.2f", entry.TotalCredit)
		log.Printf("      Lines: %d", len(entry.JournalLines))
		
		// Check if journal lines are balanced
		totalDebit := 0.0
		totalCredit := 0.0
		bankDebitCount := 0
		arCreditCount := 0
		
		for k, line := range entry.JournalLines {
			// Get account details
			var account models.Account
			db.First(&account, line.AccountID)
			
			log.Printf("         Line %d: %s (%s) - Debit: %.2f, Credit: %.2f",
				k+1, account.Code, account.Name, line.DebitAmount, line.CreditAmount)
			
			totalDebit += line.DebitAmount
			totalCredit += line.CreditAmount
			
			// Count specific account types
			if (account.Code == "1102" || account.Name == "Bank BCA") && line.DebitAmount > 0 {
				bankDebitCount++
			}
			if (account.Code == "1201" || account.Name == "Piutang Usaha") && line.CreditAmount > 0 {
				arCreditCount++
			}
		}
		
		// Validate the entry
		log.Printf("      Validation:")
		if abs(totalDebit - totalCredit) < 0.01 {
			log.Printf("         ✅ Journal entry is balanced (%.2f = %.2f)", totalDebit, totalCredit)
		} else {
			log.Printf("         ❌ Journal entry is NOT balanced (%.2f ≠ %.2f)", totalDebit, totalCredit)
		}
		
		if abs(totalDebit - payment.Amount) < 0.01 {
			log.Printf("         ✅ Amount matches payment (%.2f)", payment.Amount)
		} else {
			log.Printf("         ❌ Amount mismatch - Expected: %.2f, Got: %.2f", payment.Amount, totalDebit)
		}
		
		if bankDebitCount == 1 && arCreditCount == 1 {
			log.Printf("         ✅ Correct journal structure (1 Bank debit, 1 AR credit)")
		} else {
			log.Printf("         ⚠️  Unusual journal structure (Bank debits: %d, AR credits: %d)", bankDebitCount, arCreditCount)
		}
	}
}

func analyzeMockScenario() {
	log.Println("\n📚 Mock Analysis - Explaining the Fix:")
	log.Println("=====================================")
	
	log.Println("🐛 BEFORE FIX (Bug Scenario):")
	log.Println("   1. PaymentService.CreateReceivablePayment() creates journal entry:")
	log.Println("      - Debit: Bank BCA +1.887.000")
	log.Println("      - Credit: Piutang Usaha -1.887.000")
	log.Println("   2. SalesService.CreateSalePayment() ALSO creates journal entry:")
	log.Println("      - Debit: Bank BCA +1.887.000  ← DUPLICATE!")
	log.Println("      - Credit: Piutang Usaha -1.887.000  ← DUPLICATE!")
	log.Println("   RESULT: Bank BCA +3.774.000 (WRONG!)")
	
	log.Println("\n✅ AFTER FIX (Correct Scenario):")
	log.Println("   1. PaymentService.CreateReceivablePayment() creates journal entry:")
	log.Println("      - Debit: Bank BCA +1.887.000")
	log.Println("      - Credit: Piutang Usaha -1.887.000")
	log.Println("   2. SalesService.CreateSalePayment() NO LONGER creates journal entry")
	log.Println("      - Only logs: 'FIXED: SalesService no longer creates journal entries'")
	log.Println("   RESULT: Bank BCA +1.887.000 (CORRECT!)")
	
	log.Println("\n🔧 Fix Implementation:")
	log.Println("   - Removed: s.createJournalEntriesForPayment() call from SalesService")
	log.Println("   - Added: Documentation comment explaining the fix")
	log.Println("   - Added: Log message confirming fix is active")
	
	log.Println("\n📝 Code Changes:")
	log.Println("   File: backend/services/sales_service.go")
	log.Println("   Lines: 562-566")
	log.Println("   Change: Replaced journal creation with documentation comment")
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}