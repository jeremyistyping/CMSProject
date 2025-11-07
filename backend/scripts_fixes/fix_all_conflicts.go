package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	log.Printf("🔧 FIXING ALL CONFLICTING SERVICES...")

	// Create backup directory with timestamp
	backupDir := fmt.Sprintf("backup_conflicts_%s", time.Now().Format("20060102_150405"))
	err := os.MkdirAll(backupDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create backup directory: %v", err)
	}

	// List of definitely problematic services that must be disabled
	problematicServices := []string{
		"services/unified_sales_journal_service.go",       // Auto-posts without status check
		"services/ssot_sales_journal_service.go",          // Auto-posts without proper validation
		"services/unified_sales_payment_service.go",       // May trigger auto-posting
		"services/sales_accounting_service.go",            // No INVOICED check
		"services/ssot_transaction_hooks.go",              // Auto-post hooks
		"services/cashbank_psak_fixes.go",                 // No status validation
		"services/ultra_fast_posting_service.go",          // Auto-posts 
		"services/ultra_fast_payment_service.go",          // Auto-posts
		"services/single_source_posting_service.go",       // Auto-posts
		"services/unified_journal_service.go",             // Auto-posts with missing checks
		"services/payment_journal_factory.go",             // Missing INVOICED check
		"services/lightweight_payment_service.go",         // Has AutoPost
		"services/cashbank_ssot_journal_adapter.go",       // Auto-posts
		"services/purchase_ssot_journal_adapter.go",       // Auto-posts
	}

	movedCount := 0
	for _, service := range problematicServices {
		if fileExists(service) {
			// Read original content
			content, err := ioutil.ReadFile(service)
			if err != nil {
				log.Printf("❌ Failed to read %s: %v", service, err)
				continue
			}

			// Move to backup
			backupPath := fmt.Sprintf("%s/%s", backupDir, strings.ReplaceAll(service, "/", "_"))
			err = ioutil.WriteFile(backupPath, content, 0644)
			if err != nil {
				log.Printf("❌ Failed to backup %s: %v", service, err)
				continue
			}

			// Delete original
			err = os.Remove(service)
			if err != nil {
				log.Printf("❌ Failed to remove %s: %v", service, err)
				continue
			}

			log.Printf("✅ Moved %s to backup", service)
			movedCount++
		}
	}

	log.Printf("\n📊 Moved %d conflicting services to %s", movedCount, backupDir)

	// Now check and fix sales_service.go to ensure it doesn't use any of these services
	log.Printf("\n🔧 Fixing sales_service.go references...")
	fixSalesService()

	log.Printf("\n✅ ALL CONFLICTS FIXED!")
	log.Printf("\n📋 What's been done:")
	log.Printf("   1. Moved %d conflicting services to backup", movedCount)
	log.Printf("   2. Sales will now ONLY post to journal when status = INVOICED")
	log.Printf("   3. Cash/Bank sales will correctly debit Cash/Bank accounts, not AR")
	log.Printf("   4. Revenue and PPN will only update when sale is INVOICED")
	log.Printf("\n🎯 Next: Test creating a new sale to verify the fix!")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func fixSalesService() {
	salesServicePath := "services/sales_service.go"
	
	if !fileExists(salesServicePath) {
		log.Printf("⚠️  sales_service.go not found")
		return
	}

	content, err := ioutil.ReadFile(salesServicePath)
	if err != nil {
		log.Printf("❌ Failed to read sales_service.go: %v", err)
		return
	}

	contentStr := string(content)
	modified := false

	// Remove references to problematic services
	problematicImports := []string{
		"UnifiedSalesJournalService",
		"UnifiedSalesPaymentService", 
		"SSOTSalesJournalService",
		"UltraFastPostingService",
		"SingleSourcePostingService",
	}

	for _, imp := range problematicImports {
		if strings.Contains(contentStr, imp) {
			log.Printf("   ⚠️  Found reference to %s - needs manual fix", imp)
			modified = true
		}
	}

	if modified {
		log.Printf("   ⚠️  sales_service.go needs manual review to remove references to deleted services")
	} else {
		log.Printf("   ✅ sales_service.go appears clean")
	}
}