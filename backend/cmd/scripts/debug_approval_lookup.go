package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
)

func main() {
	// Load configuration
	_ = config.LoadConfig()

	// Connect to database
	db := database.ConnectDB()

	fmt.Println("=== Debug Approval Workflow Lookup ===")
	
	// Create approval service
	approvalService := services.NewApprovalService(db)
	
	// Test amounts
	testAmounts := []float64{
		11100000,   // 11.1M - should match first workflow
		5000000,    // 5M - should match first workflow
		30000000,   // 30M - should match second workflow
		150000000,  // 150M - should match third workflow
	}
	
	for _, amount := range testAmounts {
		fmt.Printf("\n--- Testing amount: %.0f ---\n", amount)
		
		workflow, err := approvalService.GetWorkflowByAmount(models.ApprovalModulePurchase, amount)
		
		if err != nil {
			fmt.Printf("❌ Error finding workflow: %v\n", err)
		} else if workflow == nil {
			fmt.Printf("❌ No workflow found for amount %.0f\n", amount)
		} else {
			fmt.Printf("✅ Found workflow: %s (ID: %d)\n", workflow.Name, workflow.ID)
			fmt.Printf("   Range: %.0f - %.0f\n", workflow.MinAmount, workflow.MaxAmount)
			fmt.Printf("   Requires approval: YES\n")
		}
	}
	
	// Test the specific logic from setApprovalBasisAndBase
	fmt.Printf("\n=== Testing setApprovalBasisAndBase Logic ===\n")
	amount := 11100000.0
	
	// Simulate the exact logic from the method
	requiredWorkflow, err := approvalService.GetWorkflowByAmount(models.ApprovalModulePurchase, amount)
	
	fmt.Printf("Amount: %.0f\n", amount)
	fmt.Printf("Error: %v\n", err)
	fmt.Printf("Workflow: %v\n", requiredWorkflow)
	
	if err == nil && requiredWorkflow != nil {
		fmt.Printf("Result: RequiresApproval = true\n")
		fmt.Printf("Workflow found: %s (ID: %d)\n", requiredWorkflow.Name, requiredWorkflow.ID)
	} else {
		fmt.Printf("Result: RequiresApproval = false\n")
		fmt.Printf("Reason: err=%v, workflow=%v\n", err, requiredWorkflow)
	}
}