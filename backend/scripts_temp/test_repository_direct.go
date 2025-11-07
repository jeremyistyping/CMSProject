package main

import (
	"fmt"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
)

func main() {
	db := database.ConnectDB()
	
	// Test repository directly
	fmt.Println("=== TESTING REPOSITORY FindAll() DIRECTLY ===")
	
	cashBankRepo := repositories.NewCashBankRepository(db)
	accounts, err := cashBankRepo.FindAll()
	
	if err != nil {
		fmt.Printf("Error calling FindAll(): %v\n", err)
		return
	}
	
	fmt.Printf("Total accounts returned by FindAll(): %d\n\n", len(accounts))
	
	cashCount := 0
	bankCount := 0
	totalCash := 0.0
	totalBank := 0.0
	
	for i, account := range accounts {
		fmt.Printf("%d. ID: %d | Code: %s | Name: %s | Type: %s | Balance: %.2f | Active: %t\n", 
			i+1, account.ID, account.Code, account.Name, account.Type, account.Balance, account.IsActive)
		
		if account.Type == "CASH" {
			cashCount++
			totalCash += account.Balance
		} else if account.Type == "BANK" {
			bankCount++
			totalBank += account.Balance
		}
	}
	
	fmt.Printf("\n=== REPOSITORY SUMMARY ===\n")
	fmt.Printf("Cash Accounts: %d | Total Cash: %.2f\n", cashCount, totalCash)
	fmt.Printf("Bank Accounts: %d | Total Bank: %.2f\n", bankCount, totalBank)
	fmt.Printf("Grand Total: %.2f\n", totalCash + totalBank)
	
	// Test service layer
	fmt.Println("\n=== TESTING SERVICE GetCashBankAccounts() ===")
	
	accountRepo := repositories.NewAccountRepository(db)
	cashBankService := services.NewCashBankService(db, cashBankRepo, accountRepo)
	
	serviceAccounts, err := cashBankService.GetCashBankAccounts()
	if err != nil {
		fmt.Printf("Error calling GetCashBankAccounts(): %v\n", err)
		return
	}
	
	fmt.Printf("Total accounts returned by Service: %d\n\n", len(serviceAccounts))
	
	serviceCashCount := 0
	serviceBankCount := 0
	serviceTotalCash := 0.0
	serviceTotalBank := 0.0
	
	for i, account := range serviceAccounts {
		fmt.Printf("%d. ID: %d | Code: %s | Name: %s | Type: %s | Balance: %.2f | Active: %t\n", 
			i+1, account.ID, account.Code, account.Name, account.Type, account.Balance, account.IsActive)
		
		if account.Type == "CASH" {
			serviceCashCount++
			serviceTotalCash += account.Balance
		} else if account.Type == "BANK" {
			serviceBankCount++
			serviceTotalBank += account.Balance
		}
	}
	
	fmt.Printf("\n=== SERVICE SUMMARY ===\n")
	fmt.Printf("Cash Accounts: %d | Total Cash: %.2f\n", serviceCashCount, serviceTotalCash)
	fmt.Printf("Bank Accounts: %d | Total Bank: %.2f\n", serviceBankCount, serviceTotalBank)
	fmt.Printf("Grand Total: %.2f\n", serviceTotalCash + serviceTotalBank)
	
	// Test balance summary
	fmt.Println("\n=== TESTING BALANCE SUMMARY ===")
	balanceSummary, err := cashBankService.GetBalanceSummary()
	if err != nil {
		fmt.Printf("Error getting balance summary: %v\n", err)
		return
	}
	
	fmt.Printf("Summary Total Cash: %.2f\n", balanceSummary.TotalCash)
	fmt.Printf("Summary Total Bank: %.2f\n", balanceSummary.TotalBank)
	fmt.Printf("Summary Total Balance: %.2f\n", balanceSummary.TotalBalance)
	
	// Compare results
	fmt.Println("\n=== COMPARISON ===")
	if serviceTotalCash == balanceSummary.TotalCash && serviceTotalBank == balanceSummary.TotalBank {
		fmt.Println("✅ Backend services are CONSISTENT!")
		fmt.Println("✅ Individual accounts match summary")
		fmt.Println("❌ Problem is in FRONTEND - not using the correct endpoint or data")
	} else {
		fmt.Println("❌ Backend services have INCONSISTENCY!")
		fmt.Printf("Individual vs Summary Cash: %.2f vs %.2f\n", serviceTotalCash, balanceSummary.TotalCash)
		fmt.Printf("Individual vs Summary Bank: %.2f vs %.2f\n", serviceTotalBank, balanceSummary.TotalBank)
	}
}