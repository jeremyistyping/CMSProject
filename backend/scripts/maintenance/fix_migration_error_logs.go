package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	fmt.Println("🔧 FIX MIGRATION ERROR LOGS")
	fmt.Println("===========================")
	fmt.Println()

	fmt.Println("🔍 Akan mengubah error logs menjadi info logs untuk:")
	fmt.Println("   - Duplicate migration record constraints (normal scenario)")
	fmt.Println("   - Record not found di migration checks (normal scenario)")
	fmt.Println()

	fmt.Print("Lanjutkan? (ketik 'ya' untuk konfirmasi): ")
	var confirm string
	fmt.Scanln(&confirm)
	
	if confirm != "ya" && confirm != "y" {
		fmt.Println("Fix dibatalkan.")
		return
	}

	// Files to fix
	filesToFix := []string{
		"database/sales_balance_fix_migration.go",
		"database/auto_fix_migration.go",
		"database/asset_category_migration.go",
		"database/fix_missing_columns_migration.go",
		"database/fix_remaining_issues_migration.go",
	}

	fixedCount := 0
	for _, filePath := range filesToFix {
		if fixMigrationErrorLogs(filePath) {
			fixedCount++
			fmt.Printf("   ✅ Fixed: %s\n", filePath)
		}
	}

	fmt.Printf("\n🎉 MIGRATION ERROR LOGS FIX COMPLETED!\n")
	fmt.Printf("✅ Fixed %d files\n", fixedCount)
	fmt.Printf("✅ Duplicate constraint errors will now show as info logs\n")
	fmt.Printf("✅ Error noise significantly reduced\n")
}

func fixMigrationErrorLogs(filePath string) bool {
	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("   ❌ Could not read %s: %v\n", filePath, err)
		return false
	}

	originalContent := string(content)
	modifiedContent := originalContent

	// Fix pattern 1: Duplicate key constraint errors
	duplicateKeyPattern := regexp.MustCompile(`log\.Printf\("❌ Failed to record (.+?) migration: %v", err\)`)
	if duplicateKeyPattern.MatchString(modifiedContent) {
		modifiedContent = duplicateKeyPattern.ReplaceAllStringFunc(modifiedContent, func(match string) string {
			// Extract migration name
			migrationName := duplicateKeyPattern.FindStringSubmatch(match)[1]
			
			return fmt.Sprintf(`// Check if this is just a duplicate key constraint (normal scenario)
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "uni_migration_records_migration_id") {
			log.Printf("ℹ️  %s migration record already exists (normal) - migration was successful")
		} else {
			log.Printf("❌ Failed to record %s migration: %%v", err)
		}`, migrationName, migrationName)
		})
	}

	// Fix pattern 2: Record not found errors that are normal
	recordNotFoundPattern := regexp.MustCompile(`(if err := db\.Where\(.+?\)\.First\(.+?\)\.Error; err != nil \{\s*log\.Printf\("❌.+?", err\)\s*return\s*\})`)
	if recordNotFoundPattern.MatchString(modifiedContent) {
		modifiedContent = recordNotFoundPattern.ReplaceAllStringFunc(modifiedContent, func(match string) string {
			return strings.ReplaceAll(match, `log.Printf("❌`, `// This is normal for first run
			if !strings.Contains(err.Error(), "record not found") {
				log.Printf("❌`)
		})
	}

	// Fix pattern 3: Generic migration failed messages that might be normal
	migrationFailedPattern := regexp.MustCompile(`log\.Printf\("❌ (.+?) migration failed: %v", err\)`)
	if migrationFailedPattern.MatchString(modifiedContent) {
		modifiedContent = migrationFailedPattern.ReplaceAllStringFunc(modifiedContent, func(match string) string {
			migrationDesc := migrationFailedPattern.FindStringSubmatch(match)[1]
			return fmt.Sprintf(`// Check if this is a normal scenario
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "already exists") {
			log.Printf("ℹ️  %s migration skipped (already applied)")
		} else {
			log.Printf("❌ %s migration failed: %%v", err)
		}`, migrationDesc, migrationDesc)
		})
	}

	// Only write if content changed
	if modifiedContent != originalContent {
		// Write back to file
		err = os.WriteFile(filePath, []byte(modifiedContent), 0644)
		if err != nil {
			fmt.Printf("   ❌ Could not write %s: %v\n", filePath, err)
			return false
		}
		return true
	}

	return false
}