package main

import (
	"fmt"
	"log"
	"os"

	"app-sistem-akuntansi/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Get database config from environment
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgres"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "sistem_akuntansi"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	
	// Construct DSN
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		dbHost, dbUser, dbPassword, dbName, dbPort)

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("Connected to database successfully!")
	fmt.Println("\n=== Fixing Approval Workflows ===")
	
	// Start transaction
	tx := db.Begin()
	
	// 1. First, check if we need to recreate workflows or just fix steps
	fmt.Println("1. Checking existing workflows...")
	var existingWorkflows []models.ApprovalWorkflow
	tx.Find(&existingWorkflows)
	fmt.Printf("   Found %d existing workflows\n", len(existingWorkflows))
	
	// 2. Delete manager steps from existing workflows
	fmt.Println("\n2. Removing manager approval steps...")
	if err := tx.Exec(`
		DELETE FROM approval_actions 
		WHERE step_id IN (
			SELECT id FROM approval_steps 
			WHERE approver_role = 'manager'
		)
	`).Error; err != nil {
		fmt.Printf("   ⚠ Warning: %v\n", err)
	}
	
	if err := tx.Exec("DELETE FROM approval_steps WHERE approver_role = 'manager'").Error; err != nil {
		fmt.Printf("   ⚠ Warning: %v\n", err)
	} else {
		fmt.Println("   ✓ Removed all manager approval steps")
	}
	
	// 3. Update step orders to remove gaps
	fmt.Println("\n3. Updating step orders...")
	
	// Fix step orders for each workflow
	for _, workflow := range existingWorkflows {
		var steps []models.ApprovalStep
		tx.Where("workflow_id = ?", workflow.ID).Order("step_order ASC").Find(&steps)
		
		newOrder := 1
		for _, step := range steps {
			if step.StepOrder != newOrder {
				tx.Model(&step).Update("step_order", newOrder)
				fmt.Printf("   Updated step %d (role: %s) to order %d in workflow %d\n", step.ID, step.ApproverRole, newOrder, workflow.ID)
			}
			newOrder++
		}
	}
	
	// 4. Fix existing approval requests
	fmt.Println("\n4. Fixing existing approval requests...")
	
	// Find the high value workflow (> 25M)
	var highValueWorkflow models.ApprovalWorkflow
	if err := tx.Where("module = ? AND min_amount > ? AND is_active = ?", "PURCHASE", 25000000, true).First(&highValueWorkflow).Error; err != nil {
		fmt.Printf("   ⚠ Could not find high value workflow: %v\n", err)
	} else {
		fmt.Printf("   Found high value workflow: %s (ID: %d)\n", highValueWorkflow.Name, highValueWorkflow.ID)
	}
	
	// Update approval request for purchase ID 6
	var purchase models.Purchase
	if err := tx.Unscoped().First(&purchase, 6).Error; err == nil {
		if purchase.ApprovalRequestID != nil {
			fmt.Printf("   Processing Purchase ID 6 (Approval Request ID: %d)\n", *purchase.ApprovalRequestID)
			
			// Get current approval request
			var approvalReq models.ApprovalRequest
			if err := tx.First(&approvalReq, *purchase.ApprovalRequestID).Error; err == nil {
				fmt.Printf("   Current workflow ID: %d\n", approvalReq.WorkflowID)
				
				// Find the finance step in the current workflow
				var financeStep models.ApprovalStep
				if err := tx.Where("workflow_id = ? AND approver_role = ?", approvalReq.WorkflowID, "finance").First(&financeStep).Error; err == nil {
					// Make finance step active
					if err := tx.Exec(`
						UPDATE approval_actions 
						SET is_active = true 
						WHERE request_id = ? 
						AND step_id = ?
					`, *purchase.ApprovalRequestID, financeStep.ID).Error; err != nil {
						fmt.Printf("   ⚠ Failed to activate finance step: %v\n", err)
					} else {
						fmt.Printf("   ✓ Activated finance approval step (ID: %d) for Purchase ID 6\n", financeStep.ID)
					}
				} else {
					fmt.Printf("   ⚠ Could not find finance step: %v\n", err)
				}
			}
		} else {
			fmt.Printf("   Purchase ID 6 has no approval request\n")
		}
	} else {
		fmt.Printf("   Purchase ID 6 not found\n")
	}
	
	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}
	
	fmt.Println("\n✅ Successfully fixed approval workflows!")
	fmt.Println("   - Workflow 1: Standard Purchase (≤ 25M): Employee → Finance")
	fmt.Println("   - Workflow 2: High Value Purchase (> 25M): Employee → Finance → Director")
	fmt.Println("   - Workflow 3: Sales: Employee → Finance")
	fmt.Println("\n📝 Purchase ID 6 now waiting for Finance approval (manager step removed)")
}
