package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
)

func main() {
	// Connect to database
	db := database.ConnectDB()

	fmt.Println("=== FIXING ASSET CONSTRAINT FOR SOFT DELETE ===")
	
	// 1. Check current constraint
	fmt.Println("\n1. CHECKING CURRENT CONSTRAINT:")
	var constraints []struct {
		ConstraintName string `gorm:"column:conname"`
		ConstraintType string `gorm:"column:contype"`
	}
	
	db.Raw(`
		SELECT conname, contype 
		FROM pg_constraint 
		WHERE conrelid = 'assets'::regclass 
		AND contype = 'u'
		AND conname LIKE '%code%'
	`).Scan(&constraints)
	
	for _, c := range constraints {
		fmt.Printf("   - %s (Unique constraint)\n", c.ConstraintName)
	}
	
	// 2. Drop the old constraint
	fmt.Println("\n2. DROPPING OLD CONSTRAINT:")
	err := db.Exec("ALTER TABLE assets DROP CONSTRAINT IF EXISTS assets_code_key").Error
	if err != nil {
		log.Printf("Error dropping constraint: %v", err)
	} else {
		fmt.Println("   ✅ Dropped assets_code_key constraint")
	}
	
	// 3. Create partial unique index (supports soft delete)
	fmt.Println("\n3. CREATING PARTIAL UNIQUE INDEX:")
	err = db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_assets_code_active 
		ON assets (code) 
		WHERE deleted_at IS NULL
	`).Error
	
	if err != nil {
		log.Printf("Error creating partial index: %v", err)
	} else {
		fmt.Println("   ✅ Created partial unique index idx_assets_code_active")
	}
	
	// 4. Verify the changes
	fmt.Println("\n4. VERIFYING NEW INDEX:")
	var indexes []struct {
		IndexName string `gorm:"column:indexname"`
		IndexDef  string `gorm:"column:indexdef"`
	}
	
	db.Raw(`
		SELECT indexname, indexdef 
		FROM pg_indexes 
		WHERE tablename = 'assets' 
		AND indexname LIKE '%code%'
	`).Scan(&indexes)
	
	if len(indexes) > 0 {
		for _, idx := range indexes {
			fmt.Printf("   - %s\n", idx.IndexName)
			fmt.Printf("     %s\n", idx.IndexDef)
		}
	} else {
		fmt.Println("   No code-related indexes found")
	}
	
	fmt.Println("\n=== CONSTRAINT FIX COMPLETE ===")
	fmt.Println("✅ Asset codes now support soft delete properly!")
	fmt.Println("✅ You can reuse deleted asset codes without conflicts!")
}
