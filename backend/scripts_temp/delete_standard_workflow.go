package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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

	fmt.Println("🗑️  Deleting Standard Purchase Approval workflow for testing...")
	fmt.Println("✅ Connected to database")

	// Find Standard Purchase Approval workflow
	var workflow ApprovalWorkflow
	result := db.Where("name = ? AND module = ?", "Standard Purchase Approval", "PURCHASE").First(&workflow)
	
	if result.Error != nil {
		fmt.Println("❌ Standard Purchase Approval workflow not found")
		return
	}

	fmt.Printf("📋 Found Standard Purchase Approval workflow (ID: %d)\n", workflow.ID)

	// Delete all approval_requests that reference this workflow (for testing purposes)
	fmt.Println("🔄 Deleting approval_requests that reference this workflow...")
	result = db.Exec("DELETE FROM approval_requests WHERE workflow_id = ?", workflow.ID)
	if result.Error != nil {
		log.Printf("⚠️  Warning: Error deleting approval_requests: %v", result.Error)
	} else {
		fmt.Printf("   ✅ Deleted %d approval_request(s)\n", result.RowsAffected)
	}
	
	// Also delete any approval_actions that might reference the steps
	fmt.Println("🔄 Deleting approval_actions for this workflow...")
	result = db.Exec("DELETE FROM approval_actions WHERE step_id IN (SELECT id FROM approval_steps WHERE workflow_id = ?)", workflow.ID)
	if result.Error != nil {
		log.Printf("⚠️  Warning: Error deleting approval_actions: %v", result.Error)
	} else {
		fmt.Printf("   ✅ Deleted %d approval_action(s)\n", result.RowsAffected)
	}

	// Delete workflow steps
	fmt.Println("🔄 Deleting workflow steps...")
	result = db.Where("workflow_id = ?", workflow.ID).Delete(&ApprovalStep{})
	if result.Error != nil {
		log.Printf("⚠️  Warning: Error deleting workflow steps: %v", result.Error)
	} else {
		fmt.Printf("   ✅ Deleted %d workflow steps\n", result.RowsAffected)
	}

	// Delete workflow
	fmt.Println("🔄 Deleting workflow...")
	result = db.Delete(&workflow)
	if result.Error != nil {
		log.Fatal("❌ Error deleting workflow:", result.Error)
	}
	fmt.Printf("✅ Deleted Standard Purchase Approval workflow\n")

	// Verify deletion
	var count int64
	db.Where("module = ?", "PURCHASE").Model(&ApprovalWorkflow{}).Count(&count)
	fmt.Printf("📊 Remaining PURCHASE workflows: %d\n", count)

	fmt.Println("🎯 Standard Purchase Approval workflow deleted successfully!")
	fmt.Println("💡 Now restart the backend to test auto-migration")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}