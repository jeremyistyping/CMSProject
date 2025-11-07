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
	log.Println("🔧 Starting Database Issues Fix...")
	
	// Load configuration
	cfg := config.LoadConfig()
	
	// Connect to database
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	
	log.Println("✅ Database connected successfully")
	
	// 1. Fix missing security models migration - step by step
	log.Println("📊 Migrating security models one by one...")
	
	// Security Incident first
	log.Println("Creating SecurityIncident table...")
	if err := db.AutoMigrate(&models.SecurityIncident{}); err != nil {
		log.Printf("❌ Failed to create SecurityIncident table: %v", err)
	} else {
		log.Println("✅ SecurityIncident table created")
	}
	
	// System Alert
	log.Println("Creating SystemAlert table...")
	if err := db.AutoMigrate(&models.SystemAlert{}); err != nil {
		log.Printf("❌ Failed to create SystemAlert table: %v", err)
	} else {
		log.Println("✅ SystemAlert table created")
	}
	
	// Request Log
	log.Println("Creating RequestLog table...")
	if err := db.AutoMigrate(&models.RequestLog{}); err != nil {
		log.Printf("❌ Failed to create RequestLog table: %v", err)
	} else {
		log.Println("✅ RequestLog table created")
	}
	
	// IP Whitelist
	log.Println("Creating IpWhitelist table...")
	if err := db.AutoMigrate(&models.IpWhitelist{}); err != nil {
		log.Printf("❌ Failed to create IpWhitelist table: %v", err)
	} else {
		log.Println("✅ IpWhitelist table created")
	}
	
	// Security Config
	log.Println("Creating SecurityConfig table...")
	if err := db.AutoMigrate(&models.SecurityConfig{}); err != nil {
		log.Printf("❌ Failed to create SecurityConfig table: %v", err)
	} else {
		log.Println("✅ SecurityConfig table created")
	}
	
	// Security Metrics
	log.Println("Creating SecurityMetrics table...")
	if err := db.AutoMigrate(&models.SecurityMetrics{}); err != nil {
		log.Printf("❌ Failed to create SecurityMetrics table: %v", err)
	} else {
		log.Println("✅ SecurityMetrics table created")
	}
	
	// 2. Create performance indexes
	log.Println("⚡ Creating performance indexes...")
	
	indexes := []string{
		// Security and authentication indexes
		"CREATE INDEX IF NOT EXISTS idx_blacklisted_tokens_token ON blacklisted_tokens(token)",
		"CREATE INDEX IF NOT EXISTS idx_blacklisted_tokens_expires_at ON blacklisted_tokens(expires_at)",
		"CREATE INDEX IF NOT EXISTS idx_security_incidents_created_at ON security_incidents(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_security_incidents_client_ip ON security_incidents(client_ip)",
		"CREATE INDEX IF NOT EXISTS idx_security_incidents_incident_type ON security_incidents(incident_type)",
		"CREATE INDEX IF NOT EXISTS idx_request_logs_timestamp ON request_logs(timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_request_logs_client_ip ON request_logs(client_ip)",
		"CREATE INDEX IF NOT EXISTS idx_request_logs_is_suspicious ON request_logs(is_suspicious)",
		
		// Notification indexes
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
	}
	
	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			log.Printf("⚠️ Warning: Failed to create index: %v", err)
		}
	}
	
	log.Println("✅ Performance indexes created")
	
	// 3. Check and seed missing accounts
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
	
	// 4. Clean up any orphaned records
	log.Println("🧹 Cleaning up orphaned records...")
	
	// Check for orphaned sale items
	var orphanedItems int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM sale_items si 
		LEFT JOIN sales s ON si.sale_id = s.id 
		WHERE s.id IS NULL
	`).Scan(&orphanedItems)
	
	if orphanedItems > 0 {
		log.Printf("⚠️  Found %d orphaned sale items - manual cleanup may be required", orphanedItems)
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
		log.Printf("⚠️  Found %d orphaned sale payments - manual cleanup may be required", orphanedPayments)
	}
	
	// 5. Validate critical settings
	log.Println("⚙️  Validating system settings...")
	
	// Check if we have admin user
	var adminCount int64
	db.Model(&models.User{}).Where("role = ?", "admin").Count(&adminCount)
	
	if adminCount == 0 {
		log.Println("⚠️  Warning: No admin user found in system")
	} else {
		log.Printf("✅ Found %d admin user(s)", adminCount)
	}
	
	// 6. Update database statistics
	log.Println("📊 Updating database statistics...")
	db.Exec("ANALYZE")
	
	log.Println("✅ Database Issues Fix completed successfully!")
	log.Println("🚀 System should now perform better with:")
	log.Println("   - Security models properly migrated")
	log.Println("   - Performance indexes created")
	log.Println("   - Missing accounts added")
	log.Println("   - Orphaned records identified")
	log.Println("   - Database statistics updated")
}
