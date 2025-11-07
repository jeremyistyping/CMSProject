package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ColumnInfo struct {
	ColumnName    string `gorm:"column:column_name"`
	DataType      string `gorm:"column:data_type"`
	IsNullable    string `gorm:"column:is_nullable"`
	ColumnDefault string `gorm:"column:column_default"`
}

func main() {
	fmt.Printf("🔍 CHECKING DATABASE SCHEMA...\n\n")
	
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	fmt.Printf("🔗 Connecting to database...\n")
	gormDB, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB:", err)
	}
	defer sqlDB.Close()

	// Check unified_journal_lines table schema
	fmt.Printf("\n=== UNIFIED_JOURNAL_LINES TABLE SCHEMA ===\n")
	var ujlColumns []ColumnInfo
	
	ujlQuery := `
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns 
		WHERE table_name = 'unified_journal_lines' 
		ORDER BY ordinal_position
	`
	
	err = gormDB.Raw(ujlQuery).Scan(&ujlColumns).Error
	if err != nil {
		log.Fatal("Failed to get unified_journal_lines schema:", err)
	}

	if len(ujlColumns) == 0 {
		fmt.Printf("❌ unified_journal_lines table not found!\n")
	} else {
		fmt.Printf("📋 Columns in unified_journal_lines:\n\n")
		fmt.Printf("%-25s | %-15s | %-8s | %s\n", "Column Name", "Data Type", "Nullable", "Default")
		fmt.Printf("%-25s-+%-15s-+%-8s-+-%s\n", "-------------------------", "---------------", "--------", "--------")
		
		for _, col := range ujlColumns {
			fmt.Printf("%-25s | %-15s | %-8s | %s\n", 
				col.ColumnName, col.DataType, col.IsNullable, col.ColumnDefault)
		}
	}

	// Check unified_journal_ledger table schema
	fmt.Printf("\n=== UNIFIED_JOURNAL_LEDGER TABLE SCHEMA ===\n")
	var ledgerColumns []ColumnInfo
	
	ledgerQuery := `
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns 
		WHERE table_name = 'unified_journal_ledger' 
		ORDER BY ordinal_position
	`
	
	err = gormDB.Raw(ledgerQuery).Scan(&ledgerColumns).Error
	if err != nil {
		log.Printf("Error getting unified_journal_ledger schema: %v", err)
	} else {
		if len(ledgerColumns) == 0 {
			fmt.Printf("❌ unified_journal_ledger table not found!\n")
		} else {
			fmt.Printf("📋 Columns in unified_journal_ledger:\n\n")
			fmt.Printf("%-25s | %-15s | %-8s | %s\n", "Column Name", "Data Type", "Nullable", "Default")
			fmt.Printf("%-25s-+%-15s-+%-8s-+-%s\n", "-------------------------", "---------------", "--------", "--------")
			
			for _, col := range ledgerColumns {
				fmt.Printf("%-25s | %-15s | %-8s | %s\n", 
					col.ColumnName, col.DataType, col.IsNullable, col.ColumnDefault)
			}
		}
	}

	// Check accounts table schema
	fmt.Printf("\n=== ACCOUNTS TABLE SCHEMA ===\n")
	var accountColumns []ColumnInfo
	
	accountQuery := `
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns 
		WHERE table_name = 'accounts' 
		ORDER BY ordinal_position
	`
	
	err = gormDB.Raw(accountQuery).Scan(&accountColumns).Error
	if err != nil {
		log.Printf("Error getting accounts schema: %v", err)
	} else {
		if len(accountColumns) == 0 {
			fmt.Printf("❌ accounts table not found!\n")
		} else {
			fmt.Printf("📋 Columns in accounts:\n\n")
			fmt.Printf("%-25s | %-15s | %-8s | %s\n", "Column Name", "Data Type", "Nullable", "Default")
			fmt.Printf("%-25s-+%-15s-+%-8s-+-%s\n", "-------------------------", "---------------", "--------", "--------")
			
			for _, col := range accountColumns {
				fmt.Printf("%-25s | %-15s | %-8s | %s\n", 
					col.ColumnName, col.DataType, col.IsNullable, col.ColumnDefault)
			}
		}
	}

	// Sample data from unified_journal_lines
	fmt.Printf("\n=== SAMPLE DATA FROM UNIFIED_JOURNAL_LINES ===\n")
	var sampleCount int
	err = gormDB.Raw("SELECT COUNT(*) FROM unified_journal_lines").Scan(&sampleCount).Error
	if err != nil {
		log.Printf("Error counting unified_journal_lines: %v", err)
	} else {
		fmt.Printf("Total rows in unified_journal_lines: %d\n", sampleCount)
		
		if sampleCount > 0 {
			// Get first 5 rows to understand the structure
			rows, err := sqlDB.Query("SELECT * FROM unified_journal_lines LIMIT 5")
			if err != nil {
				log.Printf("Error getting sample data: %v", err)
			} else {
				defer rows.Close()
				
				columns, _ := rows.Columns()
				fmt.Printf("\nSample data (first row):\n")
				
				if rows.Next() {
					values := make([]interface{}, len(columns))
					valuePtrs := make([]interface{}, len(columns))
					for i := range values {
						valuePtrs[i] = &values[i]
					}
					
					rows.Scan(valuePtrs...)
					
					for i, col := range columns {
						val := values[i]
						if val == nil {
							val = "NULL"
						}
						fmt.Printf("  %s: %v\n", col, val)
					}
				}
			}
		}
	}
}