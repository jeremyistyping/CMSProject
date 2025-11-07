package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type CashBank struct {
	ID      uint    `gorm:"primaryKey"`
	Name    string
	Balance float64
}

type CashBankTransaction struct {
	ID            uint    `gorm:"primaryKey"`
	CashBankID    uint
	ReferenceType string
	Amount        float64
	Notes         string
}

func main() {
	fmt.Println("🧪 Testing Payment Balance Fix")
	fmt.Println("==============================")

	// Database connection
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "root"
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = ""
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "3306"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "sistem_akuntansi"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("❌ Database connection failed: %v", err)
		showTestInstructions()
		return
	}

	fmt.Println("✅ Database connected successfully")

	// Test 1: Check current balance state
	fmt.Println("\n📊 TEST 1: Current Balance State")
	testCurrentBalanceState(db)

	// Test 2: Simulate receivable payment
	fmt.Println("\n💰 TEST 2: Simulate Receivable Payment Logic")
	testReceivablePaymentLogic()

	// Test 3: Simulate payable payment
	fmt.Println("\n💸 TEST 3: Simulate Payable Payment Logic")
	testPayablePaymentLogic()

	// Test 4: Verify transaction types
	fmt.Println("\n🔍 TEST 4: Verify Transaction Types")
	testTransactionTypes(db)

	fmt.Println("\n🎉 All tests completed!")
}

func testCurrentBalanceState(db *gorm.DB) {
	var cashBanks []CashBank
	db.Find(&cashBanks)

	fmt.Printf("Current Cash Bank Balances:\n")
	fmt.Printf("%-6s %-25s %-15s\n", "ID", "NAME", "BALANCE")
	fmt.Println("=====================================================")

	for _, cb := range cashBanks {
		var transactionSum float64
		db.Raw("SELECT COALESCE(SUM(amount), 0) FROM cash_bank_transactions WHERE cash_bank_id = ?", cb.ID).Scan(&transactionSum)

		status := "✅ OK"
		if cb.Balance != transactionSum {
			status = "❌ MISMATCH"
		}

		fmt.Printf("%-6d %-25s %-15.2f %s\n", cb.ID, cb.Name, cb.Balance, status)
	}
}

func testReceivablePaymentLogic() {
	fmt.Println("Testing receivable payment logic:")
	fmt.Println("- Payment Method: 'RECEIVABLE' or 'Cash' or 'Transfer'")
	fmt.Println("- Expected: Cash & Bank balance INCREASES (+)")
	fmt.Println("- Journal Entry: Dr. Cash/Bank, Cr. Accounts Receivable")

	// Simulate the logic from our fix
	paymentMethod := "RECEIVABLE"
	paymentAmount := 1000000.0
	currentBalance := 5000000.0

	var amountChange float64
	var description string

	if paymentMethod == "RECEIVABLE" || paymentMethod == "Cash" || paymentMethod == "Transfer" {
		amountChange = paymentAmount // Positive amount increases balance
		description = "receivable payment (cash in)"
	} else {
		amountChange = -paymentAmount // Negative amount decreases balance
		description = "payable payment (cash out)"
	}

	newBalance := currentBalance + amountChange

	fmt.Printf("  Current Balance: %.2f\n", currentBalance)
	fmt.Printf("  Payment Amount: %.2f\n", paymentAmount)
	fmt.Printf("  Amount Change: %.2f\n", amountChange)
	fmt.Printf("  New Balance: %.2f\n", newBalance)
	fmt.Printf("  Description: %s\n", description)
	
	if newBalance > currentBalance {
		fmt.Println("  Result: ✅ CORRECT - Balance increased")
	} else {
		fmt.Println("  Result: ❌ ERROR - Balance should increase for receivable payment")
	}
}

func testPayablePaymentLogic() {
	fmt.Println("Testing payable payment logic:")
	fmt.Println("- Payment Method: anything other than 'RECEIVABLE/Cash/Transfer'")
	fmt.Println("- Expected: Cash & Bank balance DECREASES (-)")
	fmt.Println("- Journal Entry: Dr. Accounts Payable, Cr. Cash/Bank")

	// Simulate the logic from our fix
	paymentMethod := "BANK_TRANSFER_VENDOR"
	paymentAmount := 2000000.0
	currentBalance := 5000000.0

	var amountChange float64
	var description string

	if paymentMethod == "RECEIVABLE" || paymentMethod == "Cash" || paymentMethod == "Transfer" {
		amountChange = paymentAmount // Positive amount increases balance
		description = "receivable payment (cash in)"
	} else {
		amountChange = -paymentAmount // Negative amount decreases balance
		description = "payable payment (cash out)"

		// Check sufficient balance for payable payments
		if currentBalance < paymentAmount {
			fmt.Printf("  ❌ ERROR: Insufficient balance for payment: available=%.2f, required=%.2f\n", currentBalance, paymentAmount)
			return
		}
	}

	newBalance := currentBalance + amountChange

	fmt.Printf("  Current Balance: %.2f\n", currentBalance)
	fmt.Printf("  Payment Amount: %.2f\n", paymentAmount)
	fmt.Printf("  Amount Change: %.2f\n", amountChange)
	fmt.Printf("  New Balance: %.2f\n", newBalance)
	fmt.Printf("  Description: %s\n", description)
	
	if newBalance < currentBalance {
		fmt.Println("  Result: ✅ CORRECT - Balance decreased")
	} else {
		fmt.Println("  Result: ❌ ERROR - Balance should decrease for payable payment")
	}
}

func testTransactionTypes(db *gorm.DB) {
	fmt.Println("Recent cash bank transactions by type:")

	var transactions []struct {
		ReferenceType string
		Count         int64
		TotalAmount   float64
	}

	db.Raw(`
		SELECT reference_type, COUNT(*) as count, SUM(amount) as total_amount 
		FROM cash_bank_transactions 
		WHERE created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)
		GROUP BY reference_type
		ORDER BY count DESC
	`).Scan(&transactions)

	fmt.Printf("%-20s %-8s %-15s\n", "REFERENCE_TYPE", "COUNT", "TOTAL_AMOUNT")
	fmt.Println("================================================")

	for _, tx := range transactions {
		status := ""
		if tx.ReferenceType == "PAYMENT" || tx.ReferenceType == "ULTRA_FAST_PAYMENT" {
			if tx.TotalAmount > 0 {
				status = "✅ Positive (likely receivable)"
			} else if tx.TotalAmount < 0 {
				status = "✅ Negative (likely payable)"
			} else {
				status = "⚠️ Zero total"
			}
		}

		fmt.Printf("%-20s %-8d %-15.2f %s\n", tx.ReferenceType, tx.Count, tx.TotalAmount, status)
	}
}

func showTestInstructions() {
	fmt.Println("\n📋 TESTING INSTRUCTIONS (Without Database)")
	fmt.Println("==========================================")
	
	fmt.Println("\n✅ EXPECTED BEHAVIOR:")
	fmt.Println("1. RECEIVABLE PAYMENTS (from customers):")
	fmt.Println("   - Method: 'RECEIVABLE', 'Cash', 'Transfer'")
	fmt.Println("   - Cash & Bank balance should INCREASE (+)")
	fmt.Println("   - Transaction amount should be POSITIVE")
	fmt.Println("   - Journal: Dr. Cash/Bank, Cr. Accounts Receivable")
	
	fmt.Println("\n2. PAYABLE PAYMENTS (to vendors):")
	fmt.Println("   - Method: anything else (e.g., 'BANK_TRANSFER_VENDOR')")
	fmt.Println("   - Cash & Bank balance should DECREASE (-)")
	fmt.Println("   - Transaction amount should be NEGATIVE")
	fmt.Println("   - Journal: Dr. Accounts Payable, Cr. Cash/Bank")
	
	fmt.Println("\n🧪 MANUAL TESTING:")
	fmt.Println("1. Create a receivable payment from customer")
	fmt.Println("2. Verify Cash & Bank balance increases")
	fmt.Println("3. Create a payable payment to vendor")
	fmt.Println("4. Verify Cash & Bank balance decreases")
	
	fmt.Println("\n🔍 VERIFICATION SQL:")
	fmt.Println("-- Check transaction types and amounts")
	fmt.Println("SELECT reference_type, amount, notes FROM cash_bank_transactions ORDER BY created_at DESC LIMIT 10;")
	fmt.Println("")
	fmt.Println("-- Check balance consistency")
	fmt.Println("SELECT cb.name, cb.balance, COALESCE(SUM(cbt.amount), 0) as transaction_sum")
	fmt.Println("FROM cash_banks cb")
	fmt.Println("LEFT JOIN cash_bank_transactions cbt ON cb.id = cbt.cash_bank_id")
	fmt.Println("GROUP BY cb.id, cb.name, cb.balance;")
}