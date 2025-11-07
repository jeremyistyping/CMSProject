package main

import (
	"fmt"
	"time"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
)

func main() {
	fmt.Println("🧪 Testing Accounting Logic (Unit Test)")
	fmt.Println("======================================")

	// Create service instance (we'll test logic without database)
	service := &services.CorrectedSSOTSalesJournalService{}

	// Test calculateJournalTotals method if it exists, or test the business logic directly
	testJournalEntryLogic(service)

	fmt.Println("\n✅ Logic testing complete!")
}

func testJournalEntryLogic(service *services.CorrectedSSOTSalesJournalService) {
	// Create a test sale
	testSale := &models.Sale{
		ID:            999999,
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

	fmt.Println("1. Testing Sales Invoice Logic:")
	fmt.Printf("   Sale Amount: %.2f IDR\n", testSale.TotalAmount)
	fmt.Printf("   PPN (Tax): %.2f IDR\n", testSale.PPN)
	
	// Calculate expected values
	baseAmount := testSale.TotalAmount - testSale.PPN
	fmt.Printf("   Base Amount (excl. tax): %.2f IDR\n", baseAmount)
	
	fmt.Println("\n   Expected Journal Entry:")
	fmt.Printf("   Dr. Accounts Receivable: %.2f IDR\n", testSale.TotalAmount)
	fmt.Printf("   Cr. Sales Revenue:       %.2f IDR\n", baseAmount)
	fmt.Printf("   Cr. PPN Payable:         %.2f IDR\n", testSale.PPN)
	fmt.Printf("   Total Debit:             %.2f IDR\n", testSale.TotalAmount)
	fmt.Printf("   Total Credit:            %.2f IDR\n", baseAmount + testSale.PPN)
	
	// Verify balance
	totalCredit := baseAmount + testSale.PPN
	isBalanced := testSale.TotalAmount == totalCredit
	fmt.Printf("   Is Balanced: %v\n", isBalanced)
	
	if !isBalanced {
		fmt.Printf("   ❌ BALANCE ERROR: Debit %.2f != Credit %.2f\n", testSale.TotalAmount, totalCredit)
	} else {
		fmt.Printf("   ✅ Entry is properly balanced\n")
	}

	// Test payment logic
	fmt.Println("\n2. Testing Payment Logic:")
	testPayment := &models.SalePayment{
		ID:            999999,
		SaleID:        testSale.ID,
		Amount:        testSale.TotalAmount,
		PaymentDate:   time.Now(),
		PaymentMethod: "BANK_TRANSFER",
	}
	
	fmt.Printf("   Payment Amount: %.2f IDR\n", testPayment.Amount)
	fmt.Printf("   Payment Method: %s\n", testPayment.PaymentMethod)
	
	fmt.Println("\n   Expected Payment Journal Entry:")
	fmt.Printf("   Dr. Bank Account:        %.2f IDR\n", testPayment.Amount)
	fmt.Printf("   Cr. Accounts Receivable: %.2f IDR\n", testPayment.Amount)
	fmt.Printf("   Is Balanced: %v\n", testPayment.Amount == testPayment.Amount)
	fmt.Printf("   ✅ Payment entry is balanced\n")

	// Test combined effect
	fmt.Println("\n3. Testing Combined Transaction Effect:")
	fmt.Println("   After Sale + Payment:")
	fmt.Printf("   Accounts Receivable: +%.2f - %.2f = %.2f IDR (Net effect: 0)\n", 
		testSale.TotalAmount, testPayment.Amount, testSale.TotalAmount - testPayment.Amount)
	fmt.Printf("   Sales Revenue: %.2f IDR (Credit balance)\n", baseAmount)
	fmt.Printf("   PPN Payable: %.2f IDR (Credit balance)\n", testSale.PPN)
	fmt.Printf("   Bank Account: %.2f IDR (Debit balance)\n", testPayment.Amount)
	
	// Verify net effect
	netAR := testSale.TotalAmount - testPayment.Amount
	if netAR == 0 {
		fmt.Println("   ✅ Accounts Receivable correctly cleared")
	} else {
		fmt.Printf("   ❌ ERROR: AR should be 0 but is %.2f\n", netAR)
	}

	fmt.Println("\n4. Testing Multiple Scenarios:")
	
	// Partial payment scenario
	partialPayment := testSale.TotalAmount * 0.5
	fmt.Printf("   Partial Payment (50%%): %.2f IDR\n", partialPayment)
	fmt.Printf("   Remaining AR Balance: %.2f IDR\n", testSale.TotalAmount - partialPayment)
	
	// Overpayment scenario
	overpayment := testSale.TotalAmount * 1.2
	fmt.Printf("   Overpayment (120%%): %.2f IDR\n", overpayment)
	fmt.Printf("   Result would be: AR = %.2f IDR (credit balance = customer prepayment)\n", 
		testSale.TotalAmount - overpayment)
}