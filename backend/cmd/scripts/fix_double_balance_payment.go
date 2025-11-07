package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type CashBank struct {
	ID        uint    `gorm:"primaryKey"`
	Name      string
	Balance   float64
	AccountID uint
}

type Account struct {
	ID      uint `gorm:"primaryKey"`
	Code    string
	Name    string
	Balance float64
}

type CashBankTransaction struct {
	ID           uint    `gorm:"primaryKey"`
	CashBankID   uint
	Amount       float64
	ReferenceType string
}

func main() {
	fmt.Println("🔧 Fix Cash Bank Double Balance Issue")
	fmt.Println("=====================================")

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
		fmt.Println("\n📋 Running in DRY RUN mode (no database connection)")
		showManualFixInstructions()
		return
	}

	fmt.Println("✅ Database connected successfully")

	// Step 1: Analyze current state
	fmt.Println("\n📊 STEP 1: Analyzing Current Balance State")
	analyzeBalanceState(db)

	// Step 2: Identify accounts with potential double balance
	fmt.Println("\n🔍 STEP 2: Identifying Accounts with Double Balance")
	doubledAccounts := identifyDoubledAccounts(db)

	if len(doubledAccounts) == 0 {
		fmt.Println("✅ No accounts with doubled balance detected")
		return
	}

	fmt.Printf("⚠️ Found %d accounts with potential doubled balance\n", len(doubledAccounts))

	// Step 3: Fix doubled balances
	fmt.Println("\n🔧 STEP 3: Fixing Doubled Balances")
	fixDoubledBalances(db, doubledAccounts)

	// Step 4: Sync GL accounts
	fmt.Println("\n🔄 STEP 4: Syncing GL Account Balances")
	syncGLAccountBalances(db)

	// Step 5: Verify fix
	fmt.Println("\n✅ STEP 5: Verifying Fix")
	analyzeBalanceState(db)

	fmt.Println("\n🎉 Fix completed successfully!")
}

func analyzeBalanceState(db *gorm.DB) {
	var cashBanks []CashBank
	db.Find(&cashBanks)

	fmt.Printf("Cash Bank Account Analysis:\n")
	fmt.Printf("%-6s %-25s %-15s %-15s %-10s\n", "ID", "NAME", "BALANCE", "CALCULATED", "STATUS")
	fmt.Println("==================================================================")

	for _, cb := range cashBanks {
		// Calculate expected balance from transactions
		var transactionSum float64
		db.Raw("SELECT COALESCE(SUM(amount), 0) FROM cash_bank_transactions WHERE cash_bank_id = ?", cb.ID).Scan(&transactionSum)

		status := "✅ OK"
		if cb.Balance != transactionSum {
			if cb.Balance == transactionSum*2 {
				status = "❌ DOUBLED"
			} else {
				status = "⚠️ MISMATCH"
			}
		}

		fmt.Printf("%-6d %-25s %-15.2f %-15.2f %-10s\n",
			cb.ID, cb.Name, cb.Balance, transactionSum, status)
	}
}

func identifyDoubledAccounts(db *gorm.DB) []CashBank {
	var doubledAccounts []CashBank
	
	// Find accounts where balance is exactly double the transaction sum
	query := `
		SELECT cb.id, cb.name, cb.balance, cb.account_id
		FROM cash_banks cb
		WHERE cb.balance = (
			SELECT COALESCE(SUM(cbt.amount), 0) * 2 
			FROM cash_bank_transactions cbt 
			WHERE cbt.cash_bank_id = cb.id
		)
		AND cb.balance > 0
	`
	
	db.Raw(query).Scan(&doubledAccounts)
	return doubledAccounts
}

func fixDoubledBalances(db *gorm.DB, doubledAccounts []CashBank) {
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("❌ Transaction rolled back due to panic: %v", r)
		}
	}()

	for _, account := range doubledAccounts {
		// Calculate correct balance (half of current balance)
		correctBalance := account.Balance / 2
		
		fmt.Printf("🔧 Fixing account %d (%s): %.2f -> %.2f\n", 
			account.ID, account.Name, account.Balance, correctBalance)

		// Update CashBank balance
		result := tx.Exec("UPDATE cash_banks SET balance = ?, updated_at = NOW() WHERE id = ?", 
			correctBalance, account.ID)
		
		if result.Error != nil {
			tx.Rollback()
			log.Fatalf("❌ Failed to update cash bank balance: %v", result.Error)
		}

		// Create corrective transaction record
		correctiveAmount := -correctBalance // Negative to show balance reduction
		tx.Exec(`
			INSERT INTO cash_bank_transactions 
			(cash_bank_id, reference_type, reference_id, amount, balance_after, transaction_date, notes, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, NOW(), ?, NOW(), NOW())
		`, account.ID, "ADJUSTMENT", 0, correctiveAmount, correctBalance, 
		   "Balance correction - Double counting fix")

		fmt.Printf("✅ Account %d balance corrected\n", account.ID)
	}

	if err := tx.Commit().Error; err != nil {
		log.Fatalf("❌ Failed to commit balance fix: %v", err)
	}

	fmt.Printf("✅ Fixed %d accounts with doubled balance\n", len(doubledAccounts))
}

func syncGLAccountBalances(db *gorm.DB) {
	// Sync GL account balances to match CashBank balances
	query := `
		UPDATE accounts a
		JOIN cash_banks cb ON a.id = cb.account_id
		SET a.balance = cb.balance, a.updated_at = NOW()
		WHERE a.balance != cb.balance
	`
	
	result := db.Exec(query)
	if result.Error != nil {
		log.Printf("❌ Failed to sync GL account balances: %v", result.Error)
		return
	}

	fmt.Printf("🔄 Synced %d GL account balances\n", result.RowsAffected)
}

func showManualFixInstructions() {
	fmt.Println("\n📋 MANUAL FIX INSTRUCTIONS (Run in MySQL/phpMyAdmin):")
	fmt.Println("=====================================================")

	fmt.Println("\n1. IDENTIFY DOUBLED BALANCES:")
	fmt.Println("```sql")
	fmt.Println(`SELECT 
    cb.id, cb.name, cb.balance as current_balance,
    COALESCE(SUM(cbt.amount), 0) as transaction_sum,
    COALESCE(SUM(cbt.amount), 0) * 2 as doubled_sum,
    CASE 
        WHEN cb.balance = COALESCE(SUM(cbt.amount), 0) * 2 THEN 'DOUBLED'
        ELSE 'OK'
    END as status
FROM cash_banks cb
LEFT JOIN cash_bank_transactions cbt ON cb.id = cbt.cash_bank_id
GROUP BY cb.id, cb.name, cb.balance
HAVING status = 'DOUBLED';`)
	fmt.Println("```")

	fmt.Println("\n2. FIX DOUBLED BALANCES:")
	fmt.Println("```sql")
	fmt.Println(`-- Fix CashBank balances
UPDATE cash_banks cb
SET balance = balance / 2, updated_at = NOW()
WHERE balance = (
    SELECT doubled_sum FROM (
        SELECT cb2.id, COALESCE(SUM(cbt.amount), 0) * 2 as doubled_sum
        FROM cash_banks cb2
        LEFT JOIN cash_bank_transactions cbt ON cb2.id = cbt.cash_bank_id
        GROUP BY cb2.id
    ) as calc WHERE calc.id = cb.id
) AND balance > 0;`)
	fmt.Println("```")

	fmt.Println("\n3. SYNC GL ACCOUNT BALANCES:")
	fmt.Println("```sql")
	fmt.Println(`-- Sync GL accounts with CashBank balances
UPDATE accounts a
JOIN cash_banks cb ON a.id = cb.account_id
SET a.balance = cb.balance, a.updated_at = NOW()
WHERE a.balance != cb.balance;`)
	fmt.Println("```")

	fmt.Println("\n4. VERIFY FIX:")
	fmt.Println("```sql")
	fmt.Println(`SELECT 
    cb.id, cb.name, cb.balance,
    a.code as gl_code, a.balance as gl_balance,
    CASE 
        WHEN cb.balance = a.balance THEN '✅ SYNCED'
        ELSE '❌ OUT_OF_SYNC'
    END as sync_status
FROM cash_banks cb
JOIN accounts a ON cb.account_id = a.id
ORDER BY cb.id;`)
	fmt.Println("```")

	fmt.Println("\n🚨 IMPORTANT: Run these queries in a transaction and backup data first!")
}