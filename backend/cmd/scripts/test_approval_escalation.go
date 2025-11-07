package main

import (
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"fmt"
	"log"
	"time"
)

func main() {
	// Initialize database
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	// Initialize services
	approvalService := services.NewApprovalService(db)
	
	fmt.Println("=== Testing Approval Escalation to Director ===")
	fmt.Println()
	
	// Step 1: Check for active Directors
	fmt.Println("Step 1: Checking for active Directors...")
	var directors []models.User
	err := db.Where("LOWER(role) = LOWER(?) AND is_active = ?", "director", true).Find(&directors).Error
	if err != nil {
		log.Fatal("Error finding directors:", err)
	}
	
	if len(directors) == 0 {
		fmt.Println("❌ No active directors found!")
		fmt.Println("Creating a test director...")
		
		// Create test director
		director := models.User{
			Username:  "test.director",
			Email:     "director@test.com",
			FirstName: "Test",
			LastName:  "Director",
			Role:      "director",
			IsActive:  true,
			Password:  "$2a$10$YourHashedPasswordHere", // Use proper hashed password in production
		}
		
		if err := db.Create(&director).Error; err != nil {
			log.Fatal("Failed to create test director:", err)
		}
		fmt.Printf("✅ Created test director: %s (ID: %d)\n", director.Username, director.ID)
		directors = append(directors, director)
	} else {
		fmt.Printf("✅ Found %d active director(s):\n", len(directors))
		for _, d := range directors {
			fmt.Printf("   - %s (ID: %d, Email: %s)\n", d.Username, d.ID, d.Email)
		}
	}
	fmt.Println()
	
	// Step 2: Find a pending purchase with approval request
	fmt.Println("Step 2: Finding a pending purchase for testing...")
	var purchase models.Purchase
	err = db.Preload("ApprovalRequest").
		Where("approval_status = ? AND approval_request_id IS NOT NULL", "PENDING").
		First(&purchase).Error
	
	if err != nil {
		fmt.Println("❌ No pending purchase with approval request found")
		fmt.Println("Please create a purchase and submit it for approval first")
		return
	}
	
	fmt.Printf("✅ Found purchase: %s (ID: %d, Amount: %.2f)\n", purchase.Code, purchase.ID, purchase.TotalAmount)
	fmt.Printf("   Approval Request ID: %d\n", *purchase.ApprovalRequestID)
	fmt.Println()
	
	// Step 3: Get Finance user
	fmt.Println("Step 3: Finding Finance user...")
	var financeUser models.User
	err = db.Where("LOWER(role) = LOWER(?) AND is_active = ?", "finance", true).First(&financeUser).Error
	if err != nil {
		fmt.Println("❌ No active finance user found")
		return
	}
	fmt.Printf("✅ Found Finance user: %s (ID: %d)\n", financeUser.Username, financeUser.ID)
	fmt.Println()
	
	// Step 4: Simulate escalation to Director
	fmt.Println("Step 4: Escalating to Director...")
	escalationReason := "Purchase amount exceeds Finance approval limit - requires Director approval"
	
	err = approvalService.EscalateToDirector(*purchase.ApprovalRequestID, financeUser.ID, escalationReason)
	if err != nil {
		fmt.Printf("❌ Failed to escalate: %v\n", err)
		return
	}
	
	fmt.Println("✅ Successfully escalated to Director!")
	fmt.Println()
	
	// Step 5: Check notifications created
	fmt.Println("Step 5: Checking notifications...")
	var notifications []models.Notification
	
	// Get notifications for all directors
	for _, director := range directors {
		var directorNotifs []models.Notification
		err = db.Where("user_id = ? AND type = ? AND created_at >= ?", 
			director.ID, 
			"approval_pending", 
			time.Now().Add(-5*time.Minute)).
			Order("created_at DESC").
			Find(&directorNotifs).Error
		
		if err != nil {
			fmt.Printf("❌ Error checking notifications for director %d: %v\n", director.ID, err)
			continue
		}
		
		if len(directorNotifs) > 0 {
			fmt.Printf("✅ Director %s has %d new notification(s):\n", director.Username, len(directorNotifs))
			for _, n := range directorNotifs {
				fmt.Printf("   - %s: %s\n", n.Title, n.Message)
				fmt.Printf("     Priority: %s, Created: %s\n", n.Priority, n.CreatedAt.Format("15:04:05"))
			}
		} else {
			fmt.Printf("⚠️  No notifications found for director %s\n", director.Username)
		}
		
		notifications = append(notifications, directorNotifs...)
	}
	fmt.Println()
	
	// Step 6: Check approval request status
	fmt.Println("Step 6: Checking approval request status...")
	var approvalRequest models.ApprovalRequest
	err = db.Preload("ApprovalSteps.Step").
		First(&approvalRequest, *purchase.ApprovalRequestID).Error
	
	if err != nil {
		fmt.Printf("❌ Failed to get approval request: %v\n", err)
		return
	}
	
	fmt.Printf("Approval Request Status: %s\n", approvalRequest.Status)
	fmt.Printf("Priority: %s\n", approvalRequest.Priority)
	fmt.Println("\nActive Steps:")
	
	for _, step := range approvalRequest.ApprovalSteps {
		if step.IsActive {
			fmt.Printf("✅ Active: %s (Role: %s, Status: %s)\n", 
				step.Step.StepName, 
				step.Step.ApproverRole, 
				step.Status)
		}
	}
	fmt.Println()
	
	// Step 7: Check approval history
	fmt.Println("Step 7: Checking approval history...")
	var history []models.ApprovalHistory
	err = db.Preload("User").
		Where("request_id = ?", *purchase.ApprovalRequestID).
		Order("created_at DESC").
		Limit(5).
		Find(&history).Error
	
	if err == nil && len(history) > 0 {
		fmt.Println("Recent approval history:")
		for _, h := range history {
			fmt.Printf("   - %s by %s: %s\n", 
				h.Action, 
				h.User.Username, 
				h.Comments)
		}
	}
	fmt.Println()
	
	// Summary
	fmt.Println("=== Test Summary ===")
	if len(notifications) > 0 {
		fmt.Printf("✅ Escalation successful!\n")
		fmt.Printf("   - Purchase escalated to Director\n")
		fmt.Printf("   - %d notification(s) created\n", len(notifications))
		fmt.Printf("   - Director can now approve/reject the purchase\n")
	} else {
		fmt.Printf("⚠️  Escalation completed but no notifications found\n")
		fmt.Printf("   Please check the notification service configuration\n")
	}
}
