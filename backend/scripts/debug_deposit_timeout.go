package main

import (
	"log"
	"time"
	
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

func main() {
	log.Println("🔍 Debug Deposit Timeout Issue")
	log.Println("================================")

	// Connect to database
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	log.Println("✅ Database connected successfully")

	// Check account types in database
	log.Println("\n📊 Step 1: Checking account types in database...")
	var accountTypes []string
	err := db.Model(&models.Account{}).Distinct("type").Pluck("type", &accountTypes).Error
	if err != nil {
		log.Printf("❌ Error getting account types: %v", err)
	} else {
		log.Printf("Found account types:")
		for _, aType := range accountTypes {
			log.Printf("  - %s", aType)
		}
	}

	// Check for revenue accounts
	log.Println("\n💰 Step 2: Checking for revenue accounts...")
	
	// Try various combinations
	testQueries := []struct {
		name  string
		query func(*gorm.DB) *gorm.DB
	}{
		{"Revenue (exact)", func(db *gorm.DB) *gorm.DB {
			return db.Where("type = ?", "Revenue")
		}},
		{"REVENUE (uppercase)", func(db *gorm.DB) *gorm.DB {
			return db.Where("type = ?", "REVENUE")
		}},
		{"revenue (lowercase)", func(db *gorm.DB) *gorm.DB {
			return db.Where("type = ?", "revenue")
		}},
		{"Type contains 'revenue' (case insensitive)", func(db *gorm.DB) *gorm.DB {
			return db.Where("type ILIKE ?", "%revenue%")
		}},
		{"Type contains 'income'", func(db *gorm.DB) *gorm.DB {
			return db.Where("type ILIKE ?", "%income%")
		}},
		{"Code 4900", func(db *gorm.DB) *gorm.DB {
			return db.Where("code = ?", "4900")
		}},
		{"Name contains 'Other Income'", func(db *gorm.DB) *gorm.DB {
			return db.Where("name ILIKE ?", "%Other Income%")
		}},
	}

	for _, test := range testQueries {
		var count int64
		var accounts []models.Account
		
		test.query(db.Model(&models.Account{})).Count(&count)
		if count > 0 {
			test.query(db.Model(&models.Account{})).Find(&accounts)
			log.Printf("  ✅ %s: Found %d accounts", test.name, count)
			for _, acc := range accounts {
				log.Printf("    - ID: %d, Code: %s, Name: %s, Type: %s", 
					acc.ID, acc.Code, acc.Name, acc.Type)
			}
		} else {
			log.Printf("  ❌ %s: No accounts found", test.name)
		}
	}

	// Check for expense accounts
	log.Println("\n💸 Step 3: Checking for expense accounts...")
	
	expenseQueries := []struct {
		name  string
		query func(*gorm.DB) *gorm.DB
	}{
		{"EXPENSE (uppercase)", func(db *gorm.DB) *gorm.DB {
			return db.Where("type = ?", "EXPENSE")
		}},
		{"Expense (exact)", func(db *gorm.DB) *gorm.DB {
			return db.Where("type = ?", "Expense")
		}},
		{"Type contains 'expense'", func(db *gorm.DB) *gorm.DB {
			return db.Where("type ILIKE ?", "%expense%")
		}},
		{"Code 5900", func(db *gorm.DB) *gorm.DB {
			return db.Where("code = ?", "5900")
		}},
	}

	for _, test := range expenseQueries {
		var count int64
		var accounts []models.Account
		
		test.query(db.Model(&models.Account{})).Count(&count)
		if count > 0 {
			test.query(db.Model(&models.Account{})).Find(&accounts)
			log.Printf("  ✅ %s: Found %d accounts", test.name, count)
			for _, acc := range accounts {
				log.Printf("    - ID: %d, Code: %s, Name: %s, Type: %s", 
					acc.ID, acc.Code, acc.Name, acc.Type)
			}
		} else {
			log.Printf("  ❌ %s: No accounts found", test.name)
		}
	}

	// Check for equity accounts
	log.Println("\n🏦 Step 4: Checking for equity accounts...")
	
	equityQueries := []struct {
		name  string
		query func(*gorm.DB) *gorm.DB
	}{
		{"EQUITY (uppercase)", func(db *gorm.DB) *gorm.DB {
			return db.Where("type = ?", "EQUITY")
		}},
		{"Equity (exact)", func(db *gorm.DB) *gorm.DB {
			return db.Where("type = ?", "Equity")
		}},
		{"Code 3101", func(db *gorm.DB) *gorm.DB {
			return db.Where("code = ?", "3101")
		}},
		{"Modal Pemilik", func(db *gorm.DB) *gorm.DB {
			return db.Where("name ILIKE ?", "%Modal Pemilik%")
		}},
	}

	for _, test := range equityQueries {
		var count int64
		var accounts []models.Account
		
		test.query(db.Model(&models.Account{})).Count(&count)
		if count > 0 {
			test.query(db.Model(&models.Account{})).Find(&accounts)
			log.Printf("  ✅ %s: Found %d accounts", test.name, count)
			for _, acc := range accounts {
				log.Printf("    - ID: %d, Code: %s, Name: %s, Type: %s", 
					acc.ID, acc.Code, acc.Name, acc.Type)
			}
		} else {
			log.Printf("  ❌ %s: No accounts found", test.name)
		}
	}

	// Test database performance
	log.Println("\n⚡ Step 5: Testing database performance...")
	start := time.Now()
	var totalAccounts int64
	db.Model(&models.Account{}).Count(&totalAccounts)
	duration := time.Since(start)
	
	log.Printf("  📊 Total accounts: %d", totalAccounts)
	log.Printf("  ⏱️ Query time: %v", duration)

	if duration > 5*time.Second {
		log.Printf("  ⚠️ Database query is slow (%v), this might cause timeouts", duration)
	} else {
		log.Printf("  ✅ Database performance looks good")
	}

	// Check cash-bank accounts
	log.Println("\n🏦 Step 6: Checking cash-bank accounts...")
	var cashBanks []models.CashBank
	err = db.Find(&cashBanks).Error
	if err != nil {
		log.Printf("❌ Error getting cash-bank accounts: %v", err)
	} else {
		log.Printf("  📊 Found %d cash-bank accounts:", len(cashBanks))
		for _, cb := range cashBanks {
			log.Printf("    - ID: %d, Name: %s, Type: %s, Balance: %.2f", 
				cb.ID, cb.Name, cb.Type, cb.Balance)
		}
	}

	// Recommendations
	log.Println("\n💡 Recommendations:")
	
	// Check if we have any revenue accounts
	var revenueCount int64
	db.Model(&models.Account{}).Where("type ILIKE ?", "%revenue%").Count(&revenueCount)
	if revenueCount == 0 {
		log.Println("  🔧 Need to create revenue accounts for deposits")
	}

	var expenseCount int64
	db.Model(&models.Account{}).Where("type ILIKE ?", "%expense%").Count(&expenseCount)
	if expenseCount == 0 {
		log.Println("  🔧 Need to create expense accounts for withdrawals")
	}

	var equityCount int64
	db.Model(&models.Account{}).Where("type ILIKE ?", "%equity%").Count(&equityCount)
	if equityCount == 0 {
		log.Println("  🔧 Need to create equity accounts for opening balances")
	}

	log.Println("\n✅ Debug analysis completed!")
}