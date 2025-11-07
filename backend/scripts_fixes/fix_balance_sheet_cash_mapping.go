package main

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	log.Println("🔧 Fixing Balance Sheet Cash Mapping Issue")
	log.Println("==========================================")

	// Connect to database
	dsn := "postgres://postgres:postgres@localhost/sistem_akuntans_test?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}

	log.Printf("Database connected successfully")

	// 1. Update account balances based on SSOT journal entries
	log.Println("\n1. UPDATING ACCOUNT BALANCES FROM SSOT:")

	updateQuery := `
	UPDATE accounts 
	SET balance = COALESCE((
		SELECT 
			CASE 
				WHEN accounts.type IN ('ASSET', 'EXPENSE') THEN 
					COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)
				ELSE 
					COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0)
			END
		FROM unified_journal_lines ujl 
		LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
		WHERE ujl.account_id = accounts.id 
		AND uje.status = 'POSTED'
	), 0)
	WHERE accounts.id IN (
		SELECT DISTINCT ujl.account_id 
		FROM unified_journal_lines ujl 
		LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
		WHERE uje.status = 'POSTED'
	)`

	result := db.Exec(updateQuery)
	if result.Error != nil {
		log.Printf("❌ Error updating account balances: %v", result.Error)
	} else {
		log.Printf("✅ Updated %d account balances from SSOT journal entries", result.RowsAffected)
	}

	// 2. Show updated cash account balances
	log.Println("\n2. UPDATED CASH ACCOUNT BALANCES:")
	var cashAccounts []map[string]interface{}
	err = db.Table("accounts").
		Select("id, code, name, balance").
		Where("code LIKE '110%' OR code LIKE '11%' OR name ILIKE '%kas%' OR name ILIKE '%bank%'").
		Order("code").
		Find(&cashAccounts).Error
	if err != nil {
		log.Printf("❌ Error getting cash accounts: %v", err)
	} else {
		log.Printf("Cash accounts (%d):", len(cashAccounts))
		var totalCash float64
		for _, account := range cashAccounts {
			balance := account["balance"]
			if balanceFloat, ok := balance.(float64); ok && balanceFloat != 0 {
				totalCash += balanceFloat
				log.Printf("  - %s: %s = %.2f", account["code"], account["name"], balanceFloat)
			}
		}
		log.Printf("📊 Total Cash: %.2f", totalCash)
	}

	// 3. Refresh materialized view if it exists
	log.Println("\n3. REFRESHING MATERIALIZED VIEWS:")
	refreshResult := db.Exec("REFRESH MATERIALIZED VIEW IF EXISTS account_balances")
	if refreshResult.Error != nil {
		log.Printf("⚠️  Warning: Could not refresh materialized view: %v", refreshResult.Error)
	} else {
		log.Printf("✅ Materialized view refreshed")
	}

	// 4. Verify the fix by showing current total assets and liabilities
	log.Println("\n4. BALANCE SHEET VERIFICATION:")
	
	// Calculate total assets
	var totalAssets float64
	db.Table("accounts").
		Select("COALESCE(SUM(balance), 0)").
		Where("type = 'ASSET'").
		Scan(&totalAssets)
	
	// Calculate total liabilities  
	var totalLiabilities float64
	db.Table("accounts").
		Select("COALESCE(SUM(balance), 0)").
		Where("type = 'LIABILITY'").
		Scan(&totalLiabilities)
	
	// Calculate total equity
	var totalEquity float64
	db.Table("accounts").
		Select("COALESCE(SUM(balance), 0)").
		Where("type = 'EQUITY'").
		Scan(&totalEquity)
	
	log.Printf("📊 Total Assets: %.2f", totalAssets)
	log.Printf("📊 Total Liabilities: %.2f", totalLiabilities) 
	log.Printf("📊 Total Equity: %.2f", totalEquity)
	log.Printf("📊 Total Liabilities + Equity: %.2f", totalLiabilities + totalEquity)
	
	difference := totalAssets - (totalLiabilities + totalEquity)
	if difference == 0 {
		log.Printf("✅ Balance Sheet is now BALANCED!")
	} else {
		log.Printf("⚠️  Balance difference: %.2f", difference)
		log.Printf("   (This might be due to missing revenue/expense accounts in equity)")
	}

	log.Println("\n🎉 Fix completed!")
	log.Println("Now try accessing your balance sheet again - the cash balances should show correctly.")
}