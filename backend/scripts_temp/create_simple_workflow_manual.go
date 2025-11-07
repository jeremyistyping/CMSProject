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
	fmt.Println("🔧 Creating Simplified Purchase Approval Workflow")

	// Connect to database menggunakan environment variable atau default
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "sistem_akuntansi")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		dbHost, dbUser, dbPass, dbName, dbPort)

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("✅ Connected to database")

	// Begin transaction
	tx := db.Begin()

	// 1. Clean old workflow data
	fmt.Println("🧹 Cleaning old approval workflows...")
	
	// Clean in proper order (child first, parent last)
	tx.Exec("DELETE FROM approval_actions WHERE request_id IN (SELECT id FROM approval_requests)")
	tx.Exec("DELETE FROM approval_history WHERE request_id IN (SELECT id FROM approval_requests)")
	tx.Exec("DELETE FROM approval_requests")
	tx.Exec("DELETE FROM approval_workflow_steps WHERE workflow_id IN (SELECT id FROM approval_workflows WHERE module = 'PURCHASE')")
	tx.Exec("DELETE FROM approval_workflows WHERE module = 'PURCHASE'")

	// 2. Create single simplified workflow
	workflow := ApprovalWorkflow{
		Name:            "Standard Purchase Approval",
		Module:          "PURCHASE",
		MinAmount:       0,
		MaxAmount:       999999999999, // Very high number to cover all amounts
		RequireDirector: false,         // Director is optional, controlled by checkbox
		RequireFinance:  true,          // Finance always required
		IsActive:        true,
	}

	if err := tx.Create(&workflow).Error; err != nil {
		tx.Rollback()
		log.Fatal("Failed to create workflow:", err)
	}

	fmt.Printf("✅ Created workflow: %s (ID: %d)\n", workflow.Name, workflow.ID)

	// 3. Create workflow steps
	steps := []ApprovalStep{
		{
			WorkflowID:   workflow.ID,
			StepOrder:    1,
			StepName:     "Employee Submission",
			ApproverRole: "employee",
			IsOptional:   false,
			TimeLimit:    24, // 24 hours
		},
		{
			WorkflowID:   workflow.ID,
			StepOrder:    2,
			StepName:     "Finance Approval",
			ApproverRole: "finance",
			IsOptional:   false,
			TimeLimit:    48, // 48 hours
		},
		{
			WorkflowID:   workflow.ID,
			StepOrder:    3,
			StepName:     "Director Approval",
			ApproverRole: "director",
			IsOptional:   true, // Optional - activated via escalation checkbox
			TimeLimit:    72,   // 72 hours
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
		
		fmt.Printf("✅ Created step %d: %s - %s%s\n", 
			step.StepOrder, step.StepName, step.ApproverRole, optional)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Fatal("Failed to commit transaction:", err)
	}

	fmt.Println("\n🎉 Simplified Purchase Approval Workflow Created Successfully!")
	fmt.Println("")
	fmt.Println("📋 Workflow Details:")
	fmt.Printf("   Name: %s\n", workflow.Name)
	fmt.Printf("   Module: %s\n", workflow.Module)
	fmt.Printf("   Amount Range: %.0f - %.0f\n", workflow.MinAmount, workflow.MaxAmount)
	fmt.Println("")
	fmt.Println("🔄 Approval Process:")
	fmt.Println("   1. Employee creates & submits purchase")
	fmt.Println("   2. Finance reviews & approves")
	fmt.Println("   3. Finance can escalate to Director (via checkbox)")
	fmt.Println("   4. Director approves if escalated")
	fmt.Println("")
	fmt.Println("✨ Benefits:")
	fmt.Println("   - Simple single workflow for all amounts")
	fmt.Println("   - Manual escalation control")
	fmt.Println("   - No complex amount-based rules")
	fmt.Println("   - More flexible approval process")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}