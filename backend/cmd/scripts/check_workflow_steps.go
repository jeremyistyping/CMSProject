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

	fmt.Println("=== Current Purchase Approval Workflows ===")
	
	// Get all purchase workflows with steps
	var workflows []models.ApprovalWorkflow
	err := db.Preload("Steps").Where("module = ? AND is_active = ?", "PURCHASE", true).Find(&workflows).Error
	if err != nil {
		log.Printf("Error querying workflows: %v", err)
		return
	}

	for _, w := range workflows {
		fmt.Printf("\n📋 Workflow: %s (ID: %d)\n", w.Name, w.ID)
		fmt.Printf("   Amount Range: %.0f - %.0f\n", w.MinAmount, w.MaxAmount)
		fmt.Printf("   Steps:\n")
		
		for _, s := range w.Steps {
			fmt.Printf("   %d. %s (Role: %s) - Time Limit: %dh\n", 
				s.StepOrder, s.StepName, s.ApproverRole, s.TimeLimit)
		}
	}

	fmt.Println("\n=== Issue Analysis ===")
	fmt.Println("The approval trail should start from EMPLOYEE when a purchase form is submitted.")
	fmt.Println("Current workflows seem to start from Finance/Manager instead of Employee.")
	fmt.Println("We need to update workflows to include Employee as the first step.")
}
