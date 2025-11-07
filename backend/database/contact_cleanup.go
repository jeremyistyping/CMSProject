package database

import (
	"log"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
)

// CleanupDuplicateContacts removes duplicate contacts keeping only the first one
func CleanupDuplicateContacts(db *gorm.DB) error {
	log.Println("Starting contact duplicate cleanup...")

	// Find all contacts grouped by code
	var contacts []models.Contact
	if err := db.Find(&contacts).Error; err != nil {
		return err
	}

	// Group contacts by code
	contactsByCode := make(map[string][]models.Contact)
	for _, contact := range contacts {
		contactsByCode[contact.Code] = append(contactsByCode[contact.Code], contact)
	}

	// Remove duplicates (keep the oldest one based on ID)
	for code, contactList := range contactsByCode {
		if len(contactList) > 1 {
			log.Printf("Found %d contacts with code %s", len(contactList), code)
			
			// Sort by ID to keep the first created
			for i := 1; i < len(contactList); i++ {
				log.Printf("Deleting duplicate contact: ID=%d, Name=%s, Code=%s", 
					contactList[i].ID, contactList[i].Name, contactList[i].Code)
				
				// Hard delete to avoid constraint issues
				if err := db.Unscoped().Delete(&contactList[i]).Error; err != nil {
					log.Printf("Error deleting contact %d: %v", contactList[i].ID, err)
				}
			}
		}
	}

	log.Println("Contact duplicate cleanup completed")
	return nil
}

// CheckContactCodeConflicts shows any existing code conflicts
func CheckContactCodeConflicts(db *gorm.DB) {
	log.Println("Checking for contact code conflicts...")

	var results []struct {
		Code  string
		Count int64
	}

	db.Model(&models.Contact{}).
		Select("code, count(*) as count").
		Group("code").
		Having("count(*) > 1").
		Find(&results)

	if len(results) == 0 {
		log.Println("No contact code conflicts found")
		return
	}

	log.Printf("Found %d contact code conflicts:", len(results))
	for _, result := range results {
		log.Printf("- Code '%s' appears %d times", result.Code, result.Count)
		
		// Show details of conflicting contacts
		var contacts []models.Contact
		db.Where("code = ?", result.Code).Find(&contacts)
		for _, contact := range contacts {
			log.Printf("  - ID: %d, Name: %s, Type: %s", contact.ID, contact.Name, contact.Type)
		}
	}
}
