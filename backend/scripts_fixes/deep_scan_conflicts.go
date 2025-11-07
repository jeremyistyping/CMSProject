package main

import (
	"fmt"
	"log"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func main() {
	log.Printf("🔍 DEEP SCAN: Finding ALL services that might auto-post DRAFT sales...")

	// Directories to scan
	dirs := []string{
		"services",
		"handlers", 
		"controllers",
		"api",
	}

	// Patterns that indicate auto-posting for DRAFT/creation
	suspiciousPatterns := []string{
		// Direct journal creation on sale creation
		"CreateSale.*CreateJournal",
		"CreateSale.*journal",
		"sale.Status == models.SaleStatusDraft",
		"Status.*Draft.*journal",
		"DRAFT.*CreateEntry",
		
		// Auto-posting patterns
		"AutoPost",
		"autoPost", 
		"auto_post",
		"automatic.*entry",
		"immediate.*journal",
		
		// Services that might bypass status check
		"UnifiedSales",
		"SSOTSales", 
		"DoubleEntry.*Create",
		"CreateEntry.*sale",
		
		// Balance update on creation
		"CreateSale.*UpdateBalance",
		"CreateSale.*balance",
		"sale.*balance.*update",
		
		// Wrong AR posting for cash sales
		"CASH.*Receivable",
		"cash.*1201",
		"PaymentMethodType.*ignored",
	}

	conflictingFiles := make(map[string][]string)

	for _, dir := range dirs {
		files, err := filepath.Glob(fmt.Sprintf("%s/*.go", dir))
		if err != nil {
			continue
		}

		for _, file := range files {
			content, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}

			contentStr := string(content)
			var matches []string

			for _, pattern := range suspiciousPatterns {
				if strings.Contains(strings.ToLower(contentStr), strings.ToLower(pattern)) {
					matches = append(matches, pattern)
				}
			}

			// Special checks for problematic code
			if strings.Contains(contentStr, "CreateSale") {
				if strings.Contains(contentStr, "CreateJournalEntry") ||
				   strings.Contains(contentStr, "CreateEntry") ||
				   strings.Contains(contentStr, "PostEntry") ||
				   strings.Contains(contentStr, "UpdateBalance") {
					matches = append(matches, "CreateSale with auto-posting")
				}
			}

			// Check for missing status validation
			if strings.Contains(contentStr, "CreateSaleJournalEntry") ||
			   strings.Contains(contentStr, "CreateJournalEntry") {
				if !strings.Contains(contentStr, "SaleStatusInvoiced") &&
				   !strings.Contains(contentStr, "status != models.SaleStatusInvoiced") {
					matches = append(matches, "Missing INVOICED status check")
				}
			}

			// Check for wrong payment method handling
			if strings.Contains(contentStr, "PaymentMethodType") {
				if strings.Contains(contentStr, "1201") || // AR account
				   strings.Contains(contentStr, "AccountsReceivable") {
					if !strings.Contains(contentStr, "CREDIT") {
						matches = append(matches, "Wrong AR usage for non-credit sales")
					}
				}
			}

			if len(matches) > 0 {
				conflictingFiles[file] = matches
			}
		}
	}

	if len(conflictingFiles) == 0 {
		log.Printf("✅ No obvious conflicting files found")
	} else {
		log.Printf("❌ Found %d potentially conflicting files:", len(conflictingFiles))
		for file, matches := range conflictingFiles {
			log.Printf("\n📄 %s", file)
			for _, match := range matches {
				log.Printf("   - Matched pattern: %s", match)
			}
		}
	}

	// Check for specific problematic services
	log.Printf("\n🔍 Checking for specific problematic services...")
	problematicServices := []string{
		"services/unified_sales_payment_service.go",
		"services/sales_accounting_service.go", 
		"services/ssot_transaction_hooks.go",
		"services/auto_posting_service.go",
		"services/cashbank_psak_fixes.go",
	}

	for _, service := range problematicServices {
		if fileExists(service) {
			log.Printf("⚠️  Found potentially problematic service: %s", service)
			checkServiceContent(service)
		}
	}

	log.Printf("\n✅ Deep scan completed!")
}

func fileExists(path string) bool {
	if _, err := ioutil.ReadFile(path); err == nil {
		return true
	}
	return false
}

func checkServiceContent(path string) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	contentStr := string(content)
	
	// Check for problematic patterns
	if strings.Contains(contentStr, "CreateSale") && 
	   (strings.Contains(contentStr, "UpdateBalance") || 
	    strings.Contains(contentStr, "CreateJournal")) {
		log.Printf("   ❌ This service auto-posts on CreateSale!")
	}

	if strings.Contains(contentStr, "SaleStatusDraft") && 
	   strings.Contains(contentStr, "CreateEntry") {
		log.Printf("   ❌ This service allows DRAFT posting!")
	}

	if !strings.Contains(contentStr, "SaleStatusInvoiced") {
		log.Printf("   ⚠️  This service doesn't check for INVOICED status")
	}
}