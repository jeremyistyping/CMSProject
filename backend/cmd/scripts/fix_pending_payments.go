package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

func main() {
	log.Println("=== FIXING PENDING PAYMENTS ===")
	
	// Connect to database
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	// Get all pending payments
	var pendingPayments []models.Payment
	if err := db.Where("status = ?", "PENDING").
		Preload("Contact").
		Find(&pendingPayments).Error; err != nil {
		log.Fatal("Failed to fetch pending payments:", err)
	}

	if len(pendingPayments) == 0 {
		fmt.Println("No pending payments found!")
		return
	}

	fmt.Printf("Found %d PENDING payments to process\n", len(pendingPayments))
	fmt.Println("========================================")

	for i, payment := range pendingPayments {
		fmt.Printf("\n%d. Processing Payment ID: %d (%s)\n", i+1, payment.ID, payment.Code)
		
		// Check what's missing for this payment
		hasJournal := checkJournalEntry(db, payment.ID)
		hasCashBankTx := checkCashBankTransaction(db, payment.ID)
		
		fmt.Printf("   Has Journal Entry: %t\n", hasJournal)
		fmt.Printf("   Has Cash/Bank Tx: %t\n", hasCashBankTx)
		
		// Fix the payment by completing the missing steps
		if err := completePendingPayment(db, &payment); err != nil {
			fmt.Printf("   ❌ Error fixing payment %d: %v\n", payment.ID, err)
			continue
		}
		
		fmt.Printf("   ✅ Payment %s fixed successfully\n", payment.Code)
	}

	// Verify fixes
	fmt.Println("\n=== VERIFICATION ===")
	var remainingPending int64
	db.Model(&models.Payment{}).Where("status = ?", "PENDING").Count(&remainingPending)
	fmt.Printf("Remaining PENDING payments: %d\n", remainingPending)
	
	if remainingPending == 0 {
		fmt.Println("🎉 All payments fixed successfully!")
	}
}

func checkJournalEntry(db *gorm.DB, paymentID uint) bool {
	var count int64
	db.Model(&models.JournalEntry{}).
		Where("reference_type = ? AND reference_id = ?", "PAYMENT", paymentID).
		Count(&count)
	return count > 0
}

func checkCashBankTransaction(db *gorm.DB, paymentID uint) bool {
	var count int64
	db.Model(&models.CashBankTransaction{}).
		Where("reference_type = ? AND reference_id = ?", "PAYMENT", paymentID).
		Count(&count)
	return count > 0
}

func completePendingPayment(db *gorm.DB, payment *models.Payment) error {
	// Start transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	fmt.Printf("   Processing payment %s (%.2f)\n", payment.Code, payment.Amount)

	// Get payment allocations
	var allocations []models.PaymentAllocation
	if err := tx.Where("payment_id = ?", payment.ID).Find(&allocations).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get allocations: %v", err)
	}

	fmt.Printf("   Found %d allocations\n", len(allocations))

	// Check if this is a payable payment (vendor payment)
	if payment.Method == "PAYABLE" {
		// This is a vendor payment - needs different handling
		fmt.Printf("   Processing as PAYABLE payment\n")
		
		// For payable payments, we need to:
		// 1. Create journal entries (Debit AP, Credit Cash/Bank)
		// 2. Update cash/bank balance (decrease)
		// 3. Update payment status to COMPLETED

		// Find a suitable cash/bank account for vendor payments
		var cashBank models.CashBank
		if err := tx.Where("type = ? AND is_active = ?", "BANK", true).First(&cashBank).Error; err != nil {
			// Fallback to cash account
			if err := tx.Where("type = ? AND is_active = ?", "CASH", true).First(&cashBank).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("no active cash/bank account found")
			}
		}

		fmt.Printf("   Using Cash/Bank Account: %s (ID: %d)\n", cashBank.Name, cashBank.ID)

		// Check if cash/bank has sufficient balance
		if cashBank.Balance < payment.Amount {
			fmt.Printf("   ⚠️ Insufficient balance: %.2f < %.2f\n", cashBank.Balance, payment.Amount)
			// For demo purposes, we'll still process but with a warning
			// In production, you might want to skip or handle this differently
		}

		// Create journal entries if missing
		if !checkJournalEntry(tx, payment.ID) {
			if err := createPayableJournalEntries(tx, payment, cashBank.AccountID); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create journal entries: %v", err)
			}
			fmt.Printf("   ✅ Journal entries created\n")
		}

		// Update cash/bank balance if missing transaction
		if !checkCashBankTransaction(tx, payment.ID) {
			// Update balance (decrease for vendor payment)
			cashBank.Balance -= payment.Amount
			if err := tx.Save(&cashBank).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update cash/bank balance: %v", err)
			}

			// Create transaction record
			cashBankTx := &models.CashBankTransaction{
				CashBankID:      cashBank.ID,
				ReferenceType:   "PAYMENT",
				ReferenceID:     payment.ID,
				Amount:          -payment.Amount, // Negative for outgoing
				BalanceAfter:    cashBank.Balance,
				TransactionDate: payment.Date,
				Notes:           fmt.Sprintf("Vendor Payment %s", payment.Code),
			}
			if err := tx.Create(cashBankTx).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create cash/bank transaction: %v", err)
			}
			fmt.Printf("   ✅ Cash/bank balance updated: %.2f\n", cashBank.Balance)
		}

		// Update payment status
		payment.Status = models.PaymentStatusCompleted
		if err := tx.Save(payment).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update payment status: %v", err)
		}

	} else {
		// This is a receivable payment (customer payment)
		fmt.Printf("   Processing as RECEIVABLE payment\n")
		
		// For receivable payments, we need to:
		// 1. Create journal entries (Debit Cash/Bank, Credit AR)
		// 2. Update cash/bank balance (increase)
		// 3. Update sales outstanding amounts
		// 4. Update payment status to COMPLETED

		// Find default cash account
		var cashBank models.CashBank
		if err := tx.Where("type = ? AND is_active = ?", "CASH", true).First(&cashBank).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("no active cash account found")
		}

		// Create journal entries if missing
		if !checkJournalEntry(tx, payment.ID) {
			if err := createReceivableJournalEntries(tx, payment, cashBank.AccountID); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create journal entries: %v", err)
			}
			fmt.Printf("   ✅ Journal entries created\n")
		}

		// Update cash/bank balance if missing transaction
		if !checkCashBankTransaction(tx, payment.ID) {
			cashBank.Balance += payment.Amount
			if err := tx.Save(&cashBank).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update cash/bank balance: %v", err)
			}

			// Create transaction record
			cashBankTx := &models.CashBankTransaction{
				CashBankID:      cashBank.ID,
				ReferenceType:   "PAYMENT",
				ReferenceID:     payment.ID,
				Amount:          payment.Amount, // Positive for incoming
				BalanceAfter:    cashBank.Balance,
				TransactionDate: payment.Date,
				Notes:           fmt.Sprintf("Customer Payment %s", payment.Code),
			}
			if err := tx.Create(cashBankTx).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create cash/bank transaction: %v", err)
			}
			fmt.Printf("   ✅ Cash/bank balance updated: %.2f\n", cashBank.Balance)
		}

		// Update sales allocations
		for _, alloc := range allocations {
			if alloc.InvoiceID != nil {
				var sale models.Sale
				if err := tx.First(&sale, *alloc.InvoiceID).Error; err == nil {
					sale.PaidAmount += alloc.AllocatedAmount
					sale.OutstandingAmount -= alloc.AllocatedAmount
					if sale.OutstandingAmount <= 0 {
						sale.Status = models.SaleStatusPaid
					}
					tx.Save(&sale)
					fmt.Printf("   ✅ Updated sale %d: Outstanding %.2f\n", sale.ID, sale.OutstandingAmount)
				}
			}
		}

		// Update payment status
		payment.Status = models.PaymentStatusCompleted
		if err := tx.Save(payment).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update payment status: %v", err)
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func createPayableJournalEntries(tx *gorm.DB, payment *models.Payment, cashBankAccountID uint) error {
	// Get Accounts Payable account (2101)
	var apAccount models.Account
	if err := tx.Where("code = ?", "2101").First(&apAccount).Error; err != nil {
		return fmt.Errorf("accounts payable account (2101) not found: %v", err)
	}

	// Create journal entry
	journalEntry := &models.JournalEntry{
		EntryDate:       payment.Date,
		Description:     fmt.Sprintf("Vendor Payment %s", payment.Code),
		ReferenceType:   models.JournalRefPayment,
		ReferenceID:     &payment.ID,
		Reference:       payment.Code,
		UserID:          payment.UserID,
		Status:          models.JournalStatusPosted,
		TotalDebit:      payment.Amount,
		TotalCredit:     payment.Amount,
		IsAutoGenerated: true,
	}

	// Create journal entry
	if err := tx.Create(journalEntry).Error; err != nil {
		return err
	}

	// Journal lines
	journalLines := []models.JournalLine{
		// Debit: Accounts Payable (reduce liability)
		{
			JournalEntryID: journalEntry.ID,
			AccountID:      apAccount.ID,
			Description:    fmt.Sprintf("AP reduction - %s", payment.Code),
			DebitAmount:    payment.Amount,
			CreditAmount:   0,
			LineNumber:     1,
		},
		// Credit: Cash/Bank (reduce asset)
		{
			JournalEntryID: journalEntry.ID,
			AccountID:      cashBankAccountID,
			Description:    fmt.Sprintf("Payment made - %s", payment.Code),
			DebitAmount:    0,
			CreditAmount:   payment.Amount,
			LineNumber:     2,
		},
	}

	// Create journal lines
	for _, line := range journalLines {
		if err := tx.Create(&line).Error; err != nil {
			return err
		}
	}

	return nil
}

func createReceivableJournalEntries(tx *gorm.DB, payment *models.Payment, cashBankAccountID uint) error {
	// Get Accounts Receivable account (1201)
	var arAccount models.Account
	if err := tx.Where("code = ?", "1201").First(&arAccount).Error; err != nil {
		return fmt.Errorf("accounts receivable account (1201) not found: %v", err)
	}

	// Create journal entry
	journalEntry := &models.JournalEntry{
		EntryDate:       payment.Date,
		Description:     fmt.Sprintf("Customer Payment %s", payment.Code),
		ReferenceType:   models.JournalRefPayment,
		ReferenceID:     &payment.ID,
		Reference:       payment.Code,
		UserID:          payment.UserID,
		Status:          models.JournalStatusPosted,
		TotalDebit:      payment.Amount,
		TotalCredit:     payment.Amount,
		IsAutoGenerated: true,
	}

	// Create journal entry
	if err := tx.Create(journalEntry).Error; err != nil {
		return err
	}

	// Journal lines
	journalLines := []models.JournalLine{
		// Debit: Cash/Bank (increase asset)
		{
			JournalEntryID: journalEntry.ID,
			AccountID:      cashBankAccountID,
			Description:    fmt.Sprintf("Payment received - %s", payment.Code),
			DebitAmount:    payment.Amount,
			CreditAmount:   0,
			LineNumber:     1,
		},
		// Credit: Accounts Receivable (reduce asset)
		{
			JournalEntryID: journalEntry.ID,
			AccountID:      arAccount.ID,
			Description:    fmt.Sprintf("AR reduction - %s", payment.Code),
			DebitAmount:    0,
			CreditAmount:   payment.Amount,
			LineNumber:     2,
		},
	}

	// Create journal lines
	for _, line := range journalLines {
		if err := tx.Create(&line).Error; err != nil {
			return err
		}
	}

	return nil
}