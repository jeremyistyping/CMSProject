package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	// Load configuration
	_ = config.LoadConfig()

	// Connect to database
	db := database.ConnectDB()

	fmt.Println("=== Checking Existing Approval Workflows ===")
	
	// Check existing workflows
	var workflows []models.ApprovalWorkflow
	err := db.Find(&workflows).Error
	if err != nil {
		log.Printf("Error querying workflows: %v", err)
		return
	}

	fmt.Printf("Found %d existing workflow(s):\n", len(workflows))
	for _, w := range workflows {
		fmt.Printf("  - ID: %d, Name: %s, Module: %s, Active: %t\n", w.ID, w.Name, w.Module, w.IsActive)
		fmt.Printf("    Amount Range: %.0f - %.0f (0 means no limit)\n", w.MinAmount, w.MaxAmount)
	}

	// Check if we have any active PURCHASE workflows
	var purchaseWorkflows []models.ApprovalWorkflow
	err = db.Where("module = ? AND is_active = ?", "PURCHASE", true).Find(&purchaseWorkflows).Error
	if err != nil {
		log.Printf("Error querying purchase workflows: %v", err)
		return
	}

	if len(purchaseWorkflows) == 0 {
		fmt.Println("\n❌ No active PURCHASE workflows found!")
		fmt.Println("Creating default purchase approval workflow...")
		
		// Create default workflow
		workflow := models.ApprovalWorkflow{
			Name:            "Default Purchase Approval",
			Module:          "PURCHASE",
			MinAmount:       0,
			MaxAmount:       0, // No limit
			RequireDirector: false,
			RequireFinance:  true,
			IsActive:        true,
		}

		err = db.Create(&workflow).Error
		if err != nil {
			log.Printf("Failed to create workflow: %v", err)
			return
		}

		// Create workflow steps
		step := models.ApprovalStep{
			WorkflowID:   workflow.ID,
			StepOrder:    1,
			StepName:     "Finance Approval",
			ApproverRole: "finance",
			IsOptional:   false,
			TimeLimit:    24,
		}

		err = db.Create(&step).Error
		if err != nil {
			log.Printf("Failed to create workflow step: %v", err)
			return
		}

		fmt.Printf("✅ Created default workflow with ID %d\n", workflow.ID)
		fmt.Printf("   - Workflow: %s (Module: %s)\n", workflow.Name, workflow.Module)
		fmt.Printf("   - Step: %s (Role: %s)\n", step.StepName, step.ApproverRole)
	} else {
		fmt.Printf("\n✅ Found %d active PURCHASE workflow(s):\n", len(purchaseWorkflows))
		for _, w := range purchaseWorkflows {
			fmt.Printf("   - ID: %d, Name: %s\n", w.ID, w.Name)
		}
	}

	fmt.Println("\n=== Summary ===")
	fmt.Println("Workflow setup complete. The system now has workflows for rejection tracking.")
}
