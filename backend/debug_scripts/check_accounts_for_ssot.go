package main

import (
	"fmt"
	"strings"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	fmt.Println("🔍 Checking Available Accounts for SSOT Integration")
	fmt.Println("=" + strings.Repeat("=", 50))
	
	db := database.ConnectDB()
	
	// Check critical accounts needed for sales journal
	criticalAccounts := map[string]string{
		"1201": "Accounts Receivable (Piutang Usaha)",
		"4001": "Sales Revenue (Pendapatan Penjualan)", 
		"2103": "PPN Keluaran",
		"1101": "Kas",
		"1102": "Bank BCA",
		"1104": "Bank Mandiri",
	}
	
	fmt.Println("Checking critical accounts for sales journal:")
	foundAccounts := 0
	
	for code, description := range criticalAccounts {
		var account models.Account
		err := db.Where("code = ?", code).First(&account).Error
		
		if err == nil {
			fmt.Printf("✅ %s - %s: %s (ID: %d)\n", code, description, account.Name, account.ID)
			foundAccounts++
		} else {
			fmt.Printf("❌ %s - %s: NOT FOUND\n", code, description)
		}
	}
	
	fmt.Printf("\nSummary: %d/%d critical accounts found\n", foundAccounts, len(criticalAccounts))
	
	if foundAccounts == len(criticalAccounts) {
		fmt.Println("🎉 All critical accounts available! SSOT integration should work.")
	} else {
		fmt.Printf("⚠️  Missing %d accounts. SSOT integration may fail.\n", len(criticalAccounts)-foundAccounts)
		fmt.Println("Please ensure the chart of accounts is properly seeded.")
	}
	
	// Show total accounts available
	var totalAccounts int64
	db.Model(&models.Account{}).Count(&totalAccounts)
	fmt.Printf("📊 Total accounts in database: %d\n", totalAccounts)
	
	// Show some sample accounts
	var accounts []models.Account
	db.Limit(10).Find(&accounts)
	fmt.Println("\n📋 Sample accounts:")
	for _, acc := range accounts {
		fmt.Printf("  %s - %s\n", acc.Code, acc.Name)
	}
}