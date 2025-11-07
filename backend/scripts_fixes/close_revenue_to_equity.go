package main

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	log.Println("🔧 Closing Revenue Accounts to Retained Earnings")
	log.Println("===============================================")

	// Connect to database
	dsn := "postgres://postgres:postgres@localhost/sistem_akuntans_test?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}

	log.Printf("Database connected successfully")

	// The problem: Revenue accounts (Rp 44.4M) should be closed to Retained Earnings
	// This represents income that increases equity

	log.Println("\n📋 CURRENT SITUATION ANALYSIS:")
	log.Println("Assets: Rp 171,600,000 (Cash 105M + Receivables 44.4M + Header accounts 22.2M)")
	log.Println("Liabilities: Rp 0")
	log.Println("Equity: Rp 210,000,000 (Capital 105M * 2 due to header account)")
	log.Println("Revenue: Rp 44,400,000 (Should be closed to equity)")
	log.Println("Difference: -38.4M (because revenue hasn't been closed to equity)")

	// 1. First, let's check and fix the double counting in accounts
	log.Println("\n1. FIXING DOUBLE COUNTING IN HEADER ACCOUNTS:")
	
	// The issue is that both header accounts (1000, 1200, 3000, 4000) AND detail accounts have balances
	// Header accounts should not have balances in a proper chart of accounts
	
	headerAccounts := []string{"1000", "1200", "3000", "4000"}
	for _, code := range headerAccounts {
		var currentBalance float64
		err := db.Table("accounts").
			Select("balance").
			Where("code = ?", code).
			Scan(&currentBalance).Error
		
		if err == nil && currentBalance != 0 {
			log.Printf("🔧 Clearing header account %s balance: %.2f", code, currentBalance)
			db.Exec("UPDATE accounts SET balance = 0 WHERE code = ?", code)
		}
	}

	// 2. Close revenue accounts to retained earnings
	log.Println("\n2. CLOSING REVENUE TO RETAINED EARNINGS:")
	
	var totalRevenue float64
	db.Table("accounts").
		Select("COALESCE(SUM(balance), 0)").
		Where("type = 'REVENUE' AND code NOT LIKE '%000'"). // Exclude header accounts
		Scan(&totalRevenue)
	
	log.Printf("📊 Total Revenue to close: %.2f", totalRevenue)
	
	if totalRevenue > 0 {
		// Check if retained earnings account exists
		var retainedCount int64
		db.Table("accounts").Where("code = '3201'").Count(&retainedCount)
		
		if retainedCount == 0 {
			log.Printf("❌ Retained earnings account (3201) not found")
			log.Printf("🔧 Creating retained earnings account...")
			
			createRetainedEarnings := `
			INSERT INTO accounts (code, name, type, balance, is_active, created_at, updated_at) 
			VALUES ('3201', 'Laba Ditahan', 'EQUITY', 0, true, NOW(), NOW())
			ON CONFLICT (code) DO NOTHING`
			
			db.Exec(createRetainedEarnings)
			log.Printf("✅ Created retained earnings account")
		}
		
		// Transfer revenue to retained earnings
		log.Printf("🔄 Transferring revenue %.2f to retained earnings", totalRevenue)
		
		// Add revenue to retained earnings
		updateResult := db.Exec("UPDATE accounts SET balance = balance + ? WHERE code = '3201'", totalRevenue)
		if updateResult.Error != nil {
			log.Printf("❌ Error updating retained earnings: %v", updateResult.Error)
		} else {
			log.Printf("✅ Added %.2f to retained earnings", totalRevenue)
		}
		
		// Zero out revenue accounts (period-end closing)
		zeroResult := db.Exec("UPDATE accounts SET balance = 0 WHERE type = 'REVENUE'")
		if zeroResult.Error != nil {
			log.Printf("❌ Error closing revenue accounts: %v", zeroResult.Error)
		} else {
			log.Printf("✅ Closed revenue accounts to zero")
		}
	}

	// 3. Final verification
	log.Println("\n3. FINAL BALANCE SHEET VERIFICATION:")
	
	var finalAssets float64
	db.Table("accounts").
		Select("COALESCE(SUM(balance), 0)").
		Where("type = 'ASSET'").
		Scan(&finalAssets)
	
	var finalLiabilities float64
	db.Table("accounts").
		Select("COALESCE(SUM(balance), 0)").
		Where("type = 'LIABILITY'").
		Scan(&finalLiabilities)
	
	var finalEquity float64
	db.Table("accounts").
		Select("COALESCE(SUM(balance), 0)").
		Where("type = 'EQUITY'").
		Scan(&finalEquity)
	
	var finalRevenue float64
	db.Table("accounts").
		Select("COALESCE(SUM(balance), 0)").
		Where("type = 'REVENUE'").
		Scan(&finalRevenue)
	
	log.Printf("📊 Final Assets: %.2f", finalAssets)
	log.Printf("📊 Final Liabilities: %.2f", finalLiabilities) 
	log.Printf("📊 Final Equity: %.2f", finalEquity)
	log.Printf("📊 Final Revenue: %.2f (should be 0)", finalRevenue)
	log.Printf("📊 Total Liabilities + Equity: %.2f", finalLiabilities + finalEquity)
	
	finalDifference := finalAssets - (finalLiabilities + finalEquity)
	log.Printf("📊 Balance Difference: %.2f", finalDifference)
	
	if finalDifference == 0 {
		log.Printf("🎉 Balance Sheet is now PERFECTLY BALANCED!")
	} else if finalDifference < 100 && finalDifference > -100 {
		log.Printf("✅ Balance Sheet is essentially balanced (difference < 100)")
	} else {
		log.Printf("⚠️  Balance difference: %.2f", finalDifference)
	}

	// 4. Show key cash accounts
	log.Println("\n4. CASH ACCOUNT SUMMARY:")
	var cashAccounts []map[string]interface{}
	db.Table("accounts").
		Select("code, name, balance").
		Where("(code LIKE '110%' OR code LIKE '11%' OR name ILIKE '%kas%') AND balance != 0").
		Find(&cashAccounts)
	
	for _, acc := range cashAccounts {
		log.Printf("💰 %s: %s = %.2f", acc["code"], acc["name"], acc["balance"])
	}

	log.Println("\n🎉 Fix completed!")
	log.Println("Your balance sheet should now show:")
	log.Printf("- Assets: %.2f (including Cash: 105M)", finalAssets)
	log.Printf("- Liabilities: %.2f", finalLiabilities)
	log.Printf("- Equity: %.2f (including retained earnings from revenue)", finalEquity)
}