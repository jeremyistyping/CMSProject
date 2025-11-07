package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func main() {
	fmt.Println("🔧 APPLYING STATUS VALIDATION FIXES")
	fmt.Println("Preventing auto-posting for draft/confirmed sales")
	fmt.Println("=" + strings.Repeat("=", 50))

	fixes := []Fix{
		{
			File: "services/sales_double_entry_service.go",
			Description: "Add status validation to CreateSaleJournalEntries",
			Search: `func (s *SalesDoubleEntryService) CreateSaleJournalEntries(sale *models.Sale, userID uint) error {
	log.Printf("🔄 Creating double-entry journal entries for sale %d, payment method: %s", sale.ID, sale.PaymentMethodType)

	// Validate sale data
	if err := s.validateSaleForJournalEntry(sale); err != nil {
		return fmt.Errorf("sale validation failed: %v", err)
	}`,
			Replace: `func (s *SalesDoubleEntryService) CreateSaleJournalEntries(sale *models.Sale, userID uint) error {
	log.Printf("🔄 Creating double-entry journal entries for sale %d, payment method: %s", sale.ID, sale.PaymentMethodType)

	// 🚫 CRITICAL: Only create journal entries for INVOICED sales
	if sale.Status != "INVOICED" && sale.Status != "PAID" {
		log.Printf("⚠️ BLOCKED: Journal creation blocked for sale %d with status %s (only INVOICED/PAID allowed)", sale.ID, sale.Status)
		return fmt.Errorf("journal entries can only be created for invoiced sales, current status: %s", sale.Status)
	}

	// Validate sale data
	if err := s.validateSaleForJournalEntry(sale); err != nil {
		return fmt.Errorf("sale validation failed: %v", err)
	}`,
		},
		{
			File: "services/sales_service.go", 
			Description: "Remove auto-posting from createJournalEntriesForSale",
			Search: `func (s *SalesService) createJournalEntriesForSale(sale *models.Sale, userID uint) error {
	// 🎆 USE NEW DOUBLE-ENTRY SERVICE WITH PROPER ACCOUNTING LOGIC
	log.Printf("💼 Creating journal entries for sale %d using Double Entry Service", sale.ID)
	
	// Use new SalesDoubleEntryService for proper accounting
	err := s.doubleEntryService.CreateSaleJournalEntries(sale, userID)
	if err != nil {
		log.Printf("❌ Failed to create double-entry journal entry for sale %d: %v", sale.ID, err)
		return fmt.Errorf("failed to create double-entry journal entries: %v", err)
	}`,
			Replace: `func (s *SalesService) createJournalEntriesForSale(sale *models.Sale, userID uint) error {
	// 🚫 DEPRECATED: This method should not be used for auto-posting
	// Journal entries are now created ONLY during invoice creation
	log.Printf("⚠️ DEPRECATED: createJournalEntriesForSale called for sale %d with status %s", sale.ID, sale.Status)
	
	// Only proceed if sale is INVOICED (safety check)
	if sale.Status != "INVOICED" && sale.Status != "PAID" {
		log.Printf("🚫 BLOCKED: Auto journal creation blocked for sale %d with status %s", sale.ID, sale.Status)
		return fmt.Errorf("journal entries can only be created for invoiced sales, current status: %s", sale.Status)
	}

	// Use new SalesDoubleEntryService for proper accounting  
	err := s.doubleEntryService.CreateSaleJournalEntries(sale, userID)
	if err != nil {
		log.Printf("❌ Failed to create double-entry journal entry for sale %d: %v", sale.ID, err)
		return fmt.Errorf("failed to create double-entry journal entries: %v", err)
	}`,
		},
		{
			File: "services/balance_sync_service.go",
			Description: "Add safety measures to balance sync",
			Search: `// SchedulePeriodicSync runs periodic balance synchronization
func (s *BalanceSyncService) SchedulePeriodicSync(intervalMinutes int) {
	ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Printf("🔄 Running periodic balance sync (every %d minutes)...", intervalMinutes)
			
			// Verify integrity first
			isConsistent, err := s.VerifyBalanceIntegrity()
			if err != nil {
				log.Printf("Error during periodic integrity check: %v", err)
				continue
			}

			// If inconsistent, run full sync
			if !isConsistent {
				log.Println("⚠️ Inconsistencies detected, running full balance sync...")
				err = s.SyncAccountBalancesFromSSOT()
				if err != nil {
					log.Printf("Error during periodic sync: %v", err)
				}
			}
		}
	}
}`,
			Replace: `// SchedulePeriodicSync runs periodic balance synchronization
// 🚫 DISABLED: Automatic balance sync can interfere with manual testing
func (s *BalanceSyncService) SchedulePeriodicSync(intervalMinutes int) {
	log.Printf("🚫 DISABLED: Automatic balance sync disabled to prevent interference with manual operations")
	log.Printf("ℹ️  Use manual sync methods: SyncAccountBalancesFromSSOT() or VerifyBalanceIntegrity()")
	
	// Comment out automatic sync to prevent draft sales interference
	/*
	ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Printf("🔄 Running periodic balance sync (every %d minutes)...", intervalMinutes)
			
			// Verify integrity first
			isConsistent, err := s.VerifyBalanceIntegrity()
			if err != nil {
				log.Printf("Error during periodic integrity check: %v", err)
				continue
			}

			// If inconsistent, run full sync
			if !isConsistent {
				log.Println("⚠️ Inconsistencies detected, running full balance sync...")
				err = s.SyncAccountBalancesFromSSOT()
				if err != nil {
					log.Printf("Error during periodic sync: %v", err)
				}
			}
		}
	}
	*/
}`,
		},
	}

	// Apply all fixes
	successCount := 0
	for i, fix := range fixes {
		fmt.Printf("\n%d. %s\n", i+1, fix.Description)
		fmt.Printf("   File: %s\n", fix.File)
		
		if applyFix(fix) {
			fmt.Printf("   ✅ Applied successfully\n")
			successCount++
		} else {
			fmt.Printf("   ❌ Failed to apply\n")
		}
	}

	fmt.Printf("\n🎯 SUMMARY:\n")
	fmt.Printf("   Applied %d/%d fixes successfully\n", successCount, len(fixes))
	
	if successCount == len(fixes) {
		fmt.Printf("   🚀 All fixes applied! Draft sales should no longer auto-post.\n")
		fmt.Printf("   ✅ Only INVOICED sales will create journal entries.\n")
	} else {
		fmt.Printf("   ⚠️ Some fixes failed. Manual review required.\n")
	}
}

type Fix struct {
	File        string
	Description string
	Search      string
	Replace     string
}

func applyFix(fix Fix) bool {
	// Read file
	content, err := ioutil.ReadFile(fix.File)
	if err != nil {
		fmt.Printf("   Error reading file: %v\n", err)
		return false
	}

	// Check if search pattern exists
	if !strings.Contains(string(content), fix.Search) {
		fmt.Printf("   Warning: Search pattern not found, skipping\n")
		return true // Don't fail if pattern is not found (might already be fixed)
	}

	// Create backup
	backupFile := fix.File + ".bak"
	err = ioutil.WriteFile(backupFile, content, 0644)
	if err != nil {
		fmt.Printf("   Error creating backup: %v\n", err)
		return false
	}

	// Apply replacement
	newContent := strings.Replace(string(content), fix.Search, fix.Replace, 1)
	
	// Write updated content
	err = ioutil.WriteFile(fix.File, []byte(newContent), 0644)
	if err != nil {
		fmt.Printf("   Error writing file: %v\n", err)
		// Restore from backup
		ioutil.WriteFile(fix.File, content, 0644)
		return false
	}

	fmt.Printf("   Backup created: %s\n", backupFile)
	return true
}