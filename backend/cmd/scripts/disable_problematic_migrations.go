package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	fmt.Println("🔧 MIGRATION FILE DISABLER")
	fmt.Println("===========================")
	
	// Get migration directory
	migrationDir := "../../migrations" // Relative to cmd/scripts
	if _, err := os.Stat(migrationDir); os.IsNotExist(err) {
		migrationDir = "../migrations" // Alternative path
		if _, err := os.Stat(migrationDir); os.IsNotExist(err) {
			migrationDir = "migrations" // Current directory
		}
	}
	
	// Check if migration directory exists
	if _, err := os.Stat(migrationDir); os.IsNotExist(err) {
		fmt.Printf("❌ Migration directory not found: %s\n", migrationDir)
		return
	}
	
	fmt.Printf("📁 Using migration directory: %s\n", migrationDir)
	
	// List of problematic migrations to disable
	problematicMigrations := []string{
		"021_install_purchase_balance_system.sql",
		"021_install_purchase_balance_validation.sql",
		"022_purchase_balance_validation_postgresql.sql",
		"023_purchase_balance_validation_go_compatible.sql",
		"024_purchase_balance_simple.sql",
		"025_purchase_balance_no_dollar_quotes.sql",
	}
	
	fmt.Println("\n🚫 Disabling problematic migration files...")
	
	disabledCount := 0
	for _, migrationName := range problematicMigrations {
		originalPath := filepath.Join(migrationDir, migrationName)
		disabledPath := filepath.Join(migrationDir, migrationName + ".disabled")
		
		// Check if original file exists
		if _, err := os.Stat(originalPath); os.IsNotExist(err) {
			fmt.Printf("   ℹ️  %s - not found (already disabled?)\n", migrationName)
			continue
		}
		
		// Check if disabled file already exists
		if _, err := os.Stat(disabledPath); err == nil {
			fmt.Printf("   ℹ️  %s - already disabled\n", migrationName)
			continue
		}
		
		// Rename file to disable it
		err := os.Rename(originalPath, disabledPath)
		if err != nil {
			fmt.Printf("   ❌ %s - failed to disable: %v\n", migrationName, err)
		} else {
			fmt.Printf("   ✅ %s - disabled successfully\n", migrationName)
			disabledCount++
		}
	}
	
	fmt.Printf("\n📊 Results:\n")
	fmt.Printf("   Disabled: %d files\n", disabledCount)
	fmt.Printf("   Total checked: %d files\n", len(problematicMigrations))
	
	if disabledCount > 0 {
		fmt.Println("\n💡 What this does:")
		fmt.Println("   • Renames .sql files to .sql.disabled")
		fmt.Println("   • Migration system will skip these files")
		fmt.Println("   • Working migration (026_purchase_balance_minimal.sql) stays active")
		fmt.Println("   • You can re-enable later by removing .disabled extension")
		
		fmt.Println("\n🚀 Next step:")
		fmt.Println("   Run your backend - it should start without migration errors!")
		
		fmt.Println("\n📝 To re-enable later (if needed):")
		fmt.Println("   Run: rename *.sql.disabled *.sql in migrations directory")
	} else {
		fmt.Println("\n✅ No files needed to be disabled")
	}
}