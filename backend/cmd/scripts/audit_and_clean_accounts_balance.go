package main

import (
	"fmt"
	"log"
	"os"

	"app-sistem-akuntansi/database"
	"gorm.io/gorm"
)

// This script audits accounts.balance against SSOT posted-only movement.
// If CLEAN=1 is set in env, it will reset suspicious balances to 0.
// Usage:
//   go run ./cmd/scripts/audit_and_clean_accounts_balance.go         # dry-run (no changes)
//   CLEAN=1 go run ./cmd/scripts/audit_and_clean_accounts_balance.go # apply fixes

func main() {
	clean := os.Getenv("CLEAN") == "1"

	db := database.ConnectDB()
	if db == nil {
		log.Fatal("failed to connect DB")
	}

	if err := runAudit(db, clean); err != nil {
		log.Fatalf("audit failed: %v", err)
	}
}

type acc struct {
	ID      uint
	Code    string
	Name    string
	Type    string
	Balance float64
}

type sums struct {
	Debit  float64
	Credit float64
}

func runAudit(db *gorm.DB, clean bool) error {
	log.Printf("Starting accounts.balance audit (clean=%v)...", clean)

	var accounts []acc
	if err := db.Raw(`SELECT id, code, name, type, balance FROM accounts WHERE deleted_at IS NULL ORDER BY code`).Scan(&accounts).Error; err != nil {
		return err
	}

	var suspicious []acc
	for _, a := range accounts {
		var s sums
		qry := `
			SELECT COALESCE(SUM(ujl.debit_amount),0) AS debit,
			       COALESCE(SUM(ujl.credit_amount),0) AS credit
			FROM unified_journal_lines ujl
			JOIN unified_journal_ledger ujd ON ujd.id = ujl.journal_id
			LEFT JOIN sales s ON ujd.source_type = 'SALE' AND ujd.source_id = s.id
			WHERE ujl.account_id = ?
			  AND ujd.status = 'POSTED'
			  AND ujd.deleted_at IS NULL
			  AND ((ujd.source_type = 'SALE' AND s.status IN ('INVOICED','PAID')) OR (ujd.source_type <> 'SALE'))
		`
		if err := db.Raw(qry, a.ID).Scan(&s).Error; err != nil {
			return err
		}

		// Compute ssot net
		var ssotNet float64
		if a.Type == "ASSET" || a.Type == "EXPENSE" {
			ssotNet = s.Debit - s.Credit
		} else {
			ssotNet = s.Credit - s.Debit
		}

		if (s.Debit == 0 && s.Credit == 0) && a.Balance != 0 {
			// No SSOT movement but balance non-zero => suspicious legacy residue
			suspicious = append(suspicious, a)
			if clean {
				if err := db.Exec(`UPDATE accounts SET balance = 0, updated_at = NOW() WHERE id = ?`, a.ID).Error; err != nil {
					log.Printf("Failed to reset balance for %s - %s: %v", a.Code, a.Name, err)
				} else {
					log.Printf("Reset balance for %s - %s (was %.2f)", a.Code, a.Name, a.Balance)
				}
			}
			continue
		}

		// If SSOT net differs from stored balance by a large margin, consider fixing (optional)
		delta := ssotNet - a.Balance
		if clean && (delta != 0) {
			// Align stored balance to SSOT net for leaf accounts
			if err := db.Exec(`UPDATE accounts SET balance = ?, updated_at = NOW() WHERE id = ?`, ssotNet, a.ID).Error; err != nil {
				log.Printf("Failed to align balance for %s - %s: %v", a.Code, a.Name, err)
			} else {
				log.Printf("Aligned balance for %s - %s to SSOT net (%.2f -> %.2f)", a.Code, a.Name, a.Balance, ssotNet)
			}
		}
	}

	log.Printf("Audit complete. Suspicious accounts (no SSOT, non-zero balance): %d", len(suspicious))
	for _, a := range suspicious {
		fmt.Printf("- %s %s (type=%s) balance=%.2f\n", a.Code, a.Name, a.Type, a.Balance)
	}
	return nil
}