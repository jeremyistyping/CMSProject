package main

import (
	"fmt"
	"log"
	"os"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"

	"gorm.io/gorm"
)

func main() {
	log.Println("Testing Sale Confirmation Issues...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize database
	db, err := config.NewDatabase(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Test 1: Check required accounting accounts
	checkRequiredAccounts(db)

	// Test 2: Test a specific sale confirmation
	testSaleConfirmation(db)
}

func checkRequiredAccounts(db *gorm.DB) {
	log.Println("\n=== Checking Required Accounting Accounts ===")

	accountRepo := repositories.NewAccountRepository(db)
	requiredAccounts := map[string]string{
		"1201": "Piutang Usaha (Accounts Receivable)",
		"1200": "Piutang Usaha Alternative",
		"4101": "Pendapatan Penjualan (Sales Revenue)",
		"4100": "Pendapatan Penjualan Alternative", 
		"5101": "Harga Pokok Penjualan (COGS)",
		"5100": "COGS Alternative",
		"1301": "Persediaan Barang Dagangan (Inventory)",
		"1300": "Inventory Alternative",
		"2102": "Utang Pajak (Tax Payable)",
		"4102": "Shipping Revenue",
	}

	missingAccounts := []string{}
	for code, description := range requiredAccounts {
		account, err := accountRepo.GetAccountByCode(code)
		if err != nil {
			log.Printf("❌ MISSING: Account %s - %s: %v", code, description, err)
			missingAccounts = append(missingAccounts, code)
		} else {
			log.Printf("✅ FOUND: Account %s - %s (ID: %d, Name: %s)", code, description, account.ID, account.Name)
		}
	}

	if len(missingAccounts) > 0 {
		log.Printf("\n❌ Missing %d required accounts: %v", len(missingAccounts), missingAccounts)
		log.Println("This could be causing the sale confirmation failure!")
	} else {
		log.Println("\n✅ All required accounts are present")
	}
}

func testSaleConfirmation(db *gorm.DB) {
	log.Println("\n=== Testing Sale Confirmation ===")

	// Initialize repositories and services
	salesRepo := repositories.NewSalesRepository(db)
	productRepo := repositories.NewProductRepository(db)
	contactRepo := repositories.NewContactRepository(db)
	accountRepo := repositories.NewAccountRepository(db)

	// Create services (journal service can be nil for this test)
	salesService := services.NewSalesService(db, salesRepo, productRepo, contactRepo, accountRepo, nil, nil)

	// Test with sale ID 21 (or find the first draft sale)
	saleID := uint(21)
	sale, err := salesService.GetSaleByID(saleID)
	if err != nil {
		log.Printf("❌ Failed to find sale %d: %v", saleID, err)
		
		// Try to find any draft sale
		var draftSales []models.Sale
		if err := db.Where("status = ?", "DRAFT").Limit(5).Find(&draftSales).Error; err == nil && len(draftSales) > 0 {
			log.Printf("Found %d draft sales to test with:", len(draftSales))
			for _, s := range draftSales {
				log.Printf("  - Sale ID: %d, Code: %s, Status: %s, Customer ID: %d", s.ID, s.Code, s.Status, s.CustomerID)
			}
			saleID = draftSales[0].ID
			sale = &draftSales[0]
		} else {
			log.Println("❌ No draft sales found to test with")
			return
		}
	}

	log.Printf("Testing confirmation of Sale ID: %d, Code: %s, Status: %s", sale.ID, sale.Code, sale.Status)

	if sale.Status != "DRAFT" {
		log.Printf("❌ Sale %d is not in DRAFT status (current: %s), cannot confirm", saleID, sale.Status)
		return
	}

	// Test the confirmation process
	userID := uint(1) // Use admin user
	log.Printf("Attempting to confirm sale %d with user ID %d...", saleID, userID)

	err = salesService.ConfirmSale(saleID, userID)
	if err != nil {
		log.Printf("❌ FAILED to confirm sale: %v", err)
		
		// Analyze the error
		errorStr := err.Error()
		if contains(errorStr, "account not found") {
			log.Println("💡 This appears to be a missing account error")
		} else if contains(errorStr, "failed to find sale") {
			log.Println("💡 Sale lookup issue")
		} else if contains(errorStr, "failed to update inventory") {
			log.Println("💡 Inventory update issue")
		} else if contains(errorStr, "failed to create journal entries") {
			log.Println("💡 Journal entry creation issue")
		} else if contains(errorStr, "failed to update sale") {
			log.Println("💡 Database update issue")
		}
	} else {
		log.Printf("✅ SUCCESS: Sale %d confirmed successfully!", saleID)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}