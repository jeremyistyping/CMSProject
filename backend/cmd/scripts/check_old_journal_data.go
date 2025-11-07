package main

import (
	"fmt"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("📊 Checking Old Journal Data")
	fmt.Println("============================")

	_ = config.LoadConfig()
	db := database.ConnectDB()

	fmt.Println("✅ Database connected successfully\n")

	// Check journal_entries table
	var journalCount int64
	err := db.Raw("SELECT COUNT(*) FROM journal_entries").Scan(&journalCount).Error
	if err != nil {
		if db.Error == gorm.ErrRecordNotFound || err.Error() == "relation \"journal_entries\" does not exist" {
			fmt.Println("❌ Table 'journal_entries' does not exist")
			fmt.Println("💡 No old journal data to migrate")
		} else {
			fmt.Printf("❌ Error checking journal_entries: %v\n", err)
		}
	} else {
		fmt.Printf("📝 Found %d records in journal_entries table\n", journalCount)
		
		if journalCount > 0 {
			// Sample some records
			var samples []struct {
				ID          uint64
				Description string
				EntryDate   string
			}
			
			db.Raw(`
				SELECT id, description, TO_CHAR(entry_date, 'YYYY-MM-DD') as entry_date 
				FROM journal_entries 
				ORDER BY created_at DESC 
				LIMIT 5
			`).Scan(&samples)
			
			fmt.Println("\n📄 Sample Records:")
			for _, sample := range samples {
				fmt.Printf("   ID: %d - %s (%s)\n", sample.ID, sample.Description, sample.EntryDate)
			}
		}
	}

	// Check for journals table (if exists)
	var journalsCount int64
	err = db.Raw("SELECT COUNT(*) FROM journals").Scan(&journalsCount).Error
	if err != nil {
		if db.Error == gorm.ErrRecordNotFound || err.Error() == "relation \"journals\" does not exist" {
			fmt.Println("\n❌ Table 'journals' does not exist")
		} else {
			fmt.Printf("❌ Error checking journals: %v\n", err)
		}
	} else {
		fmt.Printf("📚 Found %d records in journals table\n", journalsCount)
	}

	fmt.Println("\n🎯 Migration Recommendation:")
	if journalCount > 0 || journalsCount > 0 {
		fmt.Printf("✅ Found %d journal entries to migrate\n", journalCount + journalsCount)
		fmt.Println("📋 Run: make migrate-ssot")
	} else {
		fmt.Println("ℹ️  No old journal data found")
		fmt.Println("💡 You can start fresh with SSOT system")
	}
}