package main

import (
	"log"
	
	"app-sistem-akuntansi/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("⚡ Adding Performance Indexes...")
	
	// Load configuration
	cfg := config.LoadConfig()
	
	// Connect to database
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	
	log.Println("✅ Database connected successfully")
	
	// Create indexes one by one
	indexes := map[string]string{
		"Blacklisted Tokens - Token": "CREATE INDEX IF NOT EXISTS idx_blacklisted_tokens_token ON blacklisted_tokens(token)",
		"Blacklisted Tokens - Expires": "CREATE INDEX IF NOT EXISTS idx_blacklisted_tokens_expires_at ON blacklisted_tokens(expires_at)",
		"Notifications - User ID": "CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id)",
		"Notifications - Type": "CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type)",
		"Notifications - Created": "CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at)",
		"Notifications - User+Type": "CREATE INDEX IF NOT EXISTS idx_notifications_user_type ON notifications(user_id, type)",
		"Sales - Date": "CREATE INDEX IF NOT EXISTS idx_sales_date ON sales(date)",
		"Purchases - Date": "CREATE INDEX IF NOT EXISTS idx_purchases_date ON purchases(date)",
		"Audit Logs - Created": "CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at)",
	}
	
	successCount := 0
	for name, indexSQL := range indexes {
		log.Printf("Creating %s...", name)
		if err := db.Exec(indexSQL).Error; err != nil {
			log.Printf("  ⚠️ Warning: %v", err)
		} else {
			log.Printf("  ✅ Success")
			successCount++
		}
	}
	
	log.Printf("✅ Created %d/%d indexes successfully", successCount, len(indexes))
	
	// Test performance
	log.Println("🧪 Testing query performance...")
	
	// Test blacklisted tokens query
	var tokenCount int64
	if err := db.Raw("SELECT COUNT(*) FROM blacklisted_tokens WHERE expires_at > NOW()").Scan(&tokenCount).Error; err != nil {
		log.Printf("⚠️ Blacklisted tokens test failed: %v", err)
	} else {
		log.Printf("✅ Blacklisted tokens: %d active", tokenCount)
	}
	
	// Test notifications query
	var notificationCount int64
	if err := db.Raw("SELECT COUNT(*) FROM notifications WHERE created_at > NOW() - INTERVAL '7 days'").Scan(&notificationCount).Error; err != nil {
		log.Printf("⚠️ Notifications test failed: %v", err)
	} else {
		log.Printf("✅ Recent notifications: %d", notificationCount)
	}
	
	// Update stats
	log.Println("📊 Updating database statistics...")
	db.Exec("ANALYZE")
	
	log.Println("✅ Index creation completed!")
}
