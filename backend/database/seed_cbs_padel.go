package database

import (
	"app-sistem-akuntansi/models"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// SeedCBSPadelBandung seeds CBS structure for Padel Bandung project
func SeedCBSPadelBandung(db *gorm.DB) error {
	log.Println("🏗️  Starting CBS seeding for Padel Bandung project...")

	// Find Padel Bandung project
	var project models.Project
	if err := db.Where("project_name = ?", "Padel Bandung").First(&project).Error; err != nil {
		return fmt.Errorf("project 'Padel Bandung' not found: %v", err)
	}

	log.Printf("✅ Found project: %s (ID: %d)\n", project.ProjectName, project.ID)

	// Check if CBS already exists for this project
	var existingCount int64
	db.Model(&models.CBSNode{}).Where("project_id = ?", project.ID).Count(&existingCount)
	if existingCount > 0 {
		log.Printf("⚠️  CBS nodes already exist for Padel Bandung (%d nodes). Skipping...\n", existingCount)
		return nil
	}

	// Get some COA accounts for mapping (optional)
	var coaAccounts []models.COAAccount
	db.Where("type = ? AND is_active = ?", "EXPENSE", true).Limit(10).Find(&coaAccounts)

	now := time.Now()

	// Define CBS structure for Padel Court Construction
	cbsNodes := []models.CBSNode{
		// Level 1: Main Categories
		{
			ProjectID:    project.ID,
			Code:         "1.0",
			Name:         "Site Preparation & Foundation",
			Description:  "Persiapan lahan dan pekerjaan pondasi",
			BudgetAmount: 250000000, // 250 juta
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ProjectID:    project.ID,
			Code:         "2.0",
			Name:         "Court Construction",
			Description:  "Konstruksi lapangan padel",
			BudgetAmount: 800000000, // 800 juta
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ProjectID:    project.ID,
			Code:         "3.0",
			Name:         "Facilities & Amenities",
			Description:  "Fasilitas pendukung dan amenitas",
			BudgetAmount: 350000000, // 350 juta
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ProjectID:    project.ID,
			Code:         "4.0",
			Name:         "Equipment & Furnishing",
			Description:  "Peralatan dan perlengkapan",
			BudgetAmount: 200000000, // 200 juta
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ProjectID:    project.ID,
			Code:         "5.0",
			Name:         "Utilities & Systems",
			Description:  "Utilitas dan sistem pendukung",
			BudgetAmount: 150000000, // 150 juta
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}

	// Create Level 1 nodes
	for i := range cbsNodes {
		if err := db.Create(&cbsNodes[i]).Error; err != nil {
			return fmt.Errorf("failed to create CBS node %s: %v", cbsNodes[i].Code, err)
		}
		log.Printf("✅ Created: %s - %s\n", cbsNodes[i].Code, cbsNodes[i].Name)
	}

	// Level 2: Sub-categories for Site Preparation (1.0)
	sitePrep := cbsNodes[0]
	sitePrepChildren := []models.CBSNode{
		{
			ProjectID:    project.ID,
			ParentID:     &sitePrep.ID,
			Code:         "1.1",
			Name:         "Land Clearing & Grading",
			Description:  "Pembersihan dan perataan lahan",
			BudgetAmount: 50000000,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ProjectID:    project.ID,
			ParentID:     &sitePrep.ID,
			Code:         "1.2",
			Name:         "Excavation & Earthwork",
			Description:  "Penggalian dan pekerjaan tanah",
			BudgetAmount: 80000000,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ProjectID:    project.ID,
			ParentID:     &sitePrep.ID,
			Code:         "1.3",
			Name:         "Foundation Work",
			Description:  "Pekerjaan pondasi",
			BudgetAmount: 120000000,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}

	for i := range sitePrepChildren {
		if err := db.Create(&sitePrepChildren[i]).Error; err != nil {
			return fmt.Errorf("failed to create CBS node %s: %v", sitePrepChildren[i].Code, err)
		}
		log.Printf("  ✅ Created: %s - %s\n", sitePrepChildren[i].Code, sitePrepChildren[i].Name)
	}

	// Level 2: Sub-categories for Court Construction (2.0)
	courtConst := cbsNodes[1]
	courtChildren := []models.CBSNode{
		{
			ProjectID:    project.ID,
			ParentID:     &courtConst.ID,
			Code:         "2.1",
			Name:         "Court Surface & Base",
			Description:  "Permukaan dan dasar lapangan",
			BudgetAmount: 300000000,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ProjectID:    project.ID,
			ParentID:     &courtConst.ID,
			Code:         "2.2",
			Name:         "Glass Walls",
			Description:  "Dinding kaca lapangan",
			BudgetAmount: 250000000,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ProjectID:    project.ID,
			ParentID:     &courtConst.ID,
			Code:         "2.3",
			Name:         "Metal Structure & Fencing",
			Description:  "Struktur metal dan pagar",
			BudgetAmount: 150000000,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ProjectID:    project.ID,
			ParentID:     &courtConst.ID,
			Code:         "2.4",
			Name:         "Court Lighting",
			Description:  "Pencahayaan lapangan",
			BudgetAmount: 100000000,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}

	for i := range courtChildren {
		if err := db.Create(&courtChildren[i]).Error; err != nil {
			return fmt.Errorf("failed to create CBS node %s: %v", courtChildren[i].Code, err)
		}
		log.Printf("  ✅ Created: %s - %s\n", courtChildren[i].Code, courtChildren[i].Name)
	}

	// Level 2: Sub-categories for Facilities (3.0)
	facilities := cbsNodes[2]
	facilitiesChildren := []models.CBSNode{
		{
			ProjectID:    project.ID,
			ParentID:     &facilities.ID,
			Code:         "3.1",
			Name:         "Clubhouse Building",
			Description:  "Bangunan clubhouse",
			BudgetAmount: 150000000,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ProjectID:    project.ID,
			ParentID:     &facilities.ID,
			Code:         "3.2",
			Name:         "Locker Rooms & Showers",
			Description:  "Ruang ganti dan shower",
			BudgetAmount: 80000000,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ProjectID:    project.ID,
			ParentID:     &facilities.ID,
			Code:         "3.3",
			Name:         "Cafe & Lounge Area",
			Description:  "Area kafe dan lounge",
			BudgetAmount: 70000000,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ProjectID:    project.ID,
			ParentID:     &facilities.ID,
			Code:         "3.4",
			Name:         "Parking Area",
			Description:  "Area parkir",
			BudgetAmount: 50000000,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}

	for i := range facilitiesChildren {
		if err := db.Create(&facilitiesChildren[i]).Error; err != nil {
			return fmt.Errorf("failed to create CBS node %s: %v", facilitiesChildren[i].Code, err)
		}
		log.Printf("  ✅ Created: %s - %s\n", facilitiesChildren[i].Code, facilitiesChildren[i].Name)
	}

	// Level 2: Sub-categories for Equipment (4.0)
	equipment := cbsNodes[3]
	equipmentChildren := []models.CBSNode{
		{
			ProjectID:    project.ID,
			ParentID:     &equipment.ID,
			Code:         "4.1",
			Name:         "Sports Equipment",
			Description:  "Peralatan olahraga (raket, bola, dll)",
			BudgetAmount: 50000000,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ProjectID:    project.ID,
			ParentID:     &equipment.ID,
			Code:         "4.2",
			Name:         "Furniture & Fixtures",
			Description:  "Furniture dan perlengkapan",
			BudgetAmount: 80000000,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ProjectID:    project.ID,
			ParentID:     &equipment.ID,
			Code:         "4.3",
			Name:         "Audio Visual System",
			Description:  "Sistem audio visual",
			BudgetAmount: 40000000,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ProjectID:    project.ID,
			ParentID:     &equipment.ID,
			Code:         "4.4",
			Name:         "Security System",
			Description:  "Sistem keamanan (CCTV, access control)",
			BudgetAmount: 30000000,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}

	for i := range equipmentChildren {
		if err := db.Create(&equipmentChildren[i]).Error; err != nil {
			return fmt.Errorf("failed to create CBS node %s: %v", equipmentChildren[i].Code, err)
		}
		log.Printf("  ✅ Created: %s - %s\n", equipmentChildren[i].Code, equipmentChildren[i].Name)
	}

	// Level 2: Sub-categories for Utilities (5.0)
	utilities := cbsNodes[4]
	utilitiesChildren := []models.CBSNode{
		{
			ProjectID:    project.ID,
			ParentID:     &utilities.ID,
			Code:         "5.1",
			Name:         "Electrical System",
			Description:  "Sistem kelistrikan",
			BudgetAmount: 60000000,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ProjectID:    project.ID,
			ParentID:     &utilities.ID,
			Code:         "5.2",
			Name:         "Plumbing & Water System",
			Description:  "Sistem perpipaan dan air",
			BudgetAmount: 40000000,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ProjectID:    project.ID,
			ParentID:     &utilities.ID,
			Code:         "5.3",
			Name:         "HVAC System",
			Description:  "Sistem AC dan ventilasi",
			BudgetAmount: 35000000,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ProjectID:    project.ID,
			ParentID:     &utilities.ID,
			Code:         "5.4",
			Name:         "Internet & Network",
			Description:  "Sistem internet dan jaringan",
			BudgetAmount: 15000000,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}

	for i := range utilitiesChildren {
		if err := db.Create(&utilitiesChildren[i]).Error; err != nil {
			return fmt.Errorf("failed to create CBS node %s: %v", utilitiesChildren[i].Code, err)
		}
		log.Printf("  ✅ Created: %s - %s\n", utilitiesChildren[i].Code, utilitiesChildren[i].Name)
	}

	// Count total nodes created
	var totalNodes int64
	db.Model(&models.CBSNode{}).Where("project_id = ?", project.ID).Count(&totalNodes)

	log.Printf("\n🎉 Successfully created %d CBS nodes for Padel Bandung project!\n", totalNodes)
	log.Println("📊 CBS Structure Summary:")
	log.Println("   - Site Preparation & Foundation: 3 sub-items")
	log.Println("   - Court Construction: 4 sub-items")
	log.Println("   - Facilities & Amenities: 4 sub-items")
	log.Println("   - Equipment & Furnishing: 4 sub-items")
	log.Println("   - Utilities & Systems: 4 sub-items")
	log.Printf("   Total Budget: Rp %.0f\n", float64(250000000+800000000+350000000+200000000+150000000))

	return nil
}
