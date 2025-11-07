package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ApprovalWorkflow represents the approval_workflows table for auto-migration
type ApprovalWorkflow struct {
	ID              uint    `gorm:"primaryKey"`
	Name            string  `gorm:"not null;size:100"`
	Module          string  `gorm:"not null;size:50"`
	MinAmount       float64 `gorm:"type:decimal(15,2);default:0"`
	MaxAmount       float64 `gorm:"type:decimal(15,2)"`
	IsActive        bool    `gorm:"default:true"`
	RequireDirector bool    `gorm:"default:false"`
	RequireFinance  bool    `gorm:"default:false"`
}

// ApprovalStep represents the approval_steps table for auto-migration
type ApprovalStep struct {
	ID           uint   `gorm:"primaryKey"`
	WorkflowID   uint   `gorm:"not null;index"`
	StepOrder    int    `gorm:"not null"`
	StepName     string `gorm:"not null;size:100"`
	ApproverRole string `gorm:"not null;size:50"`
	IsOptional   bool   `gorm:"default:false"`
	TimeLimit    int    `gorm:"default:24"`
}

// ensureStandardPurchaseApprovalWorkflow checks and creates Standard Purchase Approval workflow if it doesn't exist
func ensureStandardPurchaseApprovalWorkflow(db *gorm.DB) error {
	log.Println("🔍 Checking Standard Purchase Approval workflow...")
	
	// Check if Standard Purchase Approval workflow exists
	var existingWorkflow ApprovalWorkflow
	result := db.Where("name = ? AND module = ?", "Standard Purchase Approval", "PURCHASE").First(&existingWorkflow)
	
	if result.Error == nil {
		log.Println("✅ Standard Purchase Approval workflow already exists")
		return nil
	}
	
	if result.Error == nil {
		log.Println("✅ Standard Purchase Approval workflow found")
		
		// Check if workflow has steps
		var stepCount int64
		db.Model(&ApprovalStep{}).Where("workflow_id = ?", existingWorkflow.ID).Count(&stepCount)
		
		if stepCount == 0 {
			log.Println("⚠️  Workflow exists but has no steps - creating steps...")
			// Create steps for existing workflow
			return createWorkflowSteps(db, existingWorkflow.ID)
		} else {
			log.Printf("✅ Workflow has %d steps - no action needed", stepCount)
			return nil
		}
	}
	
	// If not found, create it
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Println("📝 Creating Standard Purchase Approval workflow...")
		
		// Create workflow
		workflow := ApprovalWorkflow{
			Name:            "Standard Purchase Approval",
			Module:          "PURCHASE",
			MinAmount:       0,
			MaxAmount:       999999999999,
			IsActive:        true,
			RequireDirector: true,
			RequireFinance:  true,
		}
		
		if err := db.Create(&workflow).Error; err != nil {
			return fmt.Errorf("failed to create Standard Purchase Approval workflow: %v", err)
		}
		
		log.Printf("✅ Created Standard Purchase Approval workflow with ID: %d", workflow.ID)
		
		// Create workflow steps
		return createWorkflowSteps(db, workflow.ID)
	}
	
	// Other database errors
	return fmt.Errorf("failed to check existing workflow: %v", result.Error)
}

func main() {
	// Database connection using environment variables
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "sistem_akuntansi")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		dbHost, dbUser, dbPass, dbName, dbPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}

	fmt.Println("🧪 Testing auto-migration for Standard Purchase Approval workflow...")
	fmt.Println("✅ Connected to database")

	// Test the auto-migration function
	if err := ensureStandardPurchaseApprovalWorkflow(db); err != nil {
		log.Fatal("❌ Auto-migration failed:", err)
	}

	// Verify result
	var workflow ApprovalWorkflow
	result := db.Where("name = ? AND module = ?", "Standard Purchase Approval", "PURCHASE").First(&workflow)
	if result.Error != nil {
		log.Fatal("❌ Verification failed:", result.Error)
	}

	fmt.Printf("✅ Verification passed - workflow exists with ID: %d\n", workflow.ID)

	// Show steps
	var steps []ApprovalStep
	db.Where("workflow_id = ?", workflow.ID).Order("step_order").Find(&steps)
	fmt.Printf("📋 Created %d workflow steps:\n", len(steps))
	for _, step := range steps {
		optional := ""
		if step.IsOptional {
			optional = " (Optional)"
		}
		fmt.Printf("   Step %d: %s - %s%s (%dh)\n", 
			step.StepOrder, step.StepName, step.ApproverRole, optional, step.TimeLimit)
	}

	fmt.Println("🎯 Auto-migration test completed successfully!")
}

// createWorkflowSteps creates the standard approval workflow steps
func createWorkflowSteps(db *gorm.DB, workflowID uint) error {
	steps := []ApprovalStep{
		{
			WorkflowID:   workflowID,
			StepOrder:    1,
			StepName:     "Employee Submission",
			ApproverRole: "employee",
			IsOptional:   false,
			TimeLimit:    24,
		},
		{
			WorkflowID:   workflowID,
			StepOrder:    2,
			StepName:     "Finance Approval",
			ApproverRole: "finance",
			IsOptional:   false,
			TimeLimit:    48,
		},
		{
			WorkflowID:   workflowID,
			StepOrder:    3,
			StepName:     "Director Approval",
			ApproverRole: "director",
			IsOptional:   true,
			TimeLimit:    72,
		},
	}
	
	for _, step := range steps {
		if err := db.Create(&step).Error; err != nil {
			return fmt.Errorf("failed to create workflow step '%s': %v", step.StepName, err)
		}
	}
	
	log.Printf("✅ Created %d workflow steps for Standard Purchase Approval", len(steps))
	log.Println("🎯 Standard Purchase Approval workflow setup completed!")
	
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
