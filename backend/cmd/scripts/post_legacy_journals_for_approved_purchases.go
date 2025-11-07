package main

import (
	"fmt"
	"log"
	"os"

	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// This script backfills legacy COA journals for already APPROVED purchases so that Accounts.balance is up-to-date
// It does NOT touch SSOT; it only creates and posts legacy entries if missing.
func main() {
	dsn := getDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		log.Fatalf("failed to connect DB: %v", err)
	}

	purchaseRepo := repositories.NewPurchaseRepository(db)
	accountRepo := repositories.NewAccountRepository(db)
	journalRepo := repositories.NewJournalEntryRepository(db)

	// We only need minimal PurchaseService to call createAndPostPurchaseJournalEntries
	ps := services.NewPurchaseService(db, purchaseRepo, repositories.NewProductRepository(db), repositories.NewContactRepository(db), accountRepo, services.NewApprovalService(db), nil, journalRepo, services.NewPDFService(db), services.NewUnifiedJournalService(db))

	var ids []uint
	// Approved or Completed purchases missing legacy journal entry
	query := `
		SELECT p.id
		FROM purchases p
		LEFT JOIN journal_entries je
		  ON je.reference_type = 'PURCHASE' AND je.reference_id = p.id AND je.deleted_at IS NULL
		WHERE p.deleted_at IS NULL AND p.status IN ('APPROVED','COMPLETED') AND je.id IS NULL
	`
	if err := db.Raw(query).Scan(&ids).Error; err != nil {
		log.Fatalf("failed to scan purchases: %v", err)
	}

	if len(ids) == 0 {
		fmt.Println("✅ No approved purchases requiring legacy COA journal backfill.")
		return
	}

	fmt.Printf("Found %d approved purchases without legacy journal. Processing...\n", len(ids))
	for _, id := range ids {
		purchase, err := purchaseRepo.FindByID(id)
		if err != nil {
			fmt.Printf("❌ Skip %d: %v\n", id, err)
			continue
		}
		if _, err := purchase.PurchaseItems, error(nil); err != nil { // keep compiler happy (items already preloaded in repo)
		}
		if err := ps.CreateAndPostLegacyForCLI(purchase); err != nil {
			fmt.Printf("❌ %s: %v\n", purchase.Code, err)
			continue
		}
		fmt.Printf("📗 Posted legacy COA journal for %s\n", purchase.Code)
	}
	fmt.Println("✅ Backfill completed.")
}

func getDSN() string {
	// Use DATABASE_URL if set; otherwise rely on driver defaults (change to your local DSN if needed)
dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost/sistem_akuntansi?sslmode=disable"
	}
	return dsn
}
