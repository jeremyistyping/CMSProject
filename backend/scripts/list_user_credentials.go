package main

import (
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"log"
)

func main() {
	db := database.ConnectDB()

	log.Println("👥 List of User Credentials for Testing")
	log.Println("=" + string(make([]byte, 70)) + "=")
	log.Println("")

	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		log.Fatalf("Failed to fetch users: %v", err)
	}

	log.Println("📋 PURCHASE APPROVAL FLOW:")
	log.Println("   Flow: Purchasing → Cost Control → GM → Project Director → Managing Director")
	log.Println("")
	log.Println("Available Users:")
	log.Println("")

	for i, user := range users {
		log.Printf("%d. Role: %-20s | Username: %-15s | Email: %-30s", 
			i+1, user.Role, user.Username, user.Email)
		log.Printf("   Name: %s %s", user.FirstName, user.LastName)
		log.Printf("   Password: password123")
		log.Println("   ---")
	}

	log.Println("")
	log.Println("🔐 LOGIN CREDENTIALS FOR PURCHASE APPROVAL:")
	log.Println("")
	
	// Find specific roles
	var purchasing, costControl, gm, projectDirector, managingDirector models.User
	
	// Purchasing (Andi)
	if err := db.Where("role = ?", "purchasing").First(&purchasing).Error; err == nil {
		log.Println("1️⃣  PURCHASING (Create PR):")
		log.Printf("   Email: %s", purchasing.Email)
		log.Println("   Password: password123")
		log.Printf("   Role: %s", purchasing.Role)
		log.Println("   Can: Create Purchase Request & input kebutuhan material")
		log.Println("")
	} else {
		log.Println("1️⃣  PURCHASING user with role 'purchasing' not found")
	}
	
	// Cost Control (Patrick)
	if err := db.Where("role = ?", "cost_control").First(&costControl).Error; err == nil {
		log.Println("2️⃣  COST CONTROL (Approve Step 1):")
		log.Printf("   Email: %s", costControl.Email)
		log.Println("   Password: password123")
		log.Printf("   Role: %s", costControl.Role)
		log.Println("   Can: Review & Approve Purchase (Step 1)")
		log.Println("")
	} else {
		log.Println("2️⃣  COST CONTROL: NOT FOUND")
		log.Println("   Run: go run scripts/create_cost_control_user.go")
		log.Println("")
	}
	
	// GM (Pak Marlin)
	if err := db.Where("role = ?", "gm").First(&gm).Error; err == nil {
		log.Println("3️⃣  GM (Approve Step 2):")
		log.Printf("   Email: %s", gm.Email)
		log.Println("   Password: password123")
		log.Printf("   Role: %s", gm.Role)
		log.Println("   Can: Approve setelah Cost Control")
		log.Println("")
	} else {
		log.Println("3️⃣  GM user with role 'gm' not found")
	}
	
	// Project Director (Pak Christopher)
	if err := db.Where("role = ?", "project_director").First(&projectDirector).Error; err == nil {
		log.Println("4️⃣  PROJECT DIRECTOR (Approve Step 3):")
		log.Printf("   Email: %s", projectDirector.Email)
		log.Println("   Password: password123")
		log.Printf("   Role: %s", projectDirector.Role)
		log.Println("   Can: Approve setelah GM")
		log.Println("")
	} else {
		log.Println("4️⃣  PROJECT DIRECTOR user with role 'project_director' not found")
	}
	
	// Managing Director (Pak Jason)
	if err := db.Where("role = ?", "managing_director").First(&managingDirector).Error; err == nil {
		log.Println("5️⃣  MANAGING DIRECTOR (Final Approval Step 4):")
		log.Printf("   Email: %s", managingDirector.Email)
		log.Println("   Password: password123")
		log.Printf("   Role: %s", managingDirector.Role)
		log.Println("   Can: Final Approval")
		log.Println("")
	} else {
		log.Println("5️⃣  MANAGING DIRECTOR user with role 'managing_director' not found")
	}
	
	// Admin (for testing)
	var admin models.User
	if err := db.Where("role = ?", "admin").First(&admin).Error; err == nil {
		log.Println("4️⃣  ADMIN (Full Access):")
		log.Printf("   Email: %s", admin.Email)
		log.Println("   Password: password123")
		log.Printf("   Role: %s", admin.Role)
		log.Println("   Can: Everything")
		log.Println("")
	}

	log.Println("=" + string(make([]byte, 70)) + "=")
	log.Println("")
	log.Println("📝 APPROVAL WORKFLOW:")
	log.Println("   1. Login as PURCHASING → Create Purchase Request")
	log.Println("   2. Login as COST_CONTROL → Approve Purchase (Step 1)")
	log.Println("   3. Login as GM → Approve Purchase (Step 2)")
	log.Println("   4. Login as PROJECT_DIRECTOR → Approve Purchase (Step 3)")
	log.Println("   5. Login as MANAGING_DIRECTOR → Final Approve (Step 4)")
	log.Println("   4. Purchase Status: APPROVED ✅")
	log.Println("")
	log.Println("🌐 Frontend URL: http://localhost:3000")
	log.Println("   - Purchase List: http://localhost:3000/purchases")
	log.Println("   - Cost Control Dashboard: http://localhost:3000/cost-control")
}
