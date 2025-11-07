package main

import (
	"fmt"
	"log"
	"time"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
)

func main() {
	// Initialize database
	db := database.ConnectDB()

	fmt.Println("🧪 Testing Approval Workflow...")
	
	// Create services
	approvalService := services.NewApprovalService(db)
	purchaseService := services.NewPurchaseService(db)
	
	// Check for any existing pending approval requests
	fmt.Println("\n1. Checking existing pending approvals...")
	var pendingRequests []models.ApprovalRequest
	err := db.Preload("ApprovalSteps").Preload("ApprovalSteps.Step").
		Where("status = ?", models.ApprovalStatusPending).
		Find(&pendingRequests).Error
	
	if err != nil {
		log.Printf("Error getting pending requests: %v", err)
	} else {
		fmt.Printf("Found %d pending approval requests:\n", len(pendingRequests))
		for _, req := range pendingRequests {
			fmt.Printf("  Request %d: %s (Entity: %s %d)\n", 
				req.ID, req.RequestCode, req.EntityType, req.EntityID)
			
			hasActiveStep := false
			for _, step := range req.ApprovalSteps {
				if step.IsActive && step.Status == models.ApprovalStatusPending {
					fmt.Printf("    Active Step: %s (Role: %s)\n", step.Step.StepName, step.Step.ApproverRole)
					hasActiveStep = true
				}
			}
			if !hasActiveStep {
				fmt.Printf("    ⚠️  No active steps found!\n")
			}
		}
	}
	
	// Create a test purchase for approval workflow testing
	fmt.Println("\n2. Creating test purchase...")
	
	// First, get a vendor
	var vendor models.Contact
	err = db.Where("type = ? AND is_active = ?", models.ContactTypeVendor, true).First(&vendor).Error
	if err != nil {
		fmt.Printf("❌ No active vendor found: %v\n", err)
		return
	}
	
	// Get a user to be the requester
	var user models.User
	err = db.Where("is_active = ?", true).First(&user).Error
	if err != nil {
		fmt.Printf("❌ No active user found: %v\n", err)
		return
	}
	
	// Create test purchase
	testPurchase := models.Purchase{
		Code:               fmt.Sprintf("TEST-PO-%d", time.Now().Unix()),
		VendorID:           vendor.ID,
		Date:               time.Now(),
		DueDate:            time.Now().AddDate(0, 0, 30),
		SubtotalAmount:     30000000, // 30M IDR - should require director approval
		TaxAmount:          3000000,
		TotalAmount:        33000000,
		Status:             models.PurchaseStatusDraft,
		ApprovalStatus:     models.PurchaseApprovalPending,
		Notes:              "Test purchase for approval workflow testing",
		CreatedBy:          user.ID,
	}
	
	err = db.Create(&testPurchase).Error
	if err != nil {
		fmt.Printf("❌ Failed to create test purchase: %v\n", err)
		return
	}
	
	fmt.Printf("✅ Created test purchase: %s (ID: %d, Amount: %.2f)\n", 
		testPurchase.Code, testPurchase.ID, testPurchase.TotalAmount)
	
	// Test approval workflow creation
	fmt.Println("\n3. Testing approval request creation...")
	
	// Submit purchase for approval
	err = purchaseService.SubmitForApproval(testPurchase.ID, user.ID)
	if err != nil {
		fmt.Printf("❌ Failed to submit for approval: %v\n", err)
		
		// Try manual approval request creation
		fmt.Println("   Trying manual approval request creation...")
		reqDTO := models.CreateApprovalRequestDTO{
			EntityType:     models.EntityTypePurchase,
			EntityID:       testPurchase.ID,
			Amount:         testPurchase.TotalAmount,
			Priority:       models.ApprovalPriorityNormal,
			RequestTitle:   fmt.Sprintf("Purchase Order %s", testPurchase.Code),
			RequestMessage: "Test purchase approval request",
		}
		
		approvalReq, err := approvalService.CreateApprovalRequest(reqDTO, user.ID)
		if err != nil {
			fmt.Printf("❌ Failed to create manual approval request: %v\n", err)
		} else {
			fmt.Printf("✅ Created approval request: %s (ID: %d)\n", 
				approvalReq.RequestCode, approvalReq.ID)
			
			// Update purchase with approval request
			testPurchase.ApprovalRequestID = &approvalReq.ID
			db.Save(&testPurchase)
		}
	} else {
		fmt.Printf("✅ Successfully submitted purchase for approval\n")
	}
	
	// Reload purchase to see current state
	err = db.Preload("ApprovalRequest").Preload("ApprovalRequest.ApprovalSteps").
		Preload("ApprovalRequest.ApprovalSteps.Step").
		First(&testPurchase, testPurchase.ID).Error
	
	if err != nil {
		fmt.Printf("❌ Failed to reload purchase: %v\n", err)
		return
	}
	
	fmt.Println("\n4. Current approval state:")
	fmt.Printf("  Purchase Status: %s\n", testPurchase.Status)
	fmt.Printf("  Approval Status: %s\n", testPurchase.ApprovalStatus)
	
	if testPurchase.ApprovalRequestID != nil {
		fmt.Printf("  Approval Request ID: %d\n", *testPurchase.ApprovalRequestID)
		fmt.Printf("  Request Status: %s\n", testPurchase.ApprovalRequest.Status)
		
		fmt.Println("  Approval Steps:")
		for _, step := range testPurchase.ApprovalRequest.ApprovalSteps {
			activeStatus := ""
			if step.IsActive {
				activeStatus = " [ACTIVE]"
			}
			fmt.Printf("    Step %d: %s (Role: %s) - Status: %s%s\n", 
				step.Step.StepOrder, step.Step.StepName, step.Step.ApproverRole, step.Status, activeStatus)
		}
	} else {
		fmt.Printf("  No approval request linked\n")
	}
	
	fmt.Println("\n✅ Test completed!")
}
