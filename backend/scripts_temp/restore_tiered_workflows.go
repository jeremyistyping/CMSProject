package main

import (
	"fmt"
	"log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
)

func main() {
	// Database connection
	dsn := "host=localhost user=postgres password=Indihome01 dbname=app_sistem_akuntansi port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("🔄 Starting workflow restoration process...")

	// Step 1: Deactivate the current broad workflow if it exists
	var currentWorkflow models.ApprovalWorkflow
	if err := db.Where("name = ? AND module = ?", "Standard Purchase Approval", models.ApprovalModulePurchase).First(&currentWorkflow).Error; err == nil {
		log.Printf("📝 Found current workflow: %s (ID: %d)", currentWorkflow.Name, currentWorkflow.ID)
		
		// Rename and deactivate it
		currentWorkflow.Name = "OLD - Standard Purchase Approval (Replaced by Tiered)"
		currentWorkflow.IsActive = false
		if err := db.Save(&currentWorkflow).Error; err != nil {
			log.Printf("⚠️ Warning: Could not deactivate old workflow: %v", err)
		} else {
			log.Printf("✅ Deactivated old workflow: %s", currentWorkflow.Name)
		}
	}

	// Step 2: Remove existing tiered workflows if any (to prevent duplicates)
	log.Println("🧹 Cleaning up any existing tiered workflows...")
	db.Where("module = ? AND name IN ?", models.ApprovalModulePurchase, []string{
		"PO <= 25M",
		"PO > 25M - 100M", 
		"PO > 100M",
	}).Delete(&models.ApprovalWorkflow{})

	// Step 3: Create tiered workflows based on original seed logic
	log.Println("🏗️ Creating tiered approval workflows...")

	// Tier A: 0 - 25,000,000 (Finance only)
	wfA := models.ApprovalWorkflow{
		Name:            "PO <= 25M",
		Module:          models.ApprovalModulePurchase,
		MinAmount:       0,
		MaxAmount:       25000000,
		IsActive:        true,
		RequireFinance:  true,
		RequireDirector: false,
	}
	if err := db.Create(&wfA).Error; err != nil {
		log.Printf("❌ Error creating Tier A workflow: %v", err)
		return
	}
	log.Printf("✅ Created Tier A workflow: %s (ID: %d)", wfA.Name, wfA.ID)

	stepA1 := models.ApprovalStep{
		WorkflowID:   wfA.ID,
		StepOrder:    1,
		StepName:     "Finance Approval",
		ApproverRole: "finance",
		TimeoutHours: 48,
		IsOptional:   false,
	}
	if err := db.Create(&stepA1).Error; err != nil {
		log.Printf("❌ Error creating Tier A step: %v", err)
	} else {
		log.Printf("✅ Created step for Tier A: %s", stepA1.StepName)
	}

	// Tier B: >25,000,000 - 100,000,000 (Finance -> Director)
	wfB := models.ApprovalWorkflow{
		Name:            "PO > 25M - 100M",
		Module:          models.ApprovalModulePurchase,
		MinAmount:       25000001, // Just above 25M
		MaxAmount:       100000000,
		IsActive:        true,
		RequireFinance:  true,
		RequireDirector: true,
	}
	if err := db.Create(&wfB).Error; err != nil {
		log.Printf("❌ Error creating Tier B workflow: %v", err)
		return
	}
	log.Printf("✅ Created Tier B workflow: %s (ID: %d)", wfB.Name, wfB.ID)

	stepB1 := models.ApprovalStep{
		WorkflowID:   wfB.ID,
		StepOrder:    1,
		StepName:     "Finance Approval",
		ApproverRole: "finance",
		TimeoutHours: 48,
		IsOptional:   false,
	}
	stepB2 := models.ApprovalStep{
		WorkflowID:   wfB.ID,
		StepOrder:    2,
		StepName:     "Director Approval",
		ApproverRole: "director",
		TimeoutHours: 72,
		IsOptional:   false,
	}
	if err := db.Create(&stepB1).Error; err != nil {
		log.Printf("❌ Error creating Tier B step 1: %v", err)
	} else {
		log.Printf("✅ Created step 1 for Tier B: %s", stepB1.StepName)
	}
	if err := db.Create(&stepB2).Error; err != nil {
		log.Printf("❌ Error creating Tier B step 2: %v", err)
	} else {
		log.Printf("✅ Created step 2 for Tier B: %s", stepB2.StepName)
	}

	// Tier C: >100,000,000 (Finance -> Director)
	wfC := models.ApprovalWorkflow{
		Name:            "PO > 100M",
		Module:          models.ApprovalModulePurchase,
		MinAmount:       100000001, // Just above 100M
		MaxAmount:       0,          // no upper bound
		IsActive:        true,
		RequireFinance:  true,
		RequireDirector: true,
	}
	if err := db.Create(&wfC).Error; err != nil {
		log.Printf("❌ Error creating Tier C workflow: %v", err)
		return
	}
	log.Printf("✅ Created Tier C workflow: %s (ID: %d)", wfC.Name, wfC.ID)

	stepC1 := models.ApprovalStep{
		WorkflowID:   wfC.ID,
		StepOrder:    1,
		StepName:     "Finance Approval",
		ApproverRole: "finance",
		TimeoutHours: 48,
		IsOptional:   false,
	}
	stepC2 := models.ApprovalStep{
		WorkflowID:   wfC.ID,
		StepOrder:    2,
		StepName:     "Director Approval",
		ApproverRole: "director",
		TimeoutHours: 72,
		IsOptional:   false,
	}
	if err := db.Create(&stepC1).Error; err != nil {
		log.Printf("❌ Error creating Tier C step 1: %v", err)
	} else {
		log.Printf("✅ Created step 1 for Tier C: %s", stepC1.StepName)
	}
	if err := db.Create(&stepC2).Error; err != nil {
		log.Printf("❌ Error creating Tier C step 2: %v", err)
	} else {
		log.Printf("✅ Created step 2 for Tier C: %s", stepC2.StepName)
	}

	// Step 4: Verify the created workflows
	log.Println("🔍 Verifying created tiered workflows...")
	
	var workflows []models.ApprovalWorkflow
	db.Where("module = ? AND is_active = ?", models.ApprovalModulePurchase, true).
	   Order("min_amount ASC").Find(&workflows)

	log.Println("📊 Active PURCHASE workflows:")
	for _, wf := range workflows {
		var stepCount int64
		db.Model(&models.ApprovalStep{}).Where("workflow_id = ?", wf.ID).Count(&stepCount)
		
		maxAmountStr := "∞"
		if wf.MaxAmount > 0 {
			maxAmountStr = formatAmount(wf.MaxAmount)
		}
		
		log.Printf("   🔧 %s (ID: %d)", wf.Name, wf.ID)
		log.Printf("      📊 Amount Range: %s - %s", formatAmount(wf.MinAmount), maxAmountStr)
		log.Printf("      👥 Finance: %t, Director: %t", wf.RequireFinance, wf.RequireDirector)
		log.Printf("      📋 Steps: %d", stepCount)
	}

	log.Println("🎉 Tiered workflow restoration completed successfully!")
}

func formatAmount(amount int64) string {
	if amount == 0 {
		return "0"
	}
	return fmt.Sprintf("%.0fM", float64(amount)/1000000)
}
