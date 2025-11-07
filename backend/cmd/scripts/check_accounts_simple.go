package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/repositories"

	"gorm.io/gorm"
)

func main() {
	log.Println("Checking Required Accounting Accounts...")

	// Initialize database
	db := database.ConnectDB()

	checkRequiredAccounts(db)
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
			fmt.Printf("❌ MISSING: Account %s - %s: %v\n", code, description, err)
			missingAccounts = append(missingAccounts, code)
		} else {
			fmt.Printf("✅ FOUND: Account %s - %s (ID: %d, Name: %s)\n", code, description, account.ID, account.Name)
		}
	}

	if len(missingAccounts) > 0 {
		fmt.Printf("\n❌ Missing %d required accounts: %v\n", len(missingAccounts), missingAccounts)
		fmt.Println("This could be causing the sale confirmation failure!")
		fmt.Println("\nTo fix this, you need to create these accounts in the Chart of Accounts.")
	} else {
		fmt.Println("\n✅ All required accounts are present")
	}
}