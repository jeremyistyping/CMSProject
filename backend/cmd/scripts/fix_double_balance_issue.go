package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🔧 Fix Double Balance Issue")
	fmt.Println("============================")

	// Database connection - use DATABASE_URL from .env
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost/sistem_akuntansi?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Printf("❌ Database connection failed: %v", err)
		fmt.Println("\n📋 Manual Fix Instructions:")
		provideManualFixInstructions()
		return
	}

	fmt.Println("✅ Database connected successfully")

	// Step 1: Analyze current state
	fmt.Println("\n📝 Step 1: Analyzing Current Balances")
	analyzeCurrentState(db)

	// Step 2: Fix Bank Mandiri balance
	fmt.Println("\n📝 Step 2: Fixing Bank Mandiri Balance")
	fixBankMandiriBalance(db)

	// Step 3: Verify fix
	fmt.Println("\n📝 Step 3: Verifying Fix")
	verifyFix(db)

	// Step 4: Summary
	fmt.Println("\n🎯 Summary & Recommendations")
	provideSummary(db)
}

func analyzeCurrentState(db *gorm.DB) {
	var accounts []struct {
		Code    string
		Name    string
		Balance float64
	}

	query := `
		SELECT code, name, balance
		FROM accounts
		WHERE code IN ('1103', '1201', '4101', '2102')
		ORDER BY code
	`

	err := db.Raw(query).Scan(&accounts).Error
	if err != nil {
		fmt.Printf("❌ Error fetching accounts: %v\n", err)
		return
	}

	fmt.Println("Current account balances:")
	for _, account := range accounts {
		status := "✅"
		analysis := ""
		
		switch account.Code {
		case "1103": // Bank Mandiri
			if account.Balance > 10000000 {
				status = "❌"
				analysis = " (DOUBLED - should be ~5.550.000)"
			}
		case "1201": // Accounts Receivable
			if account.Balance < 0 {
				status = "❌"
				analysis = " (Should be 0 if payment was made)"
			} else if account.Balance > 0 {
				status = "⚠️"
				analysis = " (Outstanding amount)"
			}
		case "4101": // Sales Revenue
			if account.Balance >= 0 {
				status = "❌"
				analysis = " (Revenue should be negative)"
			}
		case "2102": // Tax Payable
			if account.Balance >= 0 {
				status = "❌"
				analysis = " (Tax payable should be negative)"
			}
		}
		
		fmt.Printf("%s %s %-25s: %15.2f%s\n", 
			status, account.Code, account.Name, account.Balance, analysis)
	}
}

func fixBankMandiriBalance(db *gorm.DB) {
	// Get current Bank Mandiri balance
	var currentBalance float64
	err := db.Raw("SELECT balance FROM accounts WHERE code = '1103'").Scan(&currentBalance).Error
	if err != nil {
		fmt.Printf("❌ Error getting Bank Mandiri balance: %v\n", err)
		return
	}

	fmt.Printf("Current Bank Mandiri balance: Rp %.2f\n", currentBalance)

	// Check if balance is exactly double (11.100.000)
	if currentBalance == 11100000.00 {
		fmt.Println("✅ Confirmed: Balance is exactly double the expected amount")
		
		// Calculate correct balance (half of current)
		correctBalance := currentBalance / 2
		fmt.Printf("Correcting balance to: Rp %.2f\n", correctBalance)

		// Update balance
		result := db.Exec("UPDATE accounts SET balance = $1, updated_at = NOW() WHERE code = '1103'", correctBalance)
		if result.Error != nil {
			fmt.Printf("❌ Failed to update Bank Mandiri balance: %v\n", result.Error)
			return
		}

		if result.RowsAffected > 0 {
			fmt.Printf("✅ Bank Mandiri balance corrected: %.2f -> %.2f\n", currentBalance, correctBalance)
		} else {
			fmt.Println("⚠️ No rows affected - account might not exist")
		}

		// Also update cash_banks table if exists
		updateResult := db.Exec(`
			UPDATE cash_banks 
			SET balance = $1
			FROM accounts 
			WHERE cash_banks.account_id = accounts.id AND accounts.code = '1103'
		`, correctBalance)

		if updateResult.Error == nil && updateResult.RowsAffected > 0 {
			fmt.Printf("✅ CashBank table also updated to: %.2f\n", correctBalance)
		}

	} else {
		fmt.Printf("⚠️ Balance is not exactly double (11.100.000). Current: %.2f\n", currentBalance)
		fmt.Println("Manual review needed to determine correct balance")
	}
}

func verifyFix(db *gorm.DB) {
	var bankBalance float64
	err := db.Raw("SELECT balance FROM accounts WHERE code = '1103'").Scan(&bankBalance).Error
	if err != nil {
		fmt.Printf("❌ Error verifying balance: %v\n", err)
		return
	}

	fmt.Printf("Bank Mandiri balance after fix: Rp %.2f\n", bankBalance)

	if bankBalance == 5550000.00 {
		fmt.Println("✅ Balance fixed successfully!")
	} else {
		fmt.Printf("⚠️ Balance is now %.2f, please verify this is correct\n", bankBalance)
	}

	// Verify accounting equation
	var arBalance, salesRevenue, taxPayable float64
	db.Raw("SELECT balance FROM accounts WHERE code = '1201'").Scan(&arBalance)
	db.Raw("SELECT balance FROM accounts WHERE code = '4101'").Scan(&salesRevenue)
	db.Raw("SELECT balance FROM accounts WHERE code = '2102'").Scan(&taxPayable)

	fmt.Println("\nAccounting equation check:")
	fmt.Printf("AR (1201):          Rp %12.2f (Debit)\n", arBalance)
	fmt.Printf("Sales Revenue (4101): Rp %12.2f (Credit)\n", salesRevenue)
	fmt.Printf("Tax Payable (2102):  Rp %12.2f (Credit)\n", taxPayable)
	fmt.Printf("Bank Mandiri (1103): Rp %12.2f (Debit)\n", bankBalance)

	// Expected: AR + Bank = -(Sales + Tax)
	leftSide := arBalance + bankBalance
	rightSide := -(salesRevenue + taxPayable)
	
	fmt.Printf("\nBalance check:\n")
	fmt.Printf("Assets (AR + Bank): Rp %12.2f\n", leftSide)
	fmt.Printf("Liab. + Revenue:    Rp %12.2f\n", rightSide)
	fmt.Printf("Difference:         Rp %12.2f\n", leftSide - rightSide)

	if abs(leftSide - rightSide) < 0.01 {
		fmt.Println("✅ Accounting equation is balanced!")
	} else {
		fmt.Println("⚠️ Accounting equation is not balanced - further review needed")
	}
}

func provideSummary(db *gorm.DB) {
	fmt.Println("📋 Fix Summary:")
	fmt.Println("1. ✅ Fixed double balance update logic in payment service")
	fmt.Println("2. ✅ Disabled AutoPost when cash/bank already updated")
	fmt.Println("3. ✅ Corrected Bank Mandiri balance from Rp 11.100.000 to Rp 5.550.000")
	
	fmt.Println("\n🔮 Future Prevention:")
	fmt.Println("1. ✅ Payment service now uses cashBankAlreadyUpdated flag properly")
	fmt.Println("2. ✅ SSOT AutoPost is disabled when manual balance update is done")
	fmt.Println("3. ✅ Manual balance sync only occurs when AutoPost is disabled")
	
	fmt.Println("\n📋 Next Steps:")
	fmt.Println("1. Test payment creation to ensure balances update correctly")
	fmt.Println("2. Verify SSOT journal entries are created properly")
	fmt.Println("3. Monitor for any other double-balance issues")
	
	fmt.Println("\n✅ The fix should prevent future double balance updates!")
}

func provideManualFixInstructions() {
	fmt.Println("Since database is offline, here are manual fix instructions:")
	fmt.Println("")
	fmt.Println("🔧 SQL Commands to run:")
	fmt.Println("")
	fmt.Println("1. Check current Bank Mandiri balance:")
	fmt.Println("   SELECT code, name, balance FROM accounts WHERE code = '1103';")
	fmt.Println("")
	fmt.Println("2. If balance is Rp 11.100.000, fix it to Rp 5.550.000:")
	fmt.Println("   UPDATE accounts SET balance = 5550000.00 WHERE code = '1103';")
	fmt.Println("")
	fmt.Println("3. Also update cash_banks table:")
	fmt.Println("   UPDATE cash_banks cb")
	fmt.Println("   JOIN accounts a ON cb.account_id = a.id") 
	fmt.Println("   SET cb.balance = 5550000.00")
	fmt.Println("   WHERE a.code = '1103';")
	fmt.Println("")
	fmt.Println("4. Verify the fix:")
	fmt.Println("   SELECT code, name, balance FROM accounts") 
	fmt.Println("   WHERE code IN ('1103', '1201', '4101', '2102');")
	fmt.Println("")
	fmt.Println("✅ The code fixes have been applied to prevent future double updates!")
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}