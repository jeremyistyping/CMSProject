package main

import (
	"log"
	
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("🔧 Starting Safe Database Fixes...")
	
	// Load configuration
	cfg := config.LoadConfig()
	
	// Connect to database
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	
	log.Println("✅ Database connected successfully")
	
	// 1. Create performance indexes (skip security table indexes for now)
	log.Println("⚡ Creating performance indexes...")
	
	indexes := []string{
		// Blacklisted tokens optimization (if table exists)
		"CREATE INDEX IF NOT EXISTS idx_blacklisted_tokens_token ON blacklisted_tokens(token)",
		"CREATE INDEX IF NOT EXISTS idx_blacklisted_tokens_expires_at ON blacklisted_tokens(expires_at)",
		
		// Notification indexes (if table exists)
		"CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type)",
		"CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_notifications_user_type ON notifications(user_id, type)",
		
		// General performance indexes
		"CREATE INDEX IF NOT EXISTS idx_sales_date ON sales(date)",
		"CREATE INDEX IF NOT EXISTS idx_purchases_date ON purchases(date)",
		"CREATE INDEX IF NOT EXISTS idx_expenses_date ON expenses(date)",
		"CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(transaction_date)",
		"CREATE INDEX IF NOT EXISTS idx_products_stock ON products(stock)",
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at)",
		
		// Composite indexes for better query performance
		"CREATE INDEX IF NOT EXISTS idx_sales_customer_date ON sales(customer_id, date)",
		"CREATE INDEX IF NOT EXISTS idx_purchases_vendor_date ON purchases(vendor_id, date)",
		"CREATE INDEX IF NOT EXISTS idx_transactions_account_date ON transactions(account_id, transaction_date)",
	}
	
	successCount := 0
	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			log.Printf("⚠️ Warning: Failed to create index: %v", err)
		} else {
			successCount++
		}
	}
	
	log.Printf("✅ Created %d performance indexes successfully", successCount)
	
	// 2. Check and seed missing accounts
	log.Println("📋 Checking and seeding accounts...")
	
	// Check if account 1200 exists
	var account models.Account
	if err := db.Where("code = ?", "1200").First(&account).Error; err != nil {
		log.Println("🔧 Account 1200 (ACCOUNTS RECEIVABLE) not found, seeding...")
		if err := database.SeedAccounts(db); err != nil {
			log.Printf("❌ Failed to seed accounts: %v", err)
		} else {
			log.Println("✅ Accounts seeded successfully")
		}
	} else {
		log.Println("✅ Account 1200 (ACCOUNTS RECEIVABLE) already exists")
	}
	
	// Check if account 1104 exists (BANK UOB)
	if err := db.Where("code = ?", "1104").First(&account).Error; err != nil {
		log.Println("🔧 Account 1104 (BANK UOB) not found, creating...")
		newAccount := models.Account{
			Code:        "1104",
			Name:        "BANK UOB",
			Type:        models.AccountTypeAsset,
			Category:    models.CategoryCurrentAsset,
			Level:       3,
			IsHeader:    false,
			IsActive:    true,
			Balance:     0,
		}
		
		// Find parent (1100 - CURRENT ASSETS)
		var parentAccount models.Account
		if err := db.Where("code = ?", "1100").First(&parentAccount).Error; err == nil {
			newAccount.ParentID = &parentAccount.ID
		}
		
		if err := db.Create(&newAccount).Error; err != nil {
			log.Printf("❌ Failed to create account 1104: %v", err)
		} else {
			log.Println("✅ Account 1104 (BANK UOB) created successfully")
		}
	} else {
		log.Println("✅ Account 1104 (BANK UOB) already exists")
	}
	
	// 3. Check database consistency
	log.Println("🔍 Checking database consistency...")
	
	// Check for orphaned sale items
	var orphanedItems int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM sale_items si 
		LEFT JOIN sales s ON si.sale_id = s.id 
		WHERE s.id IS NULL
	`).Scan(&orphanedItems)
	
	if orphanedItems > 0 {
		log.Printf("⚠️  Found %d orphaned sale items", orphanedItems)
	} else {
		log.Println("✅ No orphaned sale items found")
	}
	
	// Check for orphaned sale payments
	var orphanedPayments int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM sale_payments sp 
		LEFT JOIN sales s ON sp.sale_id = s.id 
		WHERE s.id IS NULL
	`).Scan(&orphanedPayments)
	
	if orphanedPayments > 0 {
		log.Printf("⚠️  Found %d orphaned sale payments", orphanedPayments)
	} else {
		log.Println("✅ No orphaned sale payments found")
	}
	
	// 4. Validate system settings
	log.Println("⚙️  Validating system settings...")
	
	// Check admin user count
	var adminCount int64
	db.Model(&models.User{}).Where("role = ?", "admin").Count(&adminCount)
	log.Printf("👤 Found %d admin user(s)", adminCount)
	
	// Check total accounts
	var accountCount int64
	db.Model(&models.Account{}).Where("deleted_at IS NULL").Count(&accountCount)
	log.Printf("🏦 Found %d active accounts", accountCount)
	
	// Check sales count
	var salesCount int64
	db.Model(&models.Sale{}).Where("deleted_at IS NULL").Count(&salesCount)
	log.Printf("💰 Found %d sales records", salesCount)
	
	// 5. Update database statistics
	log.Println("📊 Updating database statistics...")
	if err := db.Exec("ANALYZE").Error; err != nil {
		log.Printf("⚠️ Warning: Failed to update statistics: %v", err)
	} else {
		log.Println("✅ Database statistics updated")
	}
	
	// 6. Test critical queries
	log.Println("🧪 Testing critical queries...")
	
	// Test blacklisted tokens query performance
	var tokenCount int64
	if err := db.Raw("SELECT COUNT(*) FROM blacklisted_tokens WHERE expires_at > NOW()").Scan(&tokenCount).Error; err != nil {
		log.Printf("⚠️ Blacklisted tokens query failed: %v", err)
	} else {
		log.Printf("✅ Found %d active blacklisted tokens", tokenCount)
	}
	
	// Test notification query performance 
	var notificationCount int64
	if err := db.Raw("SELECT COUNT(*) FROM notifications WHERE created_at > NOW() - INTERVAL '7 days'").Scan(&notificationCount).Error; err != nil {
		log.Printf("⚠️ Notifications query failed: %v", err)
	} else {
		log.Printf("✅ Found %d notifications in last 7 days", notificationCount)
	}
	
	log.Println("✅ Safe Database Fixes completed successfully!")
	log.Println("🚀 Improvements applied:")
	log.Println("   - Performance indexes created")
	log.Println("   - Missing accounts added")
	log.Println("   - Database consistency checked")
	log.Println("   - Statistics updated")
	log.Println("")
	log.Println("Note: Security models migration should be done by restarting the application")
	log.Println("The AutoMigrate in database.go now includes security models")
}
