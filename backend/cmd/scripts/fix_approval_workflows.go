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

	fmt.Println("=== Fixing Purchase Approval Workflows to Start from Employee ===")
	
	// Begin transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get all purchase workflows
	var workflows []models.ApprovalWorkflow
	err := tx.Preload("Steps").Where("module = ? AND is_active = ?", "PURCHASE", true).Find(&workflows).Error
	if err != nil {
		tx.Rollback()
		log.Printf("Error querying workflows: %v", err)
		return
	}

	for _, workflow := range workflows {
		fmt.Printf("\n🔧 Updating Workflow: %s (ID: %d)\n", workflow.Name, workflow.ID)
		
		// First, increment all existing step orders by 1 to make room for Employee step
		err := tx.Model(&models.ApprovalStep{}).
			Where("workflow_id = ?", workflow.ID).
			Update("step_order", db.Raw("step_order + 1")).Error
		if err != nil {
			tx.Rollback()
			log.Printf("Failed to update step orders: %v", err)
			return
		}

		// Create new Employee step as step 1
		employeeStep := models.ApprovalStep{
			WorkflowID:   workflow.ID,
			StepOrder:    1,
			StepName:     "Employee Submission",
			ApproverRole: "employee",
			IsOptional:   false,
			TimeLimit:    24, // 24 hours
		}

		err = tx.Create(&employeeStep).Error
		if err != nil {
			tx.Rollback()
			log.Printf("Failed to create Employee step: %v", err)
			return
		}

		fmt.Printf("   ✅ Added Employee step as Step 1\n")
		fmt.Printf("   ✅ Shifted other steps up by 1\n")
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return
	}

	fmt.Println("\n=== Verification ===")
	
	// Verify the updated workflows
	var updatedWorkflows []models.ApprovalWorkflow
	err = db.Preload("Steps").Where("module = ? AND is_active = ?", "PURCHASE", true).Find(&updatedWorkflows).Error
	if err != nil {
		log.Printf("Error verifying workflows: %v", err)
		return
	}

	for _, w := range updatedWorkflows {
		fmt.Printf("\n📋 Updated Workflow: %s (ID: %d)\n", w.Name, w.ID)
		fmt.Printf("   Steps:\n")
		
		for _, s := range w.Steps {
			fmt.Printf("   %d. %s (Role: %s) - Time Limit: %dh\n", 
				s.StepOrder, s.StepName, s.ApproverRole, s.TimeLimit)
		}
	}

	fmt.Println("\n✅ All workflows now start from Employee!")
	fmt.Println("The approval trail will now be: Employee → Finance/Manager → Director (based on amount)")
}
