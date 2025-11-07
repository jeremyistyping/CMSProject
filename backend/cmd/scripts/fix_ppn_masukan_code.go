package main

import (
	"fmt"
	"log"
	"os"

	"app-sistem-akuntansi/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// This one-off script normalizes PPN Masukan account code from 2102 -> 1240.
// Safe rules:
// - If 1240 already exists, leave codes as-is and just mark 2102 inactive (optional arg).
// - If 1240 does NOT exist and 2102 exists, update 2102.code to 1240 and set parent to CURRENT ASSETS (1100).
// - Never touches journal lines. This is only to align master COA mapping.
func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost/sistem_akuntansi?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect DB: %v", err)
	}

	var acc2102 models.Account
	var acc1240 models.Account

	err = db.Where("code = ?", "2102").First(&acc2102).Error
	if err != nil {
		fmt.Println("✅ No 2102 account found. Nothing to do.")
		return
	}

	err1240 := db.Where("code = ?", "1240").First(&acc1240).Error
	if err1240 == nil {
		// 1240 already exists. Optionally deactivate 2102 to avoid confusion.
		fmt.Printf("ℹ️ Found both 2102 (id=%d) and 1240 (id=%d). Keeping 1240 as active PPN Masukan.\n", acc2102.ID, acc1240.ID)
		// Deactivate 2102 only if explicitly requested
		if os.Getenv("DEACTIVATE_2102") == "true" {
			if err := db.Model(&acc2102).Updates(map[string]interface{}{
				"is_active": false,
				"description": gorm.Expr("COALESCE(description,'') || ' [deprecated: replaced by 1240]'"),
			}).Error; err != nil {
				log.Fatalf("failed to deactivate 2102: %v", err)
			}
			fmt.Println("✅ 2102 deactivated (set DEACTIVATE_2102=false to skip)")
		}
		return
	}

	// 1240 not found: rename 2102 -> 1240
	fmt.Printf("🔧 Renaming account 2102 (id=%d) to 1240...\n", acc2102.ID)

	// Ensure parent 1100 exists
	var parent1100 models.Account
	if err := db.Where("code = ?", "1100").First(&parent1100).Error; err != nil {
		log.Fatalf("required parent account 1100 not found: %v", err)
	}

	updates := map[string]interface{}{
		"code":      "1240",
		"name":      "PPN Masukan",
		"type":      models.AccountTypeAsset,
		"category":  models.CategoryCurrentAsset,
		"parent_id": parent1100.ID,
		"level":     parent1100.Level + 1,
	}
	if err := db.Model(&acc2102).Updates(updates).Error; err != nil {
		log.Fatalf("failed to update account 2102->1240: %v", err)
	}

	fmt.Println("✅ Updated 2102 to 1240 successfully.")
}
