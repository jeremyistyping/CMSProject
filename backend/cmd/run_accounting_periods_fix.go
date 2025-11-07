package main

import (
	"log"
	"app-sistem-akuntansi/database"
)

func main() {
	log.Println("Running accounting_periods structure fix migration...")
	
	// Connect to database
	db := database.ConnectDB()
	
	// Run the fix
	if err := database.FixAccountingPeriodsStructure(db); err != nil {
		log.Fatalf("Failed to fix accounting_periods structure: %v", err)
	}
	
	// Verify the fix
	type Column struct {
		ColumnName string
		DataType   string
		IsNullable string
	}
	
	var columns []Column
	db.Raw(`
		SELECT column_name, data_type, is_nullable 
		FROM information_schema.columns 
		WHERE table_name = 'accounting_periods' 
		AND column_name IN ('year', 'month')
		ORDER BY column_name
	`).Scan(&columns)
	
	log.Println("\n=== Verification: year and month columns ===")
	for _, col := range columns {
		log.Printf("- %s (%s, nullable: %s)", col.ColumnName, col.DataType, col.IsNullable)
	}
	
	log.Println("\n✅ Migration completed successfully!")
}
