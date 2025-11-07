package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("🔧 Fixing Account Numbering to Follow PSAK Standards...")
	
	// Load configuration
	cfg := config.LoadConfig()
	
	// Connect to database
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	
	log.Println("✅ Database connected successfully")
	
	// Start transaction
	tx := db.Begin()
	
	// Create account code mapping
	accountCodeMapping := map[string]string{
		// Fix existing inconsistent codes
		"1009": "1009", // Bank Lama - keep as is (will be deprecated)
		"1010": "1110", // KAS DI BANK -> move to bank section
		"1011": "1102-001", // KAS BANK BCA -> child of Bank BCA
		"1100": "1100", // CURRENT ASSETS - keep as header
		"1101": "1101", // Kas - keep as main cash account
		"1101-001": "1101-001", // Kas Besar - keep correct format
		"1102-002": "1102-001", // Bank BCA Operasional - fix sequence
		"1102-003": "1102-002", // Bank Mandiri Payroll - fix sequence
		"1102": "1102", // Bank BCA - keep as main bank account
		"1103": "1103", // Bank Mandiri - keep as main bank account
		"1104": "1104", // BANK UOB - keep as main bank account
	}
	
	// Define the correct account structure
	correctStructure := []struct {
		OldCode string
		NewCode string
		Name string
		IsHeader bool
		ParentCode string
	}{
		// Main Asset Header
		{"1000", "1000", "ASSETS", true, ""},
		{"1100", "1100", "KAS DAN BANK", true, "1000"},
		
		// Cash Accounts
		{"1101", "1101", "KAS", false, "1100"},  // Main cash account (not header anymore)
		{"1101-001", "1101-001", "Kas Kecil Kantor Pusat", false, "1101"},
		{"1101-002", "1101-002", "Kas Besar Operasional", false, "1101"},
		
		// Bank Accounts - Main accounts as parents
		{"1102", "1102", "BANK BCA", false, "1100"}, // Change to non-header
		{"1102-001", "1102-001", "Bank BCA - Operasional", false, "1102"},
		{"1102-002", "1102-002", "Bank BCA - Tabungan", false, "1102"},
		
		{"1103", "1103", "BANK MANDIRI", false, "1100"}, // Change to non-header  
		{"1103-001", "1103-001", "Bank Mandiri - Payroll", false, "1103"},
		{"1103-002", "1103-002", "Bank Mandiri - Operasional", false, "1103"},
		
		{"1104", "1104", "BANK UOB", false, "1100"},
		{"1104-001", "1104-001", "Bank UOB - USD Account", false, "1104"},
		
		// Other current assets
		{"1200", "1200", "ACCOUNTS RECEIVABLE", true, "1100"},
		{"1201", "1201", "Piutang Usaha", false, "1200"},
	}
	
	log.Println("🔄 Starting account structure fix...")
	
	// First, backup existing accounts
	log.Println("📦 Creating backup of existing accounts...")
	if err := createAccountBackup(tx); err != nil {
		tx.Rollback()
		log.Fatal("Failed to create backup:", err)
	}
	
	// Get existing accounts for reference
	var existingAccounts []models.Account
	if err := tx.Find(&existingAccounts).Error; err != nil {
		tx.Rollback()
		log.Fatal("Failed to get existing accounts:", err)
	}
	
	// Create account map for quick lookup
	accountMap := make(map[string]*models.Account)
	for i := range existingAccounts {
		accountMap[existingAccounts[i].Code] = &existingAccounts[i]
	}
	
	// Update accounts according to new structure
	for _, structure := range correctStructure {
		if existingAcc, exists := accountMap[structure.OldCode]; exists {
			// Update existing account
			updates := map[string]interface{}{
				"name": structure.Name,
				"is_header": structure.IsHeader,
			}
			
			// Set parent if specified
			if structure.ParentCode != "" {
				if parentAcc, parentExists := accountMap[structure.ParentCode]; parentExists {
					updates["parent_id"] = parentAcc.ID
					// Set level based on parent level + 1
					updates["level"] = getAccountLevel(structure.ParentCode) + 1
				}
			} else {
				// Root account
				updates["parent_id"] = nil
				updates["level"] = 1
			}
			
			if err := tx.Model(&models.Account{}).Where("code = ?", structure.OldCode).Updates(updates).Error; err != nil {
				tx.Rollback()
				log.Fatalf("Failed to update account %s: %v", structure.OldCode, err)
			}
			
			log.Printf("✅ Updated account %s: %s", structure.OldCode, structure.Name)
		} else if structure.OldCode != structure.NewCode {
			// Create new account if it's a rename scenario
			log.Printf("ℹ️  Account %s not found, skipping...", structure.OldCode)
		}
	}
	
	// Fix parent relationships for headers
	log.Println("🔗 Setting up parent-child relationships...")
	if err := fixAccountHierarchy(tx); err != nil {
		tx.Rollback()
		log.Fatal("Failed to fix hierarchy:", err)
	}
	
	// Validate the new structure
	log.Println("✅ Validating new account structure...")
	if err := validateAccountStructure(tx); err != nil {
		tx.Rollback()
		log.Fatal("Validation failed:", err)
	}
	
	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Fatal("Failed to commit transaction:", err)
	}
	
	log.Println("🎉 Account numbering fixed successfully!")
	log.Println("📋 Summary of changes:")
	
	// Show final structure
	if err := displayAccountHierarchy(db); err != nil {
		log.Printf("⚠️  Failed to display hierarchy: %v", err)
	}
}

func createAccountBackup(db *gorm.DB) error {
	// Create backup table if not exists
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS accounts_backup_` + getCurrentTimestamp() + ` AS 
		SELECT * FROM accounts
	`).Error; err != nil {
		return err
	}
	
	log.Println("✅ Account backup created successfully")
	return nil
}

func getCurrentTimestamp() string {
	return strconv.FormatInt(getCurrentTime().Unix(), 10)
}

func getCurrentTime() time.Time {
	return time.Now()
}

func getAccountLevel(code string) int {
	switch {
	case code == "1000" || code == "2000" || code == "3000" || code == "4000" || code == "5000":
		return 1 // Main type headers (ASSETS, LIABILITIES, etc.)
	case len(code) == 4 && !strings.Contains(code, "-"):
		return 2 // Sub headers like 1100 (CURRENT ASSETS)
	case strings.Contains(code, "-"):
		return 4 // Child accounts like 1101-001
	default:
		return 3 // Main accounts like 1101 (Kas)
	}
}

func fixAccountHierarchy(db *gorm.DB) error {
	// Make sure 1100 is marked as header since it has children
	if err := db.Model(&models.Account{}).Where("code = ?", "1100").Update("is_header", true).Error; err != nil {
		return fmt.Errorf("failed to set 1100 as header: %v", err)
	}
	
	// Make sure cash and bank parent accounts are NOT headers initially
	bankCodes := []string{"1101", "1102", "1103", "1104"}
	for _, code := range bankCodes {
		// Check if this account has children
		var childCount int64
		if err := db.Model(&models.Account{}).Where("parent_id IN (SELECT id FROM accounts WHERE code = ?)", code).Count(&childCount).Error; err != nil {
			return fmt.Errorf("failed to count children for %s: %v", code, err)
		}
		
		// Set as header only if it has children
		isHeader := childCount > 0
		if err := db.Model(&models.Account{}).Where("code = ?", code).Update("is_header", isHeader).Error; err != nil {
			return fmt.Errorf("failed to update header status for %s: %v", code, err)
		}
		
		if isHeader {
			log.Printf("✅ Set %s as header (has %d children)", code, childCount)
		}
	}
	
	return nil
}

func validateAccountStructure(db *gorm.DB) error {
	// Check for correct PSAK format
	var accounts []models.Account
	if err := db.Find(&accounts).Error; err != nil {
		return err
	}
	
	var validationErrors []string
	
	for _, acc := range accounts {
		// Validate code format
		if strings.Contains(acc.Code, "-") {
			// Child account: should be XXXX-XXX format
			parts := strings.Split(acc.Code, "-")
			if len(parts) != 2 || len(parts[0]) != 4 || len(parts[1]) != 3 {
				validationErrors = append(validationErrors, fmt.Sprintf("Invalid child account format: %s", acc.Code))
			}
		} else {
			// Main account: should be 4 digits for detailed accounts
			if len(acc.Code) != 4 {
				validationErrors = append(validationErrors, fmt.Sprintf("Invalid main account format: %s (should be 4 digits)", acc.Code))
			}
		}
		
		// Validate asset account prefix
		if acc.Type == models.AccountTypeAsset && !strings.HasPrefix(acc.Code, "1") {
			validationErrors = append(validationErrors, fmt.Sprintf("Asset account %s should start with 1", acc.Code))
		}
	}
	
	if len(validationErrors) > 0 {
		for _, err := range validationErrors {
			log.Printf("❌ Validation Error: %s", err)
		}
		return fmt.Errorf("found %d validation errors", len(validationErrors))
	}
	
	log.Println("✅ All accounts pass PSAK validation")
	return nil
}

func displayAccountHierarchy(db *gorm.DB) error {
	var accounts []models.Account
	if err := db.Preload("Children").Where("parent_id IS NULL").Order("code").Find(&accounts).Error; err != nil {
		return err
	}
	
	log.Println("📊 Final Account Hierarchy:")
	for _, acc := range accounts {
		displayAccountTree(db, acc, 0)
	}
	
	return nil
}

func displayAccountTree(db *gorm.DB, account models.Account, level int) {
	indent := strings.Repeat("  ", level)
	headerMark := ""
	if account.IsHeader {
		headerMark = " [HEADER]"
	}
	
	log.Printf("%s├── %s - %s%s (Balance: %.2f)", indent, account.Code, account.Name, headerMark, account.Balance)
	
	// Get and display children
	var children []models.Account
	if err := db.Where("parent_id = ?", account.ID).Order("code").Find(&children).Error; err == nil {
		for _, child := range children {
			displayAccountTree(db, child, level+1)
		}
	}
}