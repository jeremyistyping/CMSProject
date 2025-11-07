package main

import (
	"context"
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
)

// backfill_ssot_for_approved_purchases.go
// Purpose: Ensure all APPROVED purchases have SSOT journals (POSTED) so
//          /api/v1/coa/posted-balances reflects correct balances.
// Safe: Read-mostly; creates missing SSOT journals using existing adapters.

func main() {
	log.Println("🚀 Backfill SSOT journals for APPROVED purchases (posted-only)")

	// Connect DB
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	// Dependencies
	purchaseRepo := repositories.NewPurchaseRepository(db)
	accountRepo := repositories.NewAccountRepository(db)
	unified := services.NewUnifiedJournalService(db)
	// Tax account service for flexible account mapping
	taxSvc := services.NewTaxAccountService(db)
	adapter := services.NewPurchaseSSOTJournalAdapter(db, unified, accountRepo, taxSvc)

	// Find APPROVED purchases that don't have SSOT PURCHASE journals yet
	var purchaseIDs []uint
	query := `
		SELECT p.id
		FROM purchases p
		WHERE p.status = 'APPROVED'
		  AND p.approval_status = 'APPROVED'
		  AND p.deleted_at IS NULL
		  AND NOT EXISTS (
		    SELECT 1
		    FROM unified_journal_ledger uj
		    WHERE uj.deleted_at IS NULL
		      AND uj.source_type = 'PURCHASE'
		      AND uj.source_id = p.id
		  )
		ORDER BY p.id
	`
	if err := db.Raw(query).Scan(&purchaseIDs).Error; err != nil {
		log.Fatalf("Failed to query approved purchases without SSOT journals: %v", err)
	}

	if len(purchaseIDs) == 0 {
		log.Println("✅ No missing SSOT journals for approved purchases. Nothing to backfill.")
		return
	}

	log.Printf("Found %d APPROVED purchases missing SSOT journals. Processing...", len(purchaseIDs))

	ctx := context.Background()
	var success, failed int

	for _, id := range purchaseIDs {
		purchase, err := purchaseRepo.FindByID(id)
		if err != nil {
			failed++
			log.Printf("❌ Purchase %d load failed: %v", id, err)
			continue
		}

		// Validate minimal mapping early (common source of failures)
		if err := ensureMinimumAccounts(accountRepo); err != nil {
			failed++
			log.Printf("❌ Account mapping not ready for purchase %s: %v", purchase.Code, err)
			continue
		}

		// Use creator as poster for audit; fallback to userID=1
		userID := uint64(purchase.UserID)
		if userID == 0 {
			userID = 1
		}

		entry, err := adapter.CreatePurchaseJournalEntry(ctx, purchase, userID)
		if err != nil {
			failed++
			log.Printf("❌ Failed to create SSOT journal for %s: %v", purchase.Code, err)
			continue
		}

		// Validate POSTED status (adapter uses AutoPost=true)
		if entry.Status != models.SSOTStatusPosted {
			failed++
			log.Printf("⚠️ Journal created but not POSTED for %s (status=%s)", purchase.Code, entry.Status)
			continue
		}

		success++
		log.Printf("✅ SSOT POSTED journal created for %s (JE=%s, ID=%d)", purchase.Code, entry.EntryNumber, entry.ID)
	}

	log.Printf("\n📊 Backfill summary: success=%d, failed=%d, total=%d", success, failed, len(purchaseIDs))
	if failed > 0 {
		log.Println("Some records failed. Check logs above for details (account mapping, missing tables, etc.)")
	}
	log.Println("Done.")
}

// ensureMinimumAccounts checks existence of critical accounts used by purchase journals.
func ensureMinimumAccounts(accountRepo repositories.AccountRepository) error {
	// Verify 1301 Inventory, 1240 PPN Masukan, 2101 Hutang Usaha exist
	required := []string{"1301", "1240", "2101"}
	missing := make([]string, 0)
	for _, code := range required {
		if _, err := accountRepo.GetAccountByCode(code); err != nil {
			missing = append(missing, code)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required accounts: %v", missing)
	}
	return nil
}
