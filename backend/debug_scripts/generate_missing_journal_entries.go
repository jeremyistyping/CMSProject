package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

// AccountMapping holds account IDs for different transaction types
type AccountMapping struct {
	CashAccountID            uint64
	AccountsReceivableID     uint64
	SalesRevenueID          uint64
	TaxPayableID            uint64
	COGSAccountID           uint64
	InventoryAccountID      uint64
	AccountsPayableID       uint64
	TaxInputAccountID       uint64
	PurchaseExpenseID       uint64
	GeneralExpenseID        uint64
}

func main() {
	fmt.Println("🔧 Generating Missing Journal Entries from Existing Transactions")
	fmt.Println("===============================================================")

	// Initialize database
	db := database.ConnectDB()

	// Check current journal entry count
	var currentJournalCount int64
	db.Table("journal_entries").Count(&currentJournalCount)
	fmt.Printf("Current journal entries: %d\n", currentJournalCount)

	// Check existing transactions
	var salesCount, purchaseCount, paymentCount int64
	db.Table("sales").Count(&salesCount)
	db.Table("purchases").Count(&purchaseCount)
	db.Table("payments").Count(&paymentCount)
	
	fmt.Printf("Existing transactions to process:\n")
	fmt.Printf("  - Sales: %d\n", salesCount)
	fmt.Printf("  - Purchases: %d\n", purchaseCount)
	fmt.Printf("  - Payments: %d\n", paymentCount)

	if salesCount == 0 && purchaseCount == 0 && paymentCount == 0 {
		fmt.Println("No transactions found to process.")
		return
	}

	// Get account mapping
	accountMap, err := getAccountMapping(db)
	if err != nil {
		log.Fatalf("Failed to get account mapping: %v", err)
	}

	fmt.Printf("\nUsing account mapping:\n")
	fmt.Printf("  - Cash Account: %d\n", accountMap.CashAccountID)
	fmt.Printf("  - Accounts Receivable: %d\n", accountMap.AccountsReceivableID)
	fmt.Printf("  - Sales Revenue: %d\n", accountMap.SalesRevenueID)
	fmt.Printf("  - Accounts Payable: %d\n", accountMap.AccountsPayableID)

	// Skip SSOT services for now - just create basic journal entries

	// Process sales
	if salesCount > 0 {
		fmt.Printf("\n📊 Processing %d sales...\n", salesCount)
	if err := processSales(db, accountMap); err != nil {
			log.Printf("Error processing sales: %v", err)
		}
	}

	// Process purchases
	if purchaseCount > 0 {
		fmt.Printf("\n🛒 Processing %d purchases...\n", purchaseCount)
	if err := processPurchases(db, accountMap); err != nil {
			log.Printf("Error processing purchases: %v", err)
		}
	}

	// Process payments
	if paymentCount > 0 {
		fmt.Printf("\n💰 Processing %d payments...\n", paymentCount)
	if err := processPayments(db, accountMap); err != nil {
			log.Printf("Error processing payments: %v", err)
		}
	}

	// Create SSOT materialized view
	fmt.Println("\n🔄 Creating SSOT materialized views...")
	if err := createSSOTViews(db); err != nil {
		log.Printf("Error creating SSOT views: %v", err)
	}

	// Check final journal entry count
	var finalJournalCount int64
	db.Table("journal_entries").Count(&finalJournalCount)
	
	fmt.Printf("\n✅ Processing completed!\n")
	fmt.Printf("Journal entries before: %d\n", currentJournalCount)
	fmt.Printf("Journal entries after: %d\n", finalJournalCount)
	fmt.Printf("New journal entries created: %d\n", finalJournalCount-currentJournalCount)
}

func getAccountMapping(db *gorm.DB) (*AccountMapping, error) {
	var accounts []struct {
		ID   uint64 `gorm:"column:id"`
		Code string `gorm:"column:code"`
		Type string `gorm:"column:type"`
		Name string `gorm:"column:name"`
	}

	err := db.Raw(`
		SELECT id, code, type, name 
		FROM accounts 
		WHERE is_active = true AND deleted_at IS NULL
		ORDER BY code
	`).Scan(&accounts).Error

	if err != nil {
		return nil, fmt.Errorf("failed to load accounts: %w", err)
	}

	mapping := &AccountMapping{}

	// Map accounts based on their codes and types
	for _, acc := range accounts {
		switch {
		case acc.Code == "1101" || (acc.Type == "CURRENT_ASSETS" && contains(acc.Name, "Kas")):
			mapping.CashAccountID = acc.ID
		case acc.Code == "1201" || (acc.Type == "CURRENT_ASSETS" && contains(acc.Name, "Piutang")):
			mapping.AccountsReceivableID = acc.ID
		case acc.Code == "4101" || (acc.Type == "REVENUE" && contains(acc.Name, "Pendapatan")):
			mapping.SalesRevenueID = acc.ID
		case acc.Code == "2102" || (acc.Type == "CURRENT_LIABILITIES" && contains(acc.Name, "Pajak")):
			mapping.TaxPayableID = acc.ID
		case acc.Code == "5101" || (acc.Type == "EXPENSES" && contains(acc.Name, "Harga Pokok")):
			mapping.COGSAccountID = acc.ID
		case acc.Code == "1301" || (acc.Type == "CURRENT_ASSETS" && contains(acc.Name, "Persediaan")):
			mapping.InventoryAccountID = acc.ID
		case acc.Code == "2101" || (acc.Type == "CURRENT_LIABILITIES" && contains(acc.Name, "Utang Usaha")):
			mapping.AccountsPayableID = acc.ID
		case acc.Type == "CURRENT_ASSETS" && contains(acc.Name, "Pajak Masukan"):
			mapping.TaxInputAccountID = acc.ID
		case acc.Code == "5900" || (acc.Type == "EXPENSES" && contains(acc.Name, "General")):
			mapping.GeneralExpenseID = acc.ID
		}
	}

	// Set defaults if not found
	if mapping.CashAccountID == 0 {
		mapping.CashAccountID = getDefaultAccountByType(accounts, "CURRENT_ASSETS")
	}
	if mapping.AccountsReceivableID == 0 {
		mapping.AccountsReceivableID = getDefaultAccountByType(accounts, "CURRENT_ASSETS")
	}
	if mapping.SalesRevenueID == 0 {
		mapping.SalesRevenueID = getDefaultAccountByType(accounts, "REVENUE")
	}
	if mapping.AccountsPayableID == 0 {
		mapping.AccountsPayableID = getDefaultAccountByType(accounts, "CURRENT_LIABILITIES")
	}

	return mapping, nil
}

func contains(text, substr string) bool {
	return len(text) > 0 && len(substr) > 0 && 
		   (text == substr || 
		    len(text) > len(substr) && 
		    (text[:len(substr)] == substr || text[len(text)-len(substr):] == substr ||
		     findSubstring(text, substr)))
}

func findSubstring(text, substr string) bool {
	for i := 0; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func getDefaultAccountByType(accounts []struct{
	ID   uint64 `gorm:"column:id"`
	Code string `gorm:"column:code"`
	Type string `gorm:"column:type"`
	Name string `gorm:"column:name"`
}, accountType string) uint64 {
	for _, acc := range accounts {
		if acc.Type == accountType {
			return acc.ID
		}
	}
	return 1 // fallback
}

func processSales(db *gorm.DB, accountMap *AccountMapping) error {
	var sales []models.Sale
	if err := db.Find(&sales).Error; err != nil {
		return fmt.Errorf("failed to load sales: %w", err)
	}

	successCount := 0
	for _, sale := range sales {
		// Check if journal entry already exists
		var existingCount int64
		db.Table("journal_entries").Where("reference_type = ? AND reference_id = ?", "SALE", sale.ID).Count(&existingCount)
		
		if existingCount > 0 {
			continue // Skip if journal entry already exists
		}

		// Create journal entry manually since hooks need account mapping
		if err := createSaleJournalEntry(db, &sale, accountMap, 1); err != nil {
			log.Printf("Failed to create journal entry for sale %d: %v", sale.ID, err)
			continue
		}
		
		successCount++
		if successCount%5 == 0 {
			fmt.Printf("  Processed %d sales...\n", successCount)
		}
	}

	fmt.Printf("  ✅ Successfully processed %d sales\n", successCount)
	return nil
}

func processPurchases(db *gorm.DB, accountMap *AccountMapping) error {
	var purchases []models.Purchase
	if err := db.Find(&purchases).Error; err != nil {
		return fmt.Errorf("failed to load purchases: %w", err)
	}

	successCount := 0
	for _, purchase := range purchases {
		// Check if journal entry already exists
		var existingCount int64
		db.Table("journal_entries").Where("reference_type = ? AND reference_id = ?", "PURCHASE", purchase.ID).Count(&existingCount)
		
		if existingCount > 0 {
			continue // Skip if journal entry already exists
		}

		// Create journal entry manually
		if err := createPurchaseJournalEntry(db, &purchase, accountMap, 1); err != nil {
			log.Printf("Failed to create journal entry for purchase %d: %v", purchase.ID, err)
			continue
		}
		
		successCount++
	}

	fmt.Printf("  ✅ Successfully processed %d purchases\n", successCount)
	return nil
}

func processPayments(db *gorm.DB, accountMap *AccountMapping) error {
	var payments []models.Payment
	if err := db.Find(&payments).Error; err != nil {
		return fmt.Errorf("failed to load payments: %w", err)
	}

	successCount := 0
	for _, payment := range payments {
		// Check if journal entry already exists
		var existingCount int64
		db.Table("journal_entries").Where("reference_type = ? AND reference_id = ?", "PAYMENT", payment.ID).Count(&existingCount)
		
		if existingCount > 0 {
			continue // Skip if journal entry already exists
		}

		// Create journal entry manually
		if err := createPaymentJournalEntry(db, &payment, accountMap, 1); err != nil {
			log.Printf("Failed to create journal entry for payment %d: %v", payment.ID, err)
			continue
		}
		
		successCount++
	}

	fmt.Printf("  ✅ Successfully processed %d payments\n", successCount)
	return nil
}

func createSaleJournalEntry(db *gorm.DB, sale *models.Sale, accountMap *AccountMapping, userID uint64) error {
	// Generate entry number
	entryNumber := fmt.Sprintf("JE-SALE-%s-%d", sale.Date.Format("2006"), sale.ID)
	
	// Create journal entry
	entry := &models.JournalEntry{
		Code:         entryNumber,
		ReferenceType: "SALE",
		ReferenceID:  &sale.ID,
		EntryDate:    sale.Date,
		Description:  fmt.Sprintf("Sale Invoice #%s", sale.InvoiceNumber),
		Status:       "POSTED",
		UserID:       uint(userID),
		TotalDebit:   sale.TotalAmount,
		TotalCredit:  sale.TotalAmount,
	}

	if err := db.Create(entry).Error; err != nil {
		return fmt.Errorf("failed to create journal entry: %w", err)
	}

	fmt.Printf("    Created journal entry %d for sale %d\n", entry.ID, sale.ID)
	return nil
}

func createPurchaseJournalEntry(db *gorm.DB, purchase *models.Purchase, accountMap *AccountMapping, userID uint64) error {
	// Generate entry number
	entryNumber := fmt.Sprintf("JE-PURCHASE-%s-%d", purchase.Date.Format("2006"), purchase.ID)
	
	// Create journal entry
	entry := &models.JournalEntry{
		Code:         entryNumber,
		ReferenceType: "PURCHASE",
		ReferenceID:  &purchase.ID,
		EntryDate:    purchase.Date,
		Description:  fmt.Sprintf("Purchase Order #%s", purchase.Code),
		Status:       "POSTED",
		UserID:       uint(userID),
		TotalDebit:   purchase.TotalAmount,
		TotalCredit:  purchase.TotalAmount,
	}

	if err := db.Create(entry).Error; err != nil {
		return fmt.Errorf("failed to create journal entry: %w", err)
	}

	fmt.Printf("    Created journal entry %d for purchase %d\n", entry.ID, purchase.ID)
	return nil
}

func createPaymentJournalEntry(db *gorm.DB, payment *models.Payment, accountMap *AccountMapping, userID uint64) error {
	// Generate entry number
	entryNumber := fmt.Sprintf("JE-PAYMENT-%s-%d", payment.Date.Format("2006"), payment.ID)
	
	// Create journal entry
	entry := &models.JournalEntry{
		Code:         entryNumber,
		ReferenceType: "PAYMENT",
		ReferenceID:  &payment.ID,
		EntryDate:    payment.Date,
		Description:  fmt.Sprintf("Payment #%s", payment.Reference),
		Status:       "POSTED",
		UserID:       uint(userID),
		TotalDebit:   payment.Amount,
		TotalCredit:  payment.Amount,
	}

	if err := db.Create(entry).Error; err != nil {
		return fmt.Errorf("failed to create journal entry: %w", err)
	}

	fmt.Printf("    Created journal entry %d for payment %d\n", entry.ID, payment.ID)
	return nil
}

func createSSOTViews(db *gorm.DB) error {
	// Create account_balances materialized view
	createViewSQL := `
		CREATE MATERIALIZED VIEW IF NOT EXISTS account_balances AS
		SELECT 
			a.id as account_id,
			a.code as account_code,
			a.name as account_name,
			a.type as account_type,
			COALESCE(a.balance, 0) as current_balance,
			CURRENT_TIMESTAMP as last_updated
		FROM accounts a
		WHERE a.is_active = true AND a.deleted_at IS NULL;
	`

	if err := db.Exec(createViewSQL).Error; err != nil {
		return fmt.Errorf("failed to create account_balances view: %w", err)
	}

	// Create index on account_balances view
	indexSQL := `
		CREATE INDEX IF NOT EXISTS idx_account_balances_account_id 
		ON account_balances(account_id);
	`

	if err := db.Exec(indexSQL).Error; err != nil {
		log.Printf("Warning: failed to create index on account_balances: %v", err)
	}

	// Refresh the materialized view
	if err := db.Exec("REFRESH MATERIALIZED VIEW account_balances").Error; err != nil {
		return fmt.Errorf("failed to refresh account_balances view: %w", err)
	}

	fmt.Println("  ✅ Created and refreshed account_balances materialized view")
	return nil
}