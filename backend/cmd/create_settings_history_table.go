package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Database connection
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Creating settings_history table...")

	// Create table
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS settings_history (
		id BIGSERIAL PRIMARY KEY,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP NULL DEFAULT NULL,
		
		-- Reference to settings
		settings_id BIGINT NOT NULL,
		
		-- Change tracking
		field VARCHAR(255) NOT NULL,
		old_value TEXT,
		new_value TEXT,
		action VARCHAR(50) DEFAULT 'UPDATE',
		
		-- User tracking
		changed_by BIGINT NOT NULL,
		
		-- Additional context
		ip_address VARCHAR(255),
		user_agent TEXT,
		reason TEXT,
		
		-- Foreign keys
		CONSTRAINT fk_settings_history_settings 
			FOREIGN KEY (settings_id) REFERENCES settings(id) ON DELETE CASCADE
	);
	`

	if err := db.Exec(createTableSQL).Error; err != nil {
		log.Fatal("Failed to create table:", err)
	}

	log.Println("✅ Table created successfully")

	// Create indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_settings_history_settings_id ON settings_history(settings_id);",
		"CREATE INDEX IF NOT EXISTS idx_settings_history_changed_by ON settings_history(changed_by);",
		"CREATE INDEX IF NOT EXISTS idx_settings_history_field ON settings_history(field);",
		"CREATE INDEX IF NOT EXISTS idx_settings_history_created_at ON settings_history(created_at);",
		"CREATE INDEX IF NOT EXISTS idx_settings_history_deleted_at ON settings_history(deleted_at);",
	}

	log.Println("Creating indexes...")
	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			log.Printf("⚠️  Failed to create index: %v", err)
		}
	}

	log.Println("✅ All indexes created successfully")

	// Verify table exists
	var exists bool
	err = db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'settings_history')").Scan(&exists).Error
	if err != nil {
		log.Fatal("Failed to verify table:", err)
	}

	if exists {
		log.Println("✅ Verification passed: settings_history table exists")
	} else {
		log.Println("❌ Verification failed: settings_history table does not exist")
	}

	fmt.Println("\n✅ Done! The settings_history table has been created successfully.")
}
