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

	fmt.Println("🔄 Renaming existing workflow to test auto-migration...")
	fmt.Println("✅ Connected to database")

	// Find Standard Purchase Approval workflow
	var workflow ApprovalWorkflow
	result := db.Where("name = ? AND module = ?", "Standard Purchase Approval", "PURCHASE").First(&workflow)
	
	if result.Error != nil {
		fmt.Println("❌ Standard Purchase Approval workflow not found - nothing to rename")
		return
	}

	fmt.Printf("📋 Found Standard Purchase Approval workflow (ID: %d)\n", workflow.ID)

	// Rename it to test auto-migration
	newName := "OLD - Standard Purchase Approval (Renamed for Test)"
	result = db.Model(&workflow).Update("name", newName)
	if result.Error != nil {
		log.Fatal("❌ Error renaming workflow:", result.Error)
	}

	fmt.Printf("✅ Renamed workflow to: %s\n", newName)

	// Verify
	var count int64
	db.Where("name = ? AND module = ?", "Standard Purchase Approval", "PURCHASE").Model(&ApprovalWorkflow{}).Count(&count)
	fmt.Printf("📊 Workflows with name 'Standard Purchase Approval': %d\n", count)

	fmt.Println("🎯 Workflow renamed successfully!")
	fmt.Println("💡 Now start the backend to test auto-migration - it should create a new 'Standard Purchase Approval' workflow")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}