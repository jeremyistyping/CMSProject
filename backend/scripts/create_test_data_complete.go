package main

import (
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"log"
	"time"
)

func main() {
	db := database.ConnectDB()

	log.Println("🔄 Creating complete test data for integration...")
	
	// 1. Create Vendor
	log.Println("\n📋 Creating Test Vendor...")
	var vendor models.Contact
	err := db.Where("email = ?", "vendor@test.com").First(&vendor).Error
	if err == nil {
		log.Printf("   ⏭️  Vendor exists (ID: %d)", vendor.ID)
	} else {
		vendor = models.Contact{
			Type:    "vendor",
			Name:    "PT Supplier Material Bangunan",
			Email:   "vendor@test.com",
			Phone:   "021-1234567",
			Address: "Jl. Industri No. 1, Jakarta",
		}
		if err := db.Create(&vendor).Error; err != nil {
			log.Fatalf("   ❌ Failed: %v", err)
		}
		log.Printf("   ✅ Vendor created (ID: %d)", vendor.ID)
	}
	
	// 2. Create Product Category
	log.Println("\n📋 Creating Product Category...")
	var category models.ProductCategory
	err = db.Where("name = ?", "Material Konstruksi").First(&category).Error
	if err == nil {
		log.Printf("   ⏭️  Category exists (ID: %d)", category.ID)
	} else {
		category = models.ProductCategory{
			Code: "MAT001",
			Name: "Material Konstruksi",
			Description: "Material untuk konstruksi bangunan",
		}
		if err := db.Create(&category).Error; err != nil {
			log.Fatalf("   ❌ Failed: %v", err)
		}
		log.Printf("   ✅ Category created (ID: %d)", category.ID)
	}
	
	// 3. Create Products
	log.Println("\n📋 Creating Test Products...")
	products := []models.Product{
		{
			Code: "MAT-001",
			Name: "Semen Portland",
			Description: "Semen Portland 50kg per sak",
			CategoryID: &category.ID,
			Unit: "sak",
			PurchasePrice: 65000,
			SalePrice: 70000,
			Stock: 1000,
		},
		{
			Code: "MAT-002",
			Name: "Besi Beton 12mm",
			Description: "Besi beton diameter 12mm panjang 12m",
			CategoryID: &category.ID,
			Unit: "batang",
			PurchasePrice: 95000,
			SalePrice: 105000,
			Stock: 500,
		},
		{
			Code: "MAT-003",
			Name: "Pasir Cor",
			Description: "Pasir cor kualitas premium",
			CategoryID: &category.ID,
			Unit: "m3",
			PurchasePrice: 250000,
			SalePrice: 280000,
			Stock: 100,
		},
	}
	
	for _, p := range products {
		var existing models.Product
		err := db.Where("code = ?", p.Code).First(&existing).Error
		if err == nil {
			log.Printf("   ⏭️  Product %s exists (ID: %d)", p.Code, existing.ID)
		} else {
			if err := db.Create(&p).Error; err != nil {
				log.Printf("   ⚠️  Failed to create %s: %v", p.Code, err)
			} else {
				log.Printf("   ✅ Product created: %s (ID: %d)", p.Code, p.ID)
			}
		}
	}
	
	// 4. Create Projects
	log.Println("\n📋 Creating Test Projects...")
	projects := []models.Project{
		{
			ProjectName: "Proyek Pabrik Gresik",
			ProjectDescription: "Pembangunan pabrik di Gresik",
			Customer: "PT Manufacturing Indonesia",
			City: "Gresik",
			Address: "Kawasan Industri Gresik",
			ProjectType: models.ProjectTypeNewBuild,
			Budget: 5000000000, // Rp 5 Milyar
			Deadline: time.Now().AddDate(1, 0, 0), // 1 year
			Status: models.ProjectStatusActive,
		},
		{
			ProjectName: "Renovasi Gedung Kantor",
			ProjectDescription: "Renovasi gedung kantor pusat",
			Customer: "PT Office Solutions",
			City: "Jakarta",
			Address: "Jl. Sudirman No. 123",
			ProjectType: models.ProjectTypeRenovation,
			Budget: 2000000000, // Rp 2 Milyar
			Deadline: time.Now().AddDate(0, 6, 0), // 6 months
			Status: models.ProjectStatusActive,
		},
	}
	
	for _, proj := range projects {
		var existing models.Project
		err := db.Where("project_name = ?", proj.ProjectName).First(&existing).Error
		if err == nil {
			log.Printf("   ⏭️  Project exists: %s (ID: %d)", proj.ProjectName, existing.ID)
		} else {
			if err := db.Create(&proj).Error; err != nil {
				log.Printf("   ⚠️  Failed to create %s: %v", proj.ProjectName, err)
			} else {
				log.Printf("   ✅ Project created: %s (ID: %d)", proj.ProjectName, proj.ID)
			}
		}
	}
	
	log.Println("\n✅ Test data creation completed!")
	log.Println("\nTest data summary:")
	
	// Count data
	var vendorCount, productCount, projectCount int64
	db.Model(&models.Contact{}).Where("type = ?", "vendor").Count(&vendorCount)
	db.Model(&models.Product{}).Count(&productCount)
	db.Model(&models.Project{}).Count(&projectCount)
	
	log.Printf("   Vendors: %d", vendorCount)
	log.Printf("   Products: %d", productCount)
	log.Printf("   Projects: %d", projectCount)
	
	log.Println("\nYou can now run the integration test:")
	log.Println("   go run scripts/test_project_purchase_integration.go")
}
