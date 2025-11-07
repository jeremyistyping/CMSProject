package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Purchase struct {
	ID                uint    `json:"id" gorm:"primaryKey"`
	Code              string  `json:"code"`
	TotalAmount       float64 `json:"total_amount"`
	PaidAmount        float64 `json:"paid_amount"`
	OutstandingAmount float64 `json:"outstanding_amount"`
	PaymentMethod     string  `json:"payment_method"`
	Status            string  `json:"status"`
	VendorID          uint    `json:"vendor_id"`
}

type PurchasePayment struct {
	ID         uint    `json:"id" gorm:"primaryKey"`
	PurchaseID uint    `json:"purchase_id"`
	Amount     float64 `json:"amount"`
	Method     string  `json:"method"`
	CashBankID *uint   `json:"cash_bank_id"`
	PaymentID  *uint   `json:"payment_id"`
}

type Account struct {
	ID      uint    `json:"id" gorm:"primaryKey"`
	Code    string  `json:"code"`
	Name    string  `json:"name"`
	Type    string  `json:"type"`
	Balance float64 `json:"balance"`
}

type Contact struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type CashBank struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Name      string `json:"name"`
	AccountID uint   `json:"account_id"`
}

func main() {
	// Database connection
	dsn := "accounting_user:accounting_password@tcp(localhost:3306)/accounting_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("🔍 PURCHASE BALANCE ANALYSIS")
	fmt.Println("==========================================")

	// 1. Get account balances for Hutang Usaha (Accounts Payable)
	var accountsPayable []Account
	err = db.Where("code LIKE ? OR name LIKE ? OR name LIKE ?", 
		"%2101%", "%Hutang Usaha%", "%Accounts Payable%").Find(&accountsPayable).Error
	if err != nil {
		log.Printf("Warning: Could not fetch accounts payable: %v", err)
	}

	fmt.Println("\n📊 ACCOUNTS PAYABLE BALANCES:")
	apTotalBalance := 0.0
	for _, acc := range accountsPayable {
		fmt.Printf("   %s - %s: Rp %.2f\n", acc.Code, acc.Name, acc.Balance)
		apTotalBalance += acc.Balance
	}
	fmt.Printf("   Total Accounts Payable Balance: Rp %.2f\n", apTotalBalance)

	// 2. Get bank account balances
	var bankAccounts []Account
	err = db.Where("code LIKE ? OR code LIKE ? OR code LIKE ? OR code LIKE ?", 
		"1101%", "1102%", "1103%", "1104%").Find(&bankAccounts).Error
	if err != nil {
		log.Printf("Warning: Could not fetch bank accounts: %v", err)
	}

	fmt.Println("\n🏦 BANK ACCOUNT BALANCES:")
	bankTotalBalance := 0.0
	for _, acc := range bankAccounts {
		fmt.Printf("   %s - %s: Rp %.2f\n", acc.Code, acc.Name, acc.Balance)
		bankTotalBalance += acc.Balance
	}
	fmt.Printf("   Total Bank Account Balance: Rp %.2f\n", bankTotalBalance)

	// 3. Get purchase summary
	var purchases []Purchase
	err = db.Find(&purchases).Error
	if err != nil {
		log.Fatal("Failed to fetch purchases:", err)
	}

	fmt.Println("\n📋 PURCHASE DATA SUMMARY:")
	fmt.Printf("   Total Purchases: %d\n", len(purchases))

	totalPurchaseAmount := 0.0
	totalPaidAmount := 0.0
	totalOutstanding := 0.0
	creditPurchases := 0
	paidPurchases := 0

	for _, purchase := range purchases {
		totalPurchaseAmount += purchase.TotalAmount
		totalPaidAmount += purchase.PaidAmount
		totalOutstanding += purchase.OutstandingAmount

		if purchase.PaymentMethod == "CREDIT" {
			creditPurchases++
		}
		if purchase.Status == "PAID" || purchase.OutstandingAmount == 0 {
			paidPurchases++
		}
	}

	fmt.Printf("   Total Purchase Amount: Rp %.2f\n", totalPurchaseAmount)
	fmt.Printf("   Total Paid Amount: Rp %.2f\n", totalPaidAmount)
	fmt.Printf("   Total Outstanding: Rp %.2f\n", totalOutstanding)
	fmt.Printf("   Credit Purchases: %d\n", creditPurchases)
	fmt.Printf("   Paid Purchases: %d\n", paidPurchases)

	// 4. Get purchase payment summary
	var purchasePayments []PurchasePayment
	err = db.Find(&purchasePayments).Error
	if err != nil {
		log.Printf("Warning: Could not fetch purchase payments: %v", err)
	}

	fmt.Println("\n💰 PURCHASE PAYMENTS SUMMARY:")
	fmt.Printf("   Total Purchase Payments: %d\n", len(purchasePayments))

	totalPaymentAmount := 0.0
	bankPaymentAmount := 0.0
	cashPaymentAmount := 0.0

	for _, payment := range purchasePayments {
		totalPaymentAmount += payment.Amount
		if payment.CashBankID != nil {
			bankPaymentAmount += payment.Amount
		} else {
			cashPaymentAmount += payment.Amount
		}
	}

	fmt.Printf("   Total Payment Amount: Rp %.2f\n", totalPaymentAmount)
	fmt.Printf("   Bank Payments: Rp %.2f\n", bankPaymentAmount)
	fmt.Printf("   Cash Payments: Rp %.2f\n", cashPaymentAmount)

	// 5. Balance validation analysis
	fmt.Println("\n🔍 BALANCE VALIDATION ANALYSIS:")
	
	// Calculate expected accounts payable balance
	// For purchase accounting:
	// Credit Purchase: Dr. Expense, Cr. Accounts Payable (creates liability)
	// Payment: Dr. Accounts Payable, Cr. Bank (reduces liability)
	// Expected AP Balance = Total Outstanding (should be NEGATIVE since it's a liability)
	expectedAPBalance := -totalOutstanding // Negative because it's a liability
	fmt.Printf("   Expected Accounts Payable Balance: Rp %.2f (negative = liability)\n", expectedAPBalance)
	fmt.Printf("   Actual Accounts Payable Balance: Rp %.2f\n", apTotalBalance)
	apDiscrepancy := apTotalBalance - expectedAPBalance
	fmt.Printf("   Accounts Payable Discrepancy: Rp %.2f\n", apDiscrepancy)

	// Calculate expected bank balance change from payments
	fmt.Printf("   Bank Payments Made: Rp %.2f (should reduce bank balance)\n", bankPaymentAmount)
	
	// 6. Issue identification
	fmt.Println("\n⚠️  POTENTIAL ISSUES:")
	issues := 0

	if apDiscrepancy > 1.0 || apDiscrepancy < -1.0 { // Allow for small rounding differences
		fmt.Printf("   ❌ Accounts Payable balance discrepancy: Rp %.2f\n", apDiscrepancy)
		issues++
	} else {
		fmt.Println("   ✅ Accounts Payable balance appears correct")
	}

	// Check for payment consistency
	paymentDiscrepancy := totalPaidAmount - totalPaymentAmount
	if paymentDiscrepancy > 1.0 || paymentDiscrepancy < -1.0 {
		fmt.Printf("   ❌ Purchase paid amounts vs actual payments discrepancy: Rp %.2f\n", paymentDiscrepancy)
		issues++
	} else {
		fmt.Println("   ✅ Purchase payment amounts appear consistent")
	}

	// Check for outstanding amount consistency
	calculatedOutstanding := totalPurchaseAmount - totalPaidAmount
	outstandingDiscrepancy := totalOutstanding - calculatedOutstanding
	if outstandingDiscrepancy > 1.0 || outstandingDiscrepancy < -1.0 {
		fmt.Printf("   ❌ Outstanding amount calculation discrepancy: Rp %.2f\n", outstandingDiscrepancy)
		fmt.Printf("      Calculated: Rp %.2f, Recorded: Rp %.2f\n", calculatedOutstanding, totalOutstanding)
		issues++
	} else {
		fmt.Println("   ✅ Outstanding amounts calculation appears correct")
	}

	// 7. Detailed purchase analysis (show purchases with issues)
	fmt.Println("\n📝 DETAILED PURCHASE ANALYSIS:")
	purchaseIssues := 0
	
	for _, purchase := range purchases {
		calculatedOutstanding := purchase.TotalAmount - purchase.PaidAmount
		if purchase.OutstandingAmount != calculatedOutstanding {
			fmt.Printf("   ⚠️  Purchase %s: Outstanding mismatch (Recorded: %.2f, Should be: %.2f)\n", 
				purchase.Code, purchase.OutstandingAmount, calculatedOutstanding)
			purchaseIssues++
		}
	}
	
	if purchaseIssues == 0 {
		fmt.Println("   ✅ All purchase outstanding amounts are consistent")
	} else {
		fmt.Printf("   ❌ Found %d purchases with outstanding amount issues\n", purchaseIssues)
	}

	// 8. Summary and recommendations
	fmt.Println("\n📊 SUMMARY:")
	if issues == 0 && purchaseIssues == 0 {
		fmt.Println("   ✅ ALL PURCHASE BALANCES APPEAR CORRECT")
		fmt.Println("   No action needed - accounting balances are synchronized")
	} else {
		fmt.Printf("   ❌ FOUND %d BALANCE ISSUES\n", issues + purchaseIssues)
		fmt.Println("\n🔧 RECOMMENDED ACTIONS:")
		
		if apDiscrepancy > 1.0 || apDiscrepancy < -1.0 {
			fmt.Println("   1. Run purchase balance correction script to fix Accounts Payable balance")
		}
		
		if paymentDiscrepancy > 1.0 || paymentDiscrepancy < -1.0 {
			fmt.Println("   2. Reconcile purchase payment records with actual payment transactions")
		}
		
		if purchaseIssues > 0 {
			fmt.Println("   3. Run outstanding amount recalculation for affected purchases")
		}
		
		fmt.Println("   4. Implement purchase balance validation triggers to prevent future discrepancies")
	}

	// 9. Show sample data for verification
	fmt.Println("\n🔍 SAMPLE DATA FOR VERIFICATION:")
	
	// Show first 5 credit purchases
	var creditPurchasesSample []Purchase
	db.Where("payment_method = ? AND outstanding_amount > 0", "CREDIT").Limit(5).Find(&creditPurchasesSample)
	
	if len(creditPurchasesSample) > 0 {
		fmt.Println("   Outstanding Credit Purchases (sample):")
		for _, p := range creditPurchasesSample {
			fmt.Printf("   - %s: Total=%.2f, Paid=%.2f, Outstanding=%.2f\n", 
				p.Code, p.TotalAmount, p.PaidAmount, p.OutstandingAmount)
		}
	}
	
	// Show recent payments
	var recentPayments []PurchasePayment
	db.Order("id DESC").Limit(5).Find(&recentPayments)
	
	if len(recentPayments) > 0 {
		fmt.Println("   Recent Purchase Payments (sample):")
		for _, p := range recentPayments {
			fmt.Printf("   - Payment ID %d: Amount=%.2f, Method=%s, Purchase=%d\n", 
				p.ID, p.Amount, p.Method, p.PurchaseID)
		}
	}

	fmt.Println("\n✅ Purchase balance analysis completed!")
	fmt.Println("Next step: Run purchase balance correction script if issues were found")
}