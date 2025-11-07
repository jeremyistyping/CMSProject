package main

import (
	"fmt"
	"time"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"gorm.io/gorm"
)

func main() {
	// Load config and connect to database
	_ = config.LoadConfig()
	db := database.ConnectDB()

	fmt.Println("🧪 Testing Corrected Accounting Logic")
	fmt.Println("====================================")

	// Test corrected service
	testCorrectedService(db)

	fmt.Println("\n✅ Testing complete!")
}

func testCorrectedService(db *gorm.DB) {
	correctedService := services.NewCorrectedSSOTSalesJournalService(db)

	// Create a test sale
	testSale := &models.Sale{
		ID:            999999, // Use a high ID to avoid conflicts
		Code:          "TEST-2025-001",
		InvoiceNumber: "INV/TEST/2025/001",
		Status:        models.SaleStatusInvoiced,
		Date:          time.Now(),
		TotalAmount:   3330000, // 3,330,000 IDR
		PPN:           330000,  // 330,000 IDR (11% tax)
		CustomerID:    1,
		Customer: models.Contact{
			ID:   1,
			Name: "Test Customer",
		},
	}

	fmt.Println("\n1. Testing Sales Invoice Journal Entry Creation:")
	fmt.Printf("   Sale Amount: %.2f, PPN: %.2f\n", testSale.TotalAmount, testSale.PPN)
	
	// Test creating sale journal entry
	entry, err := correctedService.CreateSaleJournalEntry(testSale, 1)
	if err != nil {
		fmt.Printf("❌ Failed to create sale journal entry: %v\n", err)
	} else {
		fmt.Printf("✅ Created sale journal entry ID: %d\n", entry.ID)
		totalDebit, _ := entry.TotalDebit.Float64()
		totalCredit, _ := entry.TotalCredit.Float64()
		fmt.Printf("   Total Debit: %.2f, Total Credit: %.2f\n", totalDebit, totalCredit)
		fmt.Printf("   Status: %s, Balanced: %v\n", entry.Status, entry.IsBalanced)
		
		// Show journal lines
		fmt.Println("   Journal Lines:")
		for _, line := range entry.Lines {
			debitAmount, _ := line.DebitAmount.Float64()
			creditAmount, _ := line.CreditAmount.Float64()
			
			if line.Account != nil {
				fmt.Printf("     %s (%d): Dr %.2f, Cr %.2f - %s\n",
					line.Account.Name, line.AccountID, debitAmount, creditAmount, line.Description)
			} else {
				fmt.Printf("     Account ID %d: Dr %.2f, Cr %.2f - %s\n",
					line.AccountID, debitAmount, creditAmount, line.Description)
			}
		}
	}

	fmt.Println("\n2. Testing Payment Journal Entry Creation:")
	
	// Create a test payment
	testPayment := &models.SalePayment{
		ID:            999999, // Use high ID to avoid conflicts
		SaleID:        testSale.ID,
		Amount:        3330000,
		PaymentDate:   time.Now(),
		PaymentMethod: "BANK_TRANSFER",
	}

	// Test creating payment journal entry
	paymentEntry, err := correctedService.CreatePaymentJournalEntry(testPayment, 1)
	if err != nil {
		fmt.Printf("❌ Failed to create payment journal entry: %v\n", err)
	} else {
		fmt.Printf("✅ Created payment journal entry ID: %d\n", paymentEntry.ID)
		totalDebit, _ := paymentEntry.TotalDebit.Float64()
		totalCredit, _ := paymentEntry.TotalCredit.Float64()
		fmt.Printf("   Total Debit: %.2f, Total Credit: %.2f\n", totalDebit, totalCredit)
		fmt.Printf("   Status: %s, Balanced: %v\n", paymentEntry.Status, paymentEntry.IsBalanced)

		// Show journal lines
		fmt.Println("   Journal Lines:")
		for _, line := range paymentEntry.Lines {
			debitAmount, _ := line.DebitAmount.Float64()
			creditAmount, _ := line.CreditAmount.Float64()
			
			if line.Account != nil {
				fmt.Printf("     %s (%d): Dr %.2f, Cr %.2f - %s\n",
					line.Account.Name, line.AccountID, debitAmount, creditAmount, line.Description)
			} else {
				fmt.Printf("     Account ID %d: Dr %.2f, Cr %.2f - %s\n",
					line.AccountID, debitAmount, creditAmount, line.Description)
			}
		}
	}

	fmt.Println("\n3. Expected Accounting Impact:")
	fmt.Println("   Sale Invoice Entry:")
	fmt.Println("     Dr. Piutang Usaha (AR)     3,330,000")
	fmt.Println("         Cr. Sales Revenue         3,000,000")
	fmt.Println("         Cr. PPN Payable             330,000")
	fmt.Println("   Payment Entry:")
	fmt.Println("     Dr. Bank Account           3,330,000")
	fmt.Println("         Cr. Piutang Usaha         3,330,000")
	
	fmt.Println("\n4. Expected Account Balance Changes:")
	fmt.Println("   Piutang Usaha: Should increase by 3,330,000 then decrease by 3,330,000 = Net 0")
	fmt.Println("   Sales Revenue: Should increase by -3,000,000 (credit balance)")
	fmt.Println("   PPN Payable: Should increase by -330,000 (credit balance)")
	fmt.Println("   Bank Account: Should increase by 3,330,000 (debit balance)")
}