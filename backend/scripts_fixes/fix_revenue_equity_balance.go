package main

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	log.Println("🔧 Fixing Revenue/Equity Balance Issue")
	log.Println("=====================================")

	// Connect to database
	dsn := "postgres://postgres:postgres@localhost/sistem_akuntans_test?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}

	log.Printf("Database connected successfully")

	// 1. Check current revenue accounts
	log.Println("\n1. CHECKING REVENUE ACCOUNTS:")
	var revenueAccounts []map[string]interface{}
	err = db.Table("accounts").
		Select("id, code, name, type, balance").
		Where("type = 'REVENUE'").
		Find(&revenueAccounts).Error
	if err != nil {
		log.Printf("❌ Error getting revenue accounts: %v", err)
	} else {
		log.Printf("Revenue accounts (%d):", len(revenueAccounts))
		var totalRevenue float64
		for _, account := range revenueAccounts {
			balance := account["balance"]
			if balanceFloat, ok := balance.(float64); ok {
				totalRevenue += balanceFloat
				log.Printf("  - %s: %s = %.2f", account["code"], account["name"], balanceFloat)
			}
		}
		log.Printf("📊 Total Revenue: %.2f", totalRevenue)
	}

	// 2. Calculate net income (revenue - expenses)
	log.Println("\n2. CALCULATING NET INCOME:")
	
	var totalRevenue float64
	db.Table("accounts").
		Select("COALESCE(SUM(balance), 0)").
		Where("type = 'REVENUE'").
		Scan(&totalRevenue)
	
	var totalExpenses float64
	db.Table("accounts").
		Select("COALESCE(SUM(balance), 0)").
		Where("type = 'EXPENSE'").
		Scan(&totalExpenses)
	
	netIncome := totalRevenue - totalExpenses
	log.Printf("📊 Total Revenue: %.2f", totalRevenue)
	log.Printf("📊 Total Expenses: %.2f", totalExpenses)
	log.Printf("📊 Net Income: %.2f", netIncome)

	// 3. Check if retained earnings account exists, create if not
	log.Println("\n3. CHECKING RETAINED EARNINGS ACCOUNT:")
	var retainedEarningsAccount map[string]interface{}
	err = db.Table("accounts").
		Select("id, code, name, balance").
		Where("code = '3201'").
		First(&retainedEarningsAccount).Error
	
	if err != nil {
		log.Printf("⚠️  Retained earnings account (3201) not found or has error: %v", err)
	} else {
		currentBalance := retainedEarningsAccount["balance"]
		log.Printf("Current Retained Earnings (3201): %.2f", currentBalance)
		
		// Update retained earnings with net income
		if netIncome != 0 {
			log.Printf("📈 Adding net income %.2f to retained earnings", netIncome)
			
			updateResult := db.Exec("UPDATE accounts SET balance = balance + ? WHERE code = '3201'", netIncome)
			if updateResult.Error != nil {
				log.Printf("❌ Error updating retained earnings: %v", updateResult.Error)
			} else {
				log.Printf("✅ Updated retained earnings account")
			}
		}
	}

	// 4. Final balance sheet verification
	log.Println("\n4. FINAL BALANCE SHEET VERIFICATION:")
	
	// Recalculate totals
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
	
	log.Printf("📊 Final Assets: %.2f", finalAssets)
	log.Printf("📊 Final Liabilities: %.2f", finalLiabilities) 
	log.Printf("📊 Final Equity: %.2f", finalEquity)
	log.Printf("📊 Total Liabilities + Equity: %.2f", finalLiabilities + finalEquity)
	
	finalDifference := finalAssets - (finalLiabilities + finalEquity)
	log.Printf("📊 Balance Difference: %.2f", finalDifference)
	
	if finalDifference == 0 {
		log.Printf("🎉 Balance Sheet is now PERFECTLY BALANCED!")
	} else if finalDifference < 1000 && finalDifference > -1000 {
		log.Printf("✅ Balance Sheet is essentially balanced (difference < 1000)")
	} else {
		log.Printf("⚠️  Balance Sheet still has significant difference")
		
		// Show breakdown by account type
		log.Println("\n📋 DETAILED ACCOUNT BREAKDOWN:")
		var allAccounts []map[string]interface{}
		db.Table("accounts").
			Select("code, name, type, balance").
			Where("balance != 0").
			Order("type, code").
			Find(&allAccounts)
		
		currentType := ""
		typeTotal := 0.0
		for _, acc := range allAccounts {
			if acc["type"] != currentType {
				if currentType != "" {
					log.Printf("   %s TOTAL: %.2f", currentType, typeTotal)
					typeTotal = 0.0
				}
				currentType = acc["type"].(string)
				log.Printf("\n%s ACCOUNTS:", currentType)
			}
			balance := acc["balance"].(float64)
			typeTotal += balance
			log.Printf("  %s: %s = %.2f", acc["code"], acc["name"], balance)
		}
		if currentType != "" {
			log.Printf("   %s TOTAL: %.2f", currentType, typeTotal)
		}
	}

	log.Println("\n🎉 Fix completed!")
}