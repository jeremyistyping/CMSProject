package main

import (
	"fmt"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/services"
)

func main() {
	fmt.Println("=== RUNNING BALANCE INTEGRITY VALIDATION ===")

	// Connect to database
	db := database.ConnectDB()
	
	// Initialize balance integrity validator
	validator := services.NewBalanceIntegrityValidator(db)
	
	fmt.Println("🔍 Starting comprehensive balance integrity validation...")
	
	// Run validation
	result, err := validator.ValidateAllBalances()
	if err != nil {
		fmt.Printf("❌ Validation error: %v\n", err)
		return
	}
	
	fmt.Printf("\n=== VALIDATION RESULTS ===\n")
	fmt.Printf("Total Checks: %d\n", result.TotalChecks)
	fmt.Printf("Failed Checks: %d\n", result.FailedChecks)
	fmt.Printf("Inconsistencies Found: %d\n", len(result.Inconsistencies))
	fmt.Printf("Validation Status: %v\n", result.Valid)
	fmt.Printf("Validation Duration: %v\n", result.ValidationDuration)
	
	if len(result.Inconsistencies) > 0 {
		fmt.Printf("\n⚠️  INCONSISTENCIES FOUND:\n")
		for i, issue := range result.Inconsistencies {
			fmt.Printf("  [%d] %s [%s] %s: %s\n", 
				i+1, issue.Severity, issue.Type, issue.EntityName, issue.Description)
			fmt.Printf("       Expected: %.2f, Actual: %.2f, Difference: %.2f\n", 
				issue.ExpectedBalance, issue.ActualBalance, issue.Difference)
		}
	}
	
	// Check for critical double-posting patterns
	doublePostingCount := 0
	for _, issue := range result.Inconsistencies {
		if issue.Type == "CASHBANK_TRANSACTION" && issue.Difference > 0 {
			doublePostingCount++
		}
	}
	if doublePostingCount > 0 {
		fmt.Printf("\n🔴 POTENTIAL DOUBLE POSTING: %d cases detected\n", doublePostingCount)
	}
	
	if result.Valid {
		fmt.Printf("\n✅ All account balances are consistent!\n")
	} else {
		fmt.Printf("\n❌ Balance inconsistencies detected. Please review and fix.\n")
		fmt.Printf("💡 Use SSOT payment endpoints (/api/v1/payments/ssot/*) to prevent future double posting\n")
	}
	
	fmt.Printf("\n=== VALIDATION COMPLETE ===\n")
}