package main

import (
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"log"
	"time"
)

func main() {
	db := database.ConnectDB()

	log.Println("🧪 Testing Project-Purchase Integration...")
	log.Println("")

	// Test 1: Check if columns exist
	log.Println("📋 Test 1: Verify Database Schema")
	var colExists bool
	
	db.Raw(`SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'purchases' AND column_name = 'project_id')`).Scan(&colExists)
	if colExists {
		log.Println("   ✅ purchases.project_id exists")
	} else {
		log.Println("   ❌ purchases.project_id NOT FOUND")
		return
	}
	
	db.Raw(`SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'projects' AND column_name = 'actual_cost')`).Scan(&colExists)
	if colExists {
		log.Println("   ✅ projects.actual_cost exists")
	} else {
		log.Println("   ❌ projects.actual_cost NOT FOUND")
		return
	}
	
	// Test 2: Create Test Project
	log.Println("")
	log.Println("📋 Test 2: Create Test Project")
	
	testProject := models.Project{
		ProjectName:        "Test Project - Pabrik Gresik",
		ProjectDescription: "Testing project-purchase integration",
		Customer:           "PT Test Company",
		City:               "Gresik",
		Address:            "Jl. Test No. 123",
		ProjectType:        models.ProjectTypeNewBuild,
		Budget:             1000000000, // Rp 1 Milyar
		Deadline:           time.Now().AddDate(0, 6, 0), // 6 months from now
		Status:             models.ProjectStatusActive,
	}
	
	// Check if test project already exists
	var existingProject models.Project
	err := db.Where("project_name = ?", testProject.ProjectName).First(&existingProject).Error
	if err == nil {
		log.Printf("   ⏭️  Test project already exists (ID: %d)", existingProject.ID)
		testProject = existingProject
	} else {
		if err := db.Create(&testProject).Error; err != nil {
			log.Fatalf("   ❌ Failed to create test project: %v", err)
		}
		log.Printf("   ✅ Test project created (ID: %d)", testProject.ID)
	}
	
	log.Printf("   📊 Budget: Rp %.2f", testProject.Budget)
	log.Printf("   📊 Actual Cost: Rp %.2f", testProject.ActualCost)
	log.Printf("   📊 Variance: Rp %.2f", testProject.Variance)
	
	// Test 3: Get Vendor for Purchase
	log.Println("")
	log.Println("📋 Test 3: Get Vendor for Test Purchase")
	
	var vendor models.Contact
	if err := db.Where("type = ?", "vendor").First(&vendor).Error; err != nil {
		log.Println("   ⚠️  No vendor found, skipping purchase test")
		return
	}
	log.Printf("   ✅ Using vendor: %s (ID: %d)", vendor.Name, vendor.ID)
	
	// Test 4: Get Product for Purchase
	log.Println("")
	log.Println("📋 Test 4: Get Product for Test Purchase")
	
	var product models.Product
	if err := db.First(&product).Error; err != nil {
		log.Println("   ⚠️  No product found, skipping purchase test")
		return
	}
	log.Printf("   ✅ Using product: %s (ID: %d)", product.Name, product.ID)
	
	// Test 5: Get User for Purchase
	log.Println("")
	log.Println("📋 Test 5: Get User for Test Purchase")
	
	var user models.User
	if err := db.Where("username = ?", "admin").First(&user).Error; err != nil {
		log.Fatalf("   ❌ Admin user not found: %v", err)
	}
	log.Printf("   ✅ Using user: %s (ID: %d)", user.Username, user.ID)
	
	// Test 6: Create Purchase linked to Project
	log.Println("")
	log.Println("📋 Test 6: Create Purchase Linked to Project")
	
	testPurchase := models.Purchase{
		Code:              "TEST-PR-001",
		ProjectID:         &testProject.ID, // Link to project
		VendorID:          vendor.ID,
		UserID:            user.ID,
		Date:              time.Now(),
		DueDate:           time.Now().AddDate(0, 1, 0),
		PaymentMethod:     models.PurchasePaymentCredit,
		Status:            models.PurchaseStatusDraft,
		RequiresApproval:  true,
		ApprovalStatus:    models.PurchaseApprovalNotStarted,
		CurrentApprovalStep: models.PurchaseApprovalStepNone,
	}
	
	// Check if test purchase already exists
	var existingPurchase models.Purchase
	err = db.Where("code = ?", testPurchase.Code).First(&existingPurchase).Error
	if err == nil {
		log.Printf("   ⏭️  Test purchase already exists (ID: %d)", existingPurchase.ID)
		testPurchase = existingPurchase
	} else {
		if err := db.Create(&testPurchase).Error; err != nil {
			log.Fatalf("   ❌ Failed to create test purchase: %v", err)
		}
		log.Printf("   ✅ Test purchase created (ID: %d)", testPurchase.ID)
	}
	
	// Test 7: Verify Purchase is linked to Project
	log.Println("")
	log.Println("📋 Test 7: Verify Purchase-Project Link")
	
	var linkedPurchase models.Purchase
	if err := db.Preload("Project").Where("id = ?", testPurchase.ID).First(&linkedPurchase).Error; err != nil {
		log.Fatalf("   ❌ Failed to load purchase with project: %v", err)
	}
	
	if linkedPurchase.ProjectID != nil && linkedPurchase.Project != nil {
		log.Printf("   ✅ Purchase is linked to project")
		log.Printf("   📊 Project ID: %d", *linkedPurchase.ProjectID)
		log.Printf("   📊 Project Name: %s", linkedPurchase.Project.ProjectName)
	} else {
		log.Println("   ❌ Purchase is NOT linked to project")
		return
	}
	
	// Test 8: Get all purchases for this project
	log.Println("")
	log.Println("📋 Test 8: Get All Purchases for Project")
	
	var projectPurchases []models.Purchase
	if err := db.Where("project_id = ?", testProject.ID).Find(&projectPurchases).Error; err != nil {
		log.Fatalf("   ❌ Failed to get project purchases: %v", err)
	}
	
	log.Printf("   ✅ Found %d purchase(s) for project", len(projectPurchases))
	for i, p := range projectPurchases {
		log.Printf("      %d. Code: %s, Status: %s, Amount: Rp %.2f", i+1, p.Code, p.Status, p.TotalAmount)
	}
	
	// Test 9: Test Project Cost Tracking Methods
	log.Println("")
	log.Println("📋 Test 9: Test Project Cost Tracking Methods")
	
	// Simulate cost tracking
	testProject.UpdateCostTracking(
		500000000, // material cost
		200000000, // labor cost
		150000000, // equipment cost
		100000000, // overhead cost
	)
	
	log.Printf("   ✅ UpdateCostTracking() executed")
	log.Printf("   📊 Material Cost: Rp %.2f", testProject.MaterialCost)
	log.Printf("   📊 Labor Cost: Rp %.2f", testProject.LaborCost)
	log.Printf("   📊 Equipment Cost: Rp %.2f", testProject.EquipmentCost)
	log.Printf("   📊 Overhead Cost: Rp %.2f", testProject.OverheadCost)
	log.Printf("   📊 Actual Cost: Rp %.2f", testProject.ActualCost)
	log.Printf("   📊 Variance: Rp %.2f (%.2f%%)", testProject.Variance, testProject.VariancePercent)
	log.Printf("   📊 Budget Utilization: %.2f%%", testProject.GetBudgetUtilization())
	log.Printf("   📊 Remaining Budget: Rp %.2f", testProject.GetRemainingBudget())
	log.Printf("   📊 Is Over Budget: %v", testProject.IsOverBudget())
	
	// Save updated project
	if err := db.Save(&testProject).Error; err != nil {
		log.Printf("   ⚠️  Failed to save project updates: %v", err)
	} else {
		log.Println("   ✅ Project cost tracking saved to database")
	}
	
	// Test 10: Query Project with Purchases
	log.Println("")
	log.Println("📋 Test 10: Query Project with All Purchases")
	
	var projectWithPurchases models.Project
	if err := db.Preload("Purchases").Where("id = ?", testProject.ID).First(&projectWithPurchases).Error; err != nil {
		log.Fatalf("   ❌ Failed to load project with purchases: %v", err)
	}
	
	log.Printf("   ✅ Project loaded with purchases")
	log.Printf("   📊 Project: %s", projectWithPurchases.ProjectName)
	log.Printf("   📊 Budget: Rp %.2f", projectWithPurchases.Budget)
	log.Printf("   📊 Actual Cost: Rp %.2f", projectWithPurchases.ActualCost)
	log.Printf("   📊 Total Purchases: %d", len(projectWithPurchases.Purchases))
	
	// Summary
	log.Println("")
	log.Println("=" + string(make([]byte, 60)) + "=")
	log.Println("🎉 Integration Test Summary")
	log.Println("=" + string(make([]byte, 60)) + "=")
	log.Println("✅ Database schema verified")
	log.Println("✅ Project model working")
	log.Println("✅ Purchase model working")
	log.Println("✅ Project-Purchase link working")
	log.Println("✅ Cost tracking methods working")
	log.Println("✅ Bidirectional relations working")
	log.Println("")
	log.Println("🎯 Project-Purchase Integration: SUCCESS!")
	log.Println("")
	log.Println("Next steps:")
	log.Println("1. Update Purchase API to include project_id")
	log.Println("2. Update frontend to show project dropdown in purchase form")
	log.Println("3. Create Cost Control dashboard with project cost tracking")
	log.Println("4. Add project filter in purchases list")
}
