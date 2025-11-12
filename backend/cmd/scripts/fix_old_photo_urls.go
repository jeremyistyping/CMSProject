package main

import (
	"log"
	"os"
	"strings"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"

	"github.com/lib/pq"
)

// This script fixes old photo URLs in the database that have Windows backslashes
// and missing leading slashes

func main() {
	log.Println("🔧 Starting Photo URL Fix Script...")
	log.Println("=" + strings.Repeat("=", 60))

	// Load config
	cfg := config.LoadConfig()
	_ = cfg // We just need it for DB connection

	// Connect to database
	db := database.ConnectDB()

	// Get all daily updates with photos
	var dailyUpdates []models.DailyUpdate
	result := db.Where("photos IS NOT NULL AND array_length(photos, 1) > 0").Find(&dailyUpdates)

	if result.Error != nil {
		log.Fatalf("❌ Failed to fetch daily updates: %v", result.Error)
	}

	log.Printf("📊 Found %d daily updates with photos\n", len(dailyUpdates))
	log.Println()

	fixedCount := 0
	skippedCount := 0

	for _, update := range dailyUpdates {
		needsUpdate := false
		newPhotos := make([]string, len(update.Photos))

		log.Printf("🔍 Checking Daily Update ID: %d", update.ID)
		log.Printf("   Project ID: %d", update.ProjectID)
		log.Printf("   Date: %s", update.Date.Format("2006-01-02"))
		log.Printf("   Photos count: %d", len(update.Photos))

		for i, photo := range update.Photos {
			log.Printf("   Photo %d (original): %s", i+1, photo)

			// Check if photo needs fixing
			if strings.Contains(photo, "\\") || !strings.HasPrefix(photo, "/") {
				// Fix the photo URL
				fixed := fixPhotoURL(photo)
				newPhotos[i] = fixed
				needsUpdate = true
				log.Printf("   Photo %d (fixed):    %s ✅", i+1, fixed)

				// Check if file exists
				checkFileExists(fixed)
			} else {
				// Already in correct format
				newPhotos[i] = photo
				log.Printf("   Photo %d (OK):       %s ✓", i+1, photo)
				checkFileExists(photo)
			}
		}

		if needsUpdate {
			// Update the database
			update.Photos = pq.StringArray(newPhotos)
			if err := db.Save(&update).Error; err != nil {
				log.Printf("   ❌ Failed to update: %v\n", err)
			} else {
				log.Printf("   ✅ Updated successfully!\n")
				fixedCount++
			}
		} else {
			log.Printf("   ℹ️  No changes needed\n")
			skippedCount++
		}
		log.Println()
	}

	log.Println(strings.Repeat("=", 60))
	log.Printf("✅ Fix completed!")
	log.Printf("   Fixed: %d updates", fixedCount)
	log.Printf("   Skipped: %d updates", skippedCount)
	log.Printf("   Total: %d updates", len(dailyUpdates))
}

func fixPhotoURL(photo string) string {
	// 1. Convert Windows backslashes to forward slashes
	fixed := strings.ReplaceAll(photo, "\\", "/")

	// 2. Remove ./uploads prefix if present
	fixed = strings.Replace(fixed, "./uploads", "/uploads", 1)

	// 3. Ensure leading slash
	if !strings.HasPrefix(fixed, "/") {
		if strings.HasPrefix(fixed, "uploads") {
			fixed = "/" + fixed
		}
	}

	return fixed
}

func checkFileExists(photoURL string) {
	// Remove leading /uploads to get file path
	filePath := strings.TrimPrefix(photoURL, "/uploads")
	filePath = "./uploads" + filePath

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("      ⚠️  WARNING: File not found on disk: %s", filePath)
	} else {
		log.Printf("      ✓ File exists on disk: %s", filePath)
	}
}

