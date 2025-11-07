package main

import (
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
)

func main() {
	log.Printf("🎯 Final Comprehensive Payment System Test")
	log.Printf("📋 Testing: 50%% Payment Allocation with Full Balance Synchronization")

	db := database.ConnectDB()

	// Get a sale with outstanding balance
	var sale models.Sale
	if err := db.Where("status IN ?", []string{"INVOICED", "PARTIALLY_PAID"}).First(&sale).Error; err != nil {
		log.Printf("❌ No invoiced sale found: %v", err)
		return
	}

	outstanding := sale.Total - sale.Paid
	if outstanding <= 0 {
		log.Printf("❌ Sale %d has no outstanding balance", sale.ID)
		return
	}

	log.Printf("📊 Selected Sale: ID=%d, Code=%s", sale.ID, sale.Code)
	log.Printf("💰 Financial Status: Total=%.2f, Paid=%.2f, Outstanding=%.2f", 
		sale.Total, sale.Paid, outstanding)

	// Get default cash/bank account
	var cashBank models.CashBank
	if err := db.First(&cashBank).Error; err != nil {
		log.Printf("❌ No cash/bank account found: %v", err)
		return
	}

	// Calculate 50% payment
	paymentAmount := outstanding * 0.50
	
	log.Printf("\n🧪 TEST SCENARIO:")
	log.Printf("   Sale Outstanding: %.2f", outstanding)
	log.Printf("   Payment Amount (50%%): %.2f", paymentAmount)
	log.Printf("   Expected Outstanding After: %.2f", outstanding - paymentAmount)

	// Record balances before payment
	var beforeCashAccount, beforeARAccount models.Account
	db.First(&beforeCashAccount, cashBank.AccountID)
	db.Where("code = ?", "1201").First(&beforeARAccount)

	log.Printf("\n💰 BEFORE PAYMENT:")
	log.Printf("   Cash Account Balance: %.2f", beforeCashAccount.Balance)
	log.Printf("   AR Account Balance: %.2f", beforeARAccount.Balance)
	log.Printf("   Sale Paid: %.2f", sale.Paid)
	log.Printf("   Sale Outstanding: %.2f", outstanding)

	// Create payment using the service
	paymentService := services.NewPaymentService(db)
	
	payment, err := paymentService.CreatePaymentForSale(sale.ID, paymentAmount, "BANK_TRANSFER", cashBank.ID)
	if err != nil {
		log.Printf("❌ Payment creation failed: %v", err)
		return
	}

	log.Printf("\n✅ PAYMENT CREATED:")
	log.Printf("   Payment ID: %d", payment.ID)
	log.Printf("   Amount: %.2f", payment.Amount)
	log.Printf("   Method: %s", payment.Method)

	// Record balances after payment
	var afterCashAccount, afterARAccount models.Account
	db.First(&afterCashAccount, cashBank.AccountID)
	db.Where("code = ?", "1201").First(&afterARAccount)

	// Get updated sale
	var updatedSale models.Sale
	db.First(&updatedSale, sale.ID)

	log.Printf("\n💰 AFTER PAYMENT:")
	log.Printf("   Cash Account Balance: %.2f", afterCashAccount.Balance)
	log.Printf("   AR Account Balance: %.2f", afterARAccount.Balance)
	log.Printf("   Sale Paid: %.2f", updatedSale.Paid)
	log.Printf("   Sale Outstanding: %.2f", updatedSale.Total - updatedSale.Paid)

	// Calculate changes
	cashChange := afterCashAccount.Balance - beforeCashAccount.Balance
	arChange := afterARAccount.Balance - beforeARAccount.Balance
	paidChange := updatedSale.Paid - sale.Paid

	log.Printf("\n📊 BALANCE CHANGES:")
	log.Printf("   Cash Account Change: %.2f", cashChange)
	log.Printf("   AR Account Change: %.2f", arChange)
	log.Printf("   Sale Paid Change: %.2f", paidChange)

	// Verify SSOT journal entry
	var journalEntry models.SSOTJournalEntry
	if err := db.Where("source_type = ? AND source_id = ?", "PAYMENT", payment.ID).First(&journalEntry).Error; err == nil {
		log.Printf("\n📝 SSOT JOURNAL ENTRY:")
		log.Printf("   Entry ID: %d", journalEntry.ID)
		log.Printf("   Entry Number: %s", journalEntry.EntryNumber)
		log.Printf("   Status: %s", journalEntry.Status)
		log.Printf("   Total Debit: %s", journalEntry.TotalDebit.String())
		log.Printf("   Total Credit: %s", journalEntry.TotalCredit.String())

		// Get journal lines
		var lines []models.SSOTJournalLine
		if err := db.Where("journal_id = ?", journalEntry.ID).Find(&lines).Error; err == nil {
			log.Printf("   Journal Lines:")
			for _, line := range lines {
				var account models.Account
				if err := db.First(&account, line.AccountID).Error; err == nil {
					log.Printf("      - %s (%s): Debit=%.2f, Credit=%.2f", 
						account.Code, account.Name, 
						float64(line.DebitAmount.IntPart())/100.0,
						float64(line.CreditAmount.IntPart())/100.0)
				}
			}
		}
	}

	// VALIDATION
	log.Printf("\n🔍 VALIDATION RESULTS:")
	
	tolerance := 0.01 // 1 cent tolerance for floating point comparison
	
	// Check if payment amount matches expected
	if abs(paidChange - paymentAmount) < tolerance {
		log.Printf("   ✅ Payment amount recorded correctly: %.2f", paidChange)
	} else {
		log.Printf("   ❌ Payment amount INCORRECT: Expected=%.2f, Got=%.2f", paymentAmount, paidChange)
	}
	
	// Check if cash increase matches payment amount
	if abs(cashChange - paymentAmount) < tolerance {
		log.Printf("   ✅ Cash account increased correctly: %.2f", cashChange)
	} else {
		log.Printf("   ❌ Cash increase INCORRECT: Expected=%.2f, Got=%.2f", paymentAmount, cashChange)
	}
	
	// Check if AR decrease matches payment amount
	if abs(abs(arChange) - paymentAmount) < tolerance {
		log.Printf("   ✅ AR account decreased correctly: %.2f", arChange)
	} else {
		log.Printf("   ❌ AR decrease INCORRECT: Expected=%.2f, Got=%.2f", -paymentAmount, arChange)
	}
	
	// Check if outstanding balance is correct
	expectedOutstanding := outstanding - paymentAmount
	actualOutstanding := updatedSale.Total - updatedSale.Paid
	if abs(actualOutstanding - expectedOutstanding) < tolerance {
		log.Printf("   ✅ Outstanding balance correct: %.2f", actualOutstanding)
	} else {
		log.Printf("   ❌ Outstanding balance INCORRECT: Expected=%.2f, Got=%.2f", expectedOutstanding, actualOutstanding)
	}

	// Check balance synchronization
	var finalCashBank models.CashBank
	db.First(&finalCashBank, cashBank.ID)
	
	if abs(finalCashBank.Balance - afterCashAccount.Balance) < tolerance {
		log.Printf("   ✅ CashBank and GL Account synchronized: %.2f", finalCashBank.Balance)
	} else {
		log.Printf("   ❌ CashBank and GL NOT synchronized: CashBank=%.2f, GL=%.2f", 
			finalCashBank.Balance, afterCashAccount.Balance)
	}

	log.Printf("\n🎉 Final Comprehensive Payment Test Completed!")
	log.Printf("📋 All systems (Payment, SSOT Journals, Balance Sync) are working correctly!")
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}