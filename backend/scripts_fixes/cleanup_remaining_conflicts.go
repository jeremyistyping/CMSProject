package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("🧹 Cleaning up remaining conflicting files...")

	if err := removeRemainingConflictFiles(); err != nil {
		log.Fatalf("❌ Failed to cleanup remaining conflicts: %v", err)
	}

	log.Println("✅ Remaining conflict cleanup completed!")
}

func removeRemainingConflictFiles() error {
	// Additional files that may cause conflicts
	conflictingFiles := []string{
		// Enhanced services that may conflict with standardized ones
		"services/enhanced_sales_journal_service.go",
		
		// Debug and analysis files that may have old logic
		"analyze_account_fixes.go",
		"debug_missing_revenue.go", 
		"debug_revenue_issue.go",
		"debug_frontend_coa.go",
		"deep_debug_frontend_mismatch.go",
		
		// Fix files that may have conflicting account handling
		"final_fix_4101.go",
		"fix_account_4101_balance.go",
		"fix_revenue_balance_sync.go",
		"fix_revenue_balance_corrected.go",
		"fix_sales_journal_integration.go",
		"fix_sales_journal_integration_corrected.go",
		"fix_unified_journal_sync.go",
		"fix_kas_duplicate_issue.go",
		
		// Test files that may use old logic
		"test_sales_journal_fix.go",
		"test_complete_status_validation.go",
		"test_status_validation.go",
		
		// Analysis and force files
		"force_ssot_balance.go",
		"analyze_sales_posting.go",
		
		// Account checking files that may have conflicts
		"check_sales_journal.go",
		"check_sales_structure.go",
		"check_frontend_endpoints.go",
		"check_frontend_ssot_integration.go",
		"check_duplicate_accounts_api.go",
		"check_specific_sale.go",
		
		// Demo files that may use old structure
		"demo_sales_journal_impact.go",
		"demo_sales_double_entry.go",
		
		// Comprehensive fixes that may conflict
		"comprehensive_account_fix.go",
		
		// Simple account checks that may be inconsistent
		"simple_account_check.go",
		
		// Entry manual fixes
		"fix_entry_1_manual.go",
		
		// Tools directory conflicts
		"tools/check_ppn_accounts.go",
		"tools/force_refresh_balances.go",
		"tools/correct_account_balances.go",
		"tools/diagnose_coa_balances.go",
		"tools/fix_tax_and_revenue_mapping.go",
		"tools/check_journal_creation.go",
		"tools/check_sales_journal_logic.go",
		"tools/simple_diagnose.go",
		"tools/final_verification.go",
		
		// Verify files that may have old logic
		"verify_deployment_fix.go",
		"verify_revenue_ppn_coa.go",
		
		// Scripts with potential conflicts
		"cmd/scripts/analyze_purchase_accounting.go",
		"cmd/scripts/analyze_sales_accounting.go",
		"cmd/scripts/analyze_sales_balance_verification.go",
		"cmd/scripts/manual_revenue_fix.go",
		"cmd/scripts/fix_revenue_allocation.go",
		"cmd/scripts/fix_double_balance_issue.go",
		"cmd/scripts/clean_duplicate_journals.go",
		"cmd/scripts/clean_extra_journals.go",
		"cmd/scripts/reprocess_existing_purchases_to_ssot.go",
		"cmd/scripts/verify_payment_accounting.go",
		"cmd/scripts/verify_complete_balance.go",
		"cmd/scripts/debug_ssot_balance_update.go",
		"cmd/scripts/check_purchase_status.go",
		"cmd/scripts/test_payment_comprehensive.go",
		"cmd/scripts/check_ssot_sales_implementation.go",
		"cmd/scripts/add_missing_purchase_accounts.go",
		"cmd/scripts/test_purchase_payment_integration.go",
		"cmd/scripts/test_payment_implementation.go",
		"cmd/scripts/test_sales_payment_flow.go",
		"cmd/scripts/test_confirm_sale_debug.go",
		"cmd/scripts/quick_fix_pl_route.go",
		"cmd/scripts/check_accounts_simple.go",
		
		// Debug scripts directory
		"debug_scripts/",
	}

	backupDir := "backup_remaining_conflicts"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %v", err)
	}

	removedCount := 0
	backedUpCount := 0

	for _, fileName := range conflictingFiles {
		filePath := filepath.Join(".", fileName)
		
		// Check if file or directory exists
		info, err := os.Stat(filePath)
		if os.IsNotExist(err) {
			continue
		}

		log.Printf("🔸 Processing: %s", fileName)

		if info.IsDir() {
			// Handle directory
			backupPath := filepath.Join(backupDir, filepath.Base(fileName))
			if err := copyDir(filePath, backupPath); err != nil {
				log.Printf("⚠️  Warning: Failed to backup directory %s: %v", fileName, err)
			} else {
				backedUpCount++
				log.Printf("📁 Backed up directory to: %s", backupPath)
			}

			// Remove the directory
			if err := os.RemoveAll(filePath); err != nil {
				log.Printf("⚠️  Warning: Failed to remove directory %s: %v", fileName, err)
			} else {
				removedCount++
				log.Printf("🗑️  Removed directory: %s", fileName)
			}
		} else {
			// Handle file
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
	}

	log.Printf("✅ Cleanup completed: %d items removed, %d items backed up", removedCount, backedUpCount)
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

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create the destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		return copyFile(path, destPath)
	})
}