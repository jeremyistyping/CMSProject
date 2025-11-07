package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	// Initialize database
	db := database.ConnectDB()

	fmt.Println("🔍 Diagnosing Approval Workflow Issue...")
	
	// Check the specific approval request that's failing
	var request models.ApprovalRequest
	err := db.Preload("Workflow").Preload("Workflow.Steps").
		Preload("ApprovalSteps").Preload("ApprovalSteps.Step").
		Where("id = ?", 29).First(&request).Error
	
	if err != nil {
		log.Printf("Failed to find approval request 29: %v", err)
		
		// Let's check what approval requests exist
		var requests []models.ApprovalRequest
		db.Select("id, request_code, workflow_id, entity_type, entity_id, status").
			Order("id DESC").Limit(10).Find(&requests)
		
		fmt.Println("\nRecent approval requests:")
		for _, req := range requests {
			fmt.Printf("ID: %d, Code: %s, WorkflowID: %d, Entity: %s(%d), Status: %s\n", 
				req.ID, req.RequestCode, req.WorkflowID, req.EntityType, req.EntityID, req.Status)
		}
		return
	}

	fmt.Printf("Found approval request 29:\n")
	fmt.Printf("  Code: %s\n", request.RequestCode)
	fmt.Printf("  Entity: %s (ID: %d)\n", request.EntityType, request.EntityID)
	fmt.Printf("  Amount: %.2f\n", request.Amount)
	fmt.Printf("  Status: %s\n", request.Status)
	fmt.Printf("  Priority: %s\n", request.Priority)
	fmt.Printf("  Workflow ID: %d\n", request.WorkflowID)

	if request.Workflow.ID != 0 {
		fmt.Printf("  Workflow: %s (%s)\n", request.Workflow.Name, request.Workflow.Module)
		fmt.Printf("  Workflow Steps:\n")
		for _, step := range request.Workflow.Steps {
			fmt.Printf("    Step %d: %s (Role: %s, Order: %d)\n", 
				step.ID, step.StepName, step.ApproverRole, step.StepOrder)
		}
	}

	fmt.Printf("  Approval Steps (Actions):\n")
	for _, action := range request.ApprovalSteps {
		approverInfo := "None"
		if action.ApproverID != nil {
			approverInfo = fmt.Sprintf("User %d", *action.ApproverID)
		}
		fmt.Printf("    Action %d: Step %d, Status: %s, Active: %t, Approver: %s\n", 
			action.ID, action.StepID, action.Status, action.IsActive, approverInfo)
		
		if action.Step.ID != 0 {
			fmt.Printf("      Step Details: %s (Role: %s, Order: %d)\n", 
				action.Step.StepName, action.Step.ApproverRole, action.Step.StepOrder)
		}
	}

	// Check for step ID 11 specifically
	var step11 models.ApprovalStep
	err = db.Where("id = ?", 11).First(&step11).Error
	if err != nil {
		fmt.Printf("\n❌ Step ID 11 not found: %v\n", err)
		
		// Show available steps
		var steps []models.ApprovalStep
		db.Select("id, workflow_id, step_name, approver_role, step_order").
			Order("id DESC").Limit(10).Find(&steps)
		
		fmt.Println("\nRecent approval steps:")
		for _, step := range steps {
			fmt.Printf("  ID: %d, WorkflowID: %d, Name: %s, Role: %s, Order: %d\n", 
				step.ID, step.WorkflowID, step.StepName, step.ApproverRole, step.StepOrder)
		}
	} else {
		fmt.Printf("\n✅ Found Step ID 11:\n")
		fmt.Printf("  Workflow ID: %d\n", step11.WorkflowID)
		fmt.Printf("  Name: %s\n", step11.StepName)
		fmt.Printf("  Role: %s\n", step11.ApproverRole)
		fmt.Printf("  Order: %d\n", step11.StepOrder)
	}

	// Check purchase 5 to see its current approval state
	var purchase models.Purchase
	err = db.Preload("ApprovalRequest").Preload("ApprovalRequest.ApprovalSteps").
		Where("id = ?", 5).First(&purchase).Error
	
	if err != nil {
		fmt.Printf("\n❌ Purchase ID 5 not found: %v\n", err)
	} else {
		fmt.Printf("\n📋 Purchase 5 Details:\n")
		fmt.Printf("  Code: %s\n", purchase.Code)
		fmt.Printf("  Status: %s\n", purchase.Status)
		fmt.Printf("  Approval Status: %s\n", purchase.ApprovalStatus)
		fmt.Printf("  Amount: %.2f\n", purchase.TotalAmount)
		
		if purchase.ApprovalRequestID != nil {
			fmt.Printf("  Approval Request ID: %d\n", *purchase.ApprovalRequestID)
		} else {
			fmt.Printf("  No approval request linked\n")
		}
	}
}
