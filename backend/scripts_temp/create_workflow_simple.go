package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Simple structs untuk workflow creation
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

type ApprovalStep struct {
	ID           uint   `gorm:"primaryKey"`
	WorkflowID   uint   `gorm:"not null;index"`
	StepOrder    int    `gorm:"not null"`
	StepName     string `gorm:"not null;size:100"`
	ApproverRole string `gorm:"not null;size:50"`
	IsOptional   bool   `gorm:"default:false"`
	TimeLimit    int    `gorm:"default:24"`
}

func main() {
	fmt.Println("🔧 Creating Purchase Approval Workflow (Simple Mode)")

	// Connect to database
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "sistem_akuntansi")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		dbHost, dbUser, dbPass, dbName, dbPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("✅ Connected to database")

	// Check if workflow already exists
	var existingCount int64
	db.Model(&ApprovalWorkflow{}).Where("module = ? AND name = ?", "PURCHASE", "Standard Purchase Approval").Count(&existingCount)
	
	if existingCount > 0 {
		fmt.Println("⚠️  Workflow already exists, skipping creation")
		return
	}

	// Begin transaction
	tx := db.Begin()

	// Create workflow
	fmt.Println("📋 Creating workflow...")
	workflow := ApprovalWorkflow{
		Name:            "Standard Purchase Approval",
		Module:          "PURCHASE",
		MinAmount:       0,
		MaxAmount:       999999999999,
		RequireDirector: false,
		RequireFinance:  true,
		IsActive:        true,
	}

	if err := tx.Create(&workflow).Error; err != nil {
		tx.Rollback()
		log.Fatal("Failed to create workflow:", err)
	}

	fmt.Printf("✅ Created workflow: %s (ID: %d)\n", workflow.Name, workflow.ID)

	// Create workflow steps
	fmt.Println("🔄 Creating workflow steps...")
	steps := []ApprovalStep{
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
			TimeLimit:    48,
		},
		{
			WorkflowID:   workflow.ID,
			StepOrder:    3,
			StepName:     "Director Approval",
			ApproverRole: "director",
			IsOptional:   true,
			TimeLimit:    72,
		},
	}

	for _, step := range steps {
		if err := tx.Create(&step).Error; err != nil {
			tx.Rollback()
			log.Fatal("Failed to create step:", err)
		}
		
		optional := ""
		if step.IsOptional {
			optional = " (Optional)"
		}
		
		fmt.Printf("   ✅ Step %d: %s - %s%s\n", 
			step.StepOrder, step.StepName, step.ApproverRole, optional)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Fatal("Failed to commit transaction:", err)
	}

	fmt.Println("\n🎉 Workflow Created Successfully!")
	
	// Verify
	var workflowCount, stepCount int64
	db.Model(&ApprovalWorkflow{}).Where("module = ?", "PURCHASE").Count(&workflowCount)
	db.Table("approval_steps").Joins("JOIN approval_workflows ON approval_workflows.id = approval_steps.workflow_id").Where("approval_workflows.module = ?", "PURCHASE").Count(&stepCount)
	
	fmt.Printf("\n📊 Verification: %d workflow(s), %d step(s)\n", workflowCount, stepCount)
	
	fmt.Println("\n✨ Simplified Approval Process:")
	fmt.Println("   1. Employee: Submit Purchase")
	fmt.Println("   2. Finance: Approve + Optional Escalate")
	fmt.Println("   3. Director: Final Approval (if escalated)")
	fmt.Println("\nNow restart the backend server and test!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}