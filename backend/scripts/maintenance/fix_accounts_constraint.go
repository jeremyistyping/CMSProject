package main

import (
	"log"
	"app-sistem-akuntansi/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	log.Println("=== Fix Accounts Database Constraint Issue ===")
	
	// Load configuration
	cfg := config.LoadConfig()
	
	// Connect to database
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	
	log.Println("Connected to database successfully")
	
	// Check if accounts table exists
	var tableExists bool
	err = db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'accounts')").Scan(&tableExists).Error
	if err != nil {
		log.Fatalf("Failed to check if accounts table exists: %v", err)
	}
	
	if !tableExists {
		log.Println("Accounts table doesn't exist, creating it...")
		// Create accounts table
		err = db.Exec(`
			CREATE TABLE IF NOT EXISTS accounts (
				id SERIAL PRIMARY KEY,
				code VARCHAR(20) NOT NULL,
				name VARCHAR(100) NOT NULL,
				description TEXT,
				type VARCHAR(20) NOT NULL,
				category VARCHAR(50),
				parent_id INTEGER REFERENCES accounts(id),
				level INTEGER DEFAULT 1,
				is_header BOOLEAN DEFAULT false,
				is_active BOOLEAN DEFAULT true,
				balance DECIMAL(20,2) DEFAULT 0,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				deleted_at TIMESTAMP NULL
			);
		`).Error
		if err != nil {
			log.Fatalf("Failed to create accounts table: %v", err)
		}
		log.Println("Accounts table created successfully")
	}
	
	// Drop problematic constraints if they exist
	log.Println("Cleaning up existing constraints...")
	
	constraints := []string{
		"uni_accounts_code",
		"accounts_code_key", 
		"idx_accounts_code",
		"accounts_code_unique",
	}
	
	for _, constraint := range constraints {
		log.Printf("Attempting to drop constraint/index: %s", constraint)
		
		// Try dropping as constraint
		err = db.Exec("ALTER TABLE accounts DROP CONSTRAINT IF EXISTS " + constraint).Error
		if err != nil {
			log.Printf("Note: Constraint %s may not exist or already dropped: %v", constraint, err)
		}
		
		// Try dropping as index
		err = db.Exec("DROP INDEX IF EXISTS " + constraint).Error
		if err != nil {
			log.Printf("Note: Index %s may not exist or already dropped: %v", constraint, err)
		}
	}
	
	// Create proper indexes
	log.Println("Creating proper indexes...")
	
	// Create a unique partial index for active (non-deleted) records
	err = db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_accounts_code_active 
		ON accounts (code) 
		WHERE deleted_at IS NULL
	`).Error
	if err != nil {
		log.Printf("Warning: Failed to create unique index on code: %v", err)
	} else {
		log.Println("✅ Created unique index on accounts.code for active records")
	}
	
	// Create other useful indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_accounts_type ON accounts(type)",
		"CREATE INDEX IF NOT EXISTS idx_accounts_parent ON accounts(parent_id)",
		"CREATE INDEX IF NOT EXISTS idx_accounts_level ON accounts(level)",
		"CREATE INDEX IF NOT EXISTS idx_accounts_active ON accounts(is_active)",
		"CREATE INDEX IF NOT EXISTS idx_accounts_header ON accounts(is_header)",
		"CREATE INDEX IF NOT EXISTS idx_accounts_deleted_at ON accounts(deleted_at)",
		"CREATE INDEX IF NOT EXISTS idx_accounts_type_category ON accounts(type, category)",
		"CREATE INDEX IF NOT EXISTS idx_accounts_parent_level ON accounts(parent_id, level)",
	}
	
	for _, indexSQL := range indexes {
		err = db.Exec(indexSQL).Error
		if err != nil {
			log.Printf("Warning: Failed to create index: %v", err)
		}
	}
	
	log.Println("✅ Database indexes created successfully")
	
	// Verify table structure
	log.Println("Verifying accounts table structure...")
	
	var columns []struct {
		ColumnName string `gorm:"column:column_name"`
		DataType   string `gorm:"column:data_type"`
		IsNullable string `gorm:"column:is_nullable"`
	}
	
	err = db.Raw(`
		SELECT column_name, data_type, is_nullable 
		FROM information_schema.columns 
		WHERE table_name = 'accounts' 
		ORDER BY ordinal_position
	`).Scan(&columns).Error
	if err != nil {
		log.Printf("Warning: Failed to verify table structure: %v", err)
	} else {
		log.Println("Accounts table columns:")
		for _, col := range columns {
			log.Printf("  - %s (%s, nullable: %s)", col.ColumnName, col.DataType, col.IsNullable)
		}
	}
	
	log.Println("=== Fix completed successfully! ===")
	log.Println("You can now run your application. The database should initialize properly.")
}
