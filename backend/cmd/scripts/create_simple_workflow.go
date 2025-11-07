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

	fmt.Println("=== Creating Simple Purchase Approval Workflow ===")
	
	// First, deactivate all existing workflows
	fmt.Println("🔄 Deactivating existing workflows...")
	err := db.Model(&models.ApprovalWorkflow{}).
		Where("module = ?", "PURCHASE").
		Update("is_active", false).Error
	
	if err != nil {
		log.Printf("Error deactivating existing workflows: %v", err)
		return
	}
	
	// Create simple workflow: Employee -> Finance (with optional Director escalation)
	workflow := models.ApprovalWorkflow{
		Name:            "Simple Purchase Approval",
		Module:          "PURCHASE",
		MinAmount:       0,      // No minimum
		MaxAmount:       0,      // No maximum (unlimited)
		RequireDirector: false,  // Optional escalation
		RequireFinance:  true,   // Always requires finance
		IsActive:        true,
	}

	err = db.Create(&workflow).Error
	if err != nil {
		log.Printf("Failed to create workflow: %v", err)
		return
	}

	fmt.Printf("✅ Created workflow: %s (ID: %d)\n", workflow.Name, workflow.ID)

	// Create workflow steps
	steps := []models.ApprovalStep{
		{
			WorkflowID:   workflow.ID,
			StepOrder:    1,
			StepName:     "Employee Submission",
			ApproverRole: "employee",
			IsOptional:   false,
			TimeLimit:    24,
		},
		{
			WorkflowID:   workflow.ID,
			StepOrder:    2,
			StepName:     "Finance Approval",
			ApproverRole: "finance",
			IsOptional:   false,
			TimeLimit:    24,
		},
		{
			WorkflowID:   workflow.ID,
			StepOrder:    3,
			StepName:     "Director Approval (Escalation)",
			ApproverRole: "director",
			IsOptional:   true,  // Only activated when escalated
			TimeLimit:    48,
		},
	}

	for _, step := range steps {
		err = db.Create(&step).Error
		if err != nil {
			log.Printf("Failed to create workflow step: %v", err)
			return
		}
		fmt.Printf("   - Step %d: %s (Role: %s)\n", step.StepOrder, step.StepName, step.ApproverRole)
	}

	fmt.Println("\n=== Summary ===")
	fmt.Println("✅ Simple workflow created successfully!")
	fmt.Println("📋 Workflow: Employee creates purchase → Finance approves (or escalates to Director)")
	fmt.Println("💡 No amount restrictions - all purchases go through this workflow")
	fmt.Println("")
	fmt.Println("🔄 Flow:")
	fmt.Println("   1. Employee creates purchase → Status: DRAFT")
	fmt.Println("   2. Employee submits for approval → Status: PENDING")
	fmt.Println("   3. Finance can:")
	fmt.Println("      - Approve directly → Status: APPROVED")
	fmt.Println("      - Escalate to Director → Stays PENDING until Director approves")
	fmt.Println("      - Reject → Status: CANCELLED")
}