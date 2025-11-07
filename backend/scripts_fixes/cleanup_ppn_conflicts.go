package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("🧹 Cleaning up conflicting PPN files...")

	if err := removeConflictingPPNFiles(); err != nil {
		log.Fatalf("❌ Failed to cleanup PPN conflicts: %v", err)
	}

	log.Println("✅ PPN conflict cleanup completed!")
}

func removeConflictingPPNFiles() error {
	// Files that potentially cause PPN conflicts
	conflictingFiles := []string{
		"fix_ppn_keluaran.go",
		"fix_ppn_keluaran_clean.go",
		"ppn_validation_service.go",
		"test_ppn_validation.go", 
		"verify_ppn_integration.go",
		"temp_check_ppn.go",
		"activate_ppn_protection.go",
		"check_ppn_mapping.go",
		"fix_ppn_hierarchy.go",
		"analyze_ppn_accounting.go",
		"fix_sales_double_entry_ppn.go",
		"test_ppn_fix_with_new_sale.go",
		"fix_account_balance_sync.go",
		"fix_balance_calculation.go",
		"tools/implement_ppn_fix.go",
		"tools/implement_ppn_fix_v2.go",
		"tools/test_ppn_fix_demo.go",
		"tools/complete_ppn_separation.go",
		"tools/verify_ppn_fix_future.go",
		"tools/fix_ppn_keluaran_name.go",
		"cmd/scripts/fix_sales_accounting_structure.go",
	}

	backupDir := "backup_ppn_conflicts"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %v", err)
	}

	removedCount := 0
	backedUpCount := 0

	for _, fileName := range conflictingFiles {
		filePath := filepath.Join(".", fileName)
		
		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			continue
		}

		log.Printf("🔸 Processing: %s", fileName)

		// Create backup first
		backupPath := filepath.Join(backupDir, filepath.Base(fileName))
		if err := copyFile(filePath, backupPath); err != nil {
			log.Printf("⚠️  Warning: Failed to backup %s: %v", fileName, err)
		} else {
			backedUpCount++
			log.Printf("📋 Backed up to: %s", backupPath)
		}

		// Remove the conflicting file
		if err := os.Remove(filePath); err != nil {
			log.Printf("⚠️  Warning: Failed to remove %s: %v", fileName, err)
		} else {
			removedCount++
			log.Printf("🗑️  Removed: %s", fileName)
		}
	}

	log.Printf("✅ Cleanup completed: %d files removed, %d files backed up", removedCount, backedUpCount)
	log.Printf("📁 Backup location: %s", backupDir)

	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy content
	_, err = destFile.ReadFrom(sourceFile)
	return err
}