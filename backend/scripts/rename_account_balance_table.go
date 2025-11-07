package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println("🔄 Renaming AccountBalance Table")
	fmt.Println("================================")

	_ = config.LoadConfig()
	db := database.ConnectDB()

	fmt.Println("✅ Database connected successfully")

	// Check if account_period_balances table exists
	var periodBalancesExists bool
	err := db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_name = 'account_period_balances' AND table_schema = 'public'
		)
	`).Scan(&periodBalancesExists).Error
	
	if err != nil {
		log.Fatalf("Failed to check account_period_balances: %v", err)
	}

	if !periodBalancesExists {
		fmt.Println("\n1. Creating account_period_balances table...")
		
		sql := `
			CREATE TABLE IF NOT EXISTS account_period_balances (
				id BIGSERIAL PRIMARY KEY,
				account_id BIGINT NOT NULL,
				period VARCHAR(7) NOT NULL,
				balance DECIMAL(20,2) DEFAULT 0,
				debit_total DECIMAL(20,2) DEFAULT 0,
				credit_total DECIMAL(20,2) DEFAULT 0,
				created_at TIMESTAMP WITH TIME ZONE,
				updated_at TIMESTAMP WITH TIME ZONE,
				deleted_at TIMESTAMP WITH TIME ZONE,
				CONSTRAINT fk_account_period_balances_account FOREIGN KEY (account_id) REFERENCES accounts(id)
			);
		`
		
		err := db.Exec(sql).Error
		if err != nil {
			log.Fatalf("Failed to create account_period_balances: %v", err)
		}
		
		fmt.Println("✅ Created account_period_balances table")
	} else {
		fmt.Println("ℹ️  account_period_balances table already exists")
	}
	
	fmt.Println("\n2. Adding indexes...")
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_account_period_balances_account_id ON account_period_balances(account_id)",
		"CREATE INDEX IF NOT EXISTS idx_account_period_balances_period ON account_period_balances(period)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_account_period_balances_account_period ON account_period_balances(account_id, period) WHERE deleted_at IS NULL",
	}

	for _, index := range indexes {
		if err := db.Exec(index).Error; err != nil {
			log.Printf("Warning: Failed to create index: %v", err)
		}
	}
	
	fmt.Println("✅ Added indexes to account_period_balances table")
	
	fmt.Println("\n3. Checking for existing data to migrate...")
	var oldTableExists bool
	err = db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_name = 'account_balances_records' AND table_schema = 'public'
		)
	`).Scan(&oldTableExists).Error
	
	if err != nil {
		log.Printf("Warning: Failed to check account_balances_records: %v", err)
	}
	
	if oldTableExists {
		// Copy data from old table to new table if old table exists
		fmt.Println("🔄 Migrating data from account_balances_records to account_period_balances...")
		
		sql := `
			INSERT INTO account_period_balances (account_id, period, balance, debit_total, credit_total, created_at, updated_at)
			SELECT account_id, period, balance, debit_total, credit_total, created_at, updated_at
			FROM account_balances_records
			ON CONFLICT (account_id, period) WHERE deleted_at IS NULL DO NOTHING
		`
		
		err := db.Exec(sql).Error
		if err != nil {
			log.Printf("Warning: Failed to migrate data: %v", err)
		} else {
			fmt.Println("✅ Data migration completed")
		}
	} else {
		fmt.Println("ℹ️  No old data to migrate")
	}
	
	fmt.Println("\n🎉 AccountBalance table rename operation completed!")
}