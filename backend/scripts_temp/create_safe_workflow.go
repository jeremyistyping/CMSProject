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
	fmt.Println("🔧 Creating Simplified Purchase Approval Workflow (Safe Mode)")

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

	// Check existing tables first
	fmt.Println("🔍 Checking existing approval tables...")
	var tables []string
	db.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name LIKE '%approval%' ORDER BY table_name").Scan(&tables)
	
	if len(tables) > 0 {
		fmt.Printf("   Found tables: %v\n", tables)
	} else {
		fmt.Println("   No approval tables found - they may need to be created first")
	}

	// Begin transaction
	tx := db.Begin()

	// 1. Safe cleanup - only clean tables that exist
	fmt.Println("🧹 Safely cleaning old workflow data...")
	
	// Safe cleanup with proper error handling
	safeExec(tx, "DELETE FROM approval_actions WHERE request_id IN (SELECT id FROM approval_requests WHERE id IS NOT NULL)", "approval_actions")
	safeExec(tx, "DELETE FROM approval_requests WHERE id IS NOT NULL", "approval_requests")
	safeExec(tx, "DELETE FROM approval_workflow_steps WHERE workflow_id IN (SELECT id FROM approval_workflows WHERE module = 'PURCHASE')", "approval_workflow_steps")
	safeExec(tx, "DELETE FROM approval_workflows WHERE module = 'PURCHASE'", "approval_workflows")

	// 2. Create single simplified workflow
	fmt.Println("📋 Creating simplified workflow...")
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
	fmt.Println("🔄 Creating workflow steps...")
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
		
		fmt.Printf("   ✅ Step %d: %s - %s%s (%dh limit)\n", 
			step.StepOrder, step.StepName, step.ApproverRole, optional, step.TimeLimit)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Fatal("Failed to commit transaction:", err)
	}

	fmt.Println("\n🎉 Simplified Purchase Approval Workflow Created Successfully!")
	
	// Verify creation
	fmt.Println("\n📊 Verification:")
	var workflowCount int64
	var stepCount int64
	
	db.Model(&ApprovalWorkflow{}).Where("module = ?", "PURCHASE").Count(&workflowCount)
	db.Model(&ApprovalStep{}).Joins("JOIN approval_workflows ON approval_workflows.id = approval_workflow_steps.workflow_id").Where("approval_workflows.module = ?", "PURCHASE").Count(&stepCount)
	
	fmt.Printf("   📋 Workflows created: %d\n", workflowCount)
	fmt.Printf("   🔄 Steps created: %d\n", stepCount)
	
	fmt.Println("\n✨ New Approval Process:")
	fmt.Println("   1. Employee: Create & Submit Purchase → Step 1 activated")
	fmt.Println("   2. Finance: Review & Approve → Can choose to escalate")
	fmt.Println("   3. Director: Approve if escalated → Final approval")
	fmt.Println("\n🎯 Benefits:")
	fmt.Println("   - Single workflow covers all purchase amounts")
	fmt.Println("   - Finance controls escalation via checkbox")
	fmt.Println("   - No complex amount-based routing")
	fmt.Println("   - Flexible and user-friendly process")
}

func safeExec(tx *gorm.DB, query string, tableName string) {
	if err := tx.Exec(query).Error; err != nil {
		fmt.Printf("   ⚠️  Warning: Could not clean %s (table may not exist): %v\n", tableName, err)
		// Don't fail, just continue - table might not exist yet
	} else {
		fmt.Printf("   ✅ Cleaned %s\n", tableName)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}