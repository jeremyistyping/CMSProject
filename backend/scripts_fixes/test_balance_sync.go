package main

import (
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
)

func main() {
	log.Println("🧪 Testing Balance Sync with New Sale...")
	
	// Connect to database
	db := database.ConnectDB()
	
	// Initialize services
	balanceSync := services.NewBalanceSyncService(db)
	
	// Check initial balances
	fmt.Println("\n=== 📊 INITIAL ACCOUNT BALANCES ===")
	var initialBalances []struct {
		Code     string  `json:"code"`
		Name     string  `json:"name"`
		Balance  float64 `json:"balance"`
	}
	
	err := db.Raw(`
		SELECT code, name, balance 
		FROM accounts 
		WHERE code IN ('1201', '4101', '2102') 
		ORDER BY code
	`).Scan(&initialBalances).Error
	
	if err != nil {
		log.Printf("Error fetching initial balances: %v", err)
		return
	}
	
	for _, balance := range initialBalances {
		fmt.Printf("  %s: %s - Balance: %.2f\n", balance.Code, balance.Name, balance.Balance)
	}
	
	// Verify integrity before test
	fmt.Println("\n=== 🔍 INTEGRITY CHECK BEFORE TEST ===")
	isConsistent, err := balanceSync.VerifyBalanceIntegrity()
	if err != nil {
		log.Printf("Error during integrity check: %v", err)
		return
	}
	
	if isConsistent {
		fmt.Println("✅ All balances are consistent before test")
	} else {
		fmt.Println("❌ Balance inconsistencies detected before test")
		balanceSync.SyncAccountBalancesFromSSOT()
	}
	
	// Create a test sale
	fmt.Println("\n=== 🛒 CREATING TEST SALE ===")
	
	// Find a customer
	var customer models.Contact
	err = db.Where("type = 'CUSTOMER'").First(&customer).Error
	if err != nil {
		log.Printf("Error finding customer: %v", err)
		return
	}
	
	// Create new sale
	newSale := &models.Sale{
		Code:          fmt.Sprintf("TST-%d", time.Now().Unix()),
		CustomerID:    customer.ID,
		UserID:        1, // Assume admin user
		Type:          models.SaleTypeInvoice,
		Date:          time.Now(),
		DueDate:       time.Now().Add(30 * 24 * time.Hour),
		InvoiceNumber: fmt.Sprintf("INV/TST/%d", time.Now().Unix()),
		Currency:      "IDR",
		ExchangeRate:  1,
		TotalAmount:   1000000, // Rp 1,000,000
		PaidAmount:    0,
		OutstandingAmount: 1000000,
		Subtotal:      909090.91, // ~909k before 11% PPN
		PPNPercent:    11,
		PPN:           90909.09,  // 11% PPN
		Status:        models.SaleStatusInvoiced,
	}
	
	err = db.Create(newSale).Error
	if err != nil {
		log.Printf("Error creating test sale: %v", err)
		return
	}
	
	fmt.Printf("✅ Created test sale ID %d: %s\n", newSale.ID, newSale.InvoiceNumber)
	
	// Create SSOT journal entry for the sale
	fmt.Println("\n=== 📝 CREATING SSOT JOURNAL ENTRY ===")
	
	ssotService := services.NewSSOTSalesJournalService(db)
	
	// Load customer relationship
	db.Preload("Customer").First(newSale, newSale.ID)
	
	journalEntry, err := ssotService.CreateSaleJournalEntry(newSale, 1)
	if err != nil {
		log.Printf("Error creating SSOT journal entry: %v", err)
		return
	}
	
	fmt.Printf("✅ Created SSOT journal entry ID: %d\n", journalEntry.ID)
	
	// Wait a moment for triggers to process
	time.Sleep(2 * time.Second)
	
	// Check final balances
	fmt.Println("\n=== 📈 FINAL ACCOUNT BALANCES ===")
	var finalBalances []struct {
		Code     string  `json:"code"`
		Name     string  `json:"name"`
		Balance  float64 `json:"balance"`
	}
	
	err = db.Raw(`
		SELECT code, name, balance 
		FROM accounts 
		WHERE code IN ('1201', '4101', '2102') 
		ORDER BY code
	`).Scan(&finalBalances).Error
	
	if err != nil {
		log.Printf("Error fetching final balances: %v", err)
		return
	}
	
	for _, balance := range finalBalances {
		fmt.Printf("  %s: %s - Balance: %.2f\n", balance.Code, balance.Name, balance.Balance)
	}
	
	// Verify integrity after test
	fmt.Println("\n=== ✅ INTEGRITY CHECK AFTER TEST ===")
	isConsistent, err = balanceSync.VerifyBalanceIntegrity()
	if err != nil {
		log.Printf("Error during final integrity check: %v", err)
		return
	}
	
	if isConsistent {
		fmt.Println("✅ All balances are consistent after test")
	} else {
		fmt.Println("❌ Balance inconsistencies detected after test")
	}
	
	// Calculate balance changes
	fmt.Println("\n=== 📊 BALANCE CHANGES SUMMARY ===")
	for i, initial := range initialBalances {
		if i < len(finalBalances) {
			final := finalBalances[i]
			change := final.Balance - initial.Balance
			fmt.Printf("  %s: %s\n", final.Code, final.Name)
			fmt.Printf("    Before: %.2f → After: %.2f → Change: %.2f\n", 
				initial.Balance, final.Balance, change)
		}
	}
	
	fmt.Println("\n🎉 Balance sync test completed successfully!")
}