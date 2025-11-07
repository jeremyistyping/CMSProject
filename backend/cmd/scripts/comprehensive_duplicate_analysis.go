package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println("COMPREHENSIVE DUPLICATE ACCOUNTS ANALYSIS")
	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println()

	// Step 1: Find ALL duplicate codes (including soft deleted)
	fmt.Println("📊 STEP 1: All Duplicate Codes (Including Soft Deleted)")
	fmt.Println(string(make([]byte, 80)))

	type DuplicateCode struct {
		Code        string
		TotalCount  int
		ActiveCount int
		DeletedCount int
	}

	var allDuplicates []DuplicateCode
	db.Raw(`
		SELECT 
			code,
			COUNT(*) as total_count,
			COUNT(CASE WHEN deleted_at IS NULL THEN 1 END) as active_count,
			COUNT(CASE WHEN deleted_at IS NOT NULL THEN 1 END) as deleted_count
		FROM accounts
		GROUP BY code
		HAVING COUNT(*) > 1
		ORDER BY total_count DESC, code
	`).Scan(&allDuplicates)

	if len(allDuplicates) > 0 {
		fmt.Printf("Found %d codes with duplicates:\n\n", len(allDuplicates))
		for _, dup := range allDuplicates {
			fmt.Printf("  Code: %-10s → Total: %d | Active: %d | Deleted: %d\n", 
				dup.Code, dup.TotalCount, dup.ActiveCount, dup.DeletedCount)
		}
	} else {
		fmt.Println("✅ NO DUPLICATES FOUND (including soft deleted)")
	}
	fmt.Println()

	// Step 2: Check the specific accounts mentioned
	fmt.Println("📊 STEP 2: Detailed Check for Specific Accounts (5101, 4101, 1301)")
	fmt.Println(string(make([]byte, 80)))

	targetCodes := []string{"5101", "4101", "1301"}

	for _, code := range targetCodes {
		type AccountDetail struct {
			ID          uint
			Code        string
			Name        string
			Type        string
			Balance     float64
			IsActive    bool
			IsHeader    bool
			Level       int
			ParentID    *uint
			CreatedAt   string
			UpdatedAt   string
			DeletedAt   *string
		}

		var accounts []AccountDetail
		db.Raw(`
			SELECT 
				id, code, name, type, balance, is_active, is_header, level, parent_id,
				created_at::text, updated_at::text, deleted_at::text
			FROM accounts
			WHERE code = ?
			ORDER BY id
		`, code).Scan(&accounts)

		fmt.Printf("\n🔍 Code: %s → Found %d records\n", code, len(accounts))
		fmt.Println(string(make([]byte, 80)))

		for i, acc := range accounts {
			deletedStatus := "✅ ACTIVE"
			if acc.DeletedAt != nil {
				deletedStatus = "❌ SOFT DELETED"
			}

			fmt.Printf("\n%d. Account ID: %d %s\n", i+1, acc.ID, deletedStatus)
			fmt.Printf("   Code: %s | Name: %s\n", acc.Code, acc.Name)
			fmt.Printf("   Type: %s | Balance: Rp %.2f\n", acc.Type, acc.Balance)
			fmt.Printf("   Header: %v (Level %d) | Active: %v\n", acc.IsHeader, acc.Level, acc.IsActive)
			
			if acc.ParentID != nil {
				var parentCode string
				db.Raw("SELECT code FROM accounts WHERE id = ?", *acc.ParentID).Scan(&parentCode)
				fmt.Printf("   Parent ID: %d (Code: %s)\n", *acc.ParentID, parentCode)
			} else {
				fmt.Printf("   Parent ID: NULL\n")
			}
			
			fmt.Printf("   Created: %s\n", acc.CreatedAt)
			fmt.Printf("   Updated: %s\n", acc.UpdatedAt)
			if acc.DeletedAt != nil {
				fmt.Printf("   Deleted: %s\n", *acc.DeletedAt)
			}
		}
	}
	fmt.Println()

	// Step 3: Check for similar names
	fmt.Println("\n📊 STEP 3: Check for Accounts with Similar/Same Names")
	fmt.Println(string(make([]byte, 80)))

	type SimilarName struct {
		Name  string
		Count int
		Codes string
	}

	var similarNames []SimilarName
	db.Raw(`
		SELECT 
			UPPER(TRIM(name)) as name,
			COUNT(*) as count,
			STRING_AGG(DISTINCT code, ', ' ORDER BY code) as codes
		FROM accounts
		WHERE deleted_at IS NULL
		GROUP BY UPPER(TRIM(name))
		HAVING COUNT(*) > 1
		ORDER BY count DESC
	`).Scan(&similarNames)

	if len(similarNames) > 0 {
		fmt.Printf("Found %d account names used multiple times:\n\n", len(similarNames))
		for _, sim := range similarNames {
			fmt.Printf("  Name: %-40s → %d accounts | Codes: %s\n", sim.Name, sim.Count, sim.Codes)
		}
		fmt.Println()
		fmt.Println("⚠️  These might be intentional (different codes) or unintentional duplicates")
	} else {
		fmt.Println("✅ No accounts with duplicate names")
	}
	fmt.Println()

	// Step 4: Check COA data shown earlier
	fmt.Println("\n📊 STEP 4: Replicate Original Query from COA Display")
	fmt.Println(string(make([]byte, 80)))
	fmt.Println("This query simulates what was shown in the COA tree view:")
	fmt.Println()

	type COADisplay struct {
		ID      uint
		Code    string
		Name    string
		Type    string
		Balance float64
	}

	// Query similar to what frontend might use
	var coaRecords []COADisplay
	db.Raw(`
		SELECT id, code, name, type, balance
		FROM accounts
		WHERE deleted_at IS NULL
		  AND code IN ('5101', '4101', '1301')
		ORDER BY code, id
	`).Scan(&coaRecords)

	fmt.Printf("Records found: %d\n\n", len(coaRecords))
	
	groupedByCode := make(map[string][]COADisplay)
	for _, rec := range coaRecords {
		groupedByCode[rec.Code] = append(groupedByCode[rec.Code], rec)
	}

	for code, records := range groupedByCode {
		fmt.Printf("Code: %s (%d records)\n", code, len(records))
		for _, rec := range records {
			fmt.Printf("  ID: %d | Name: %s | Balance: Rp %.2f\n", rec.ID, rec.Name, rec.Balance)
		}
		fmt.Println()
	}

	// Step 5: Root cause hypothesis
	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println("📋 ROOT CAUSE ANALYSIS & HYPOTHESIS")
	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println()

	if len(allDuplicates) == 0 && len(coaRecords) == 3 {
		fmt.Println("✅ FINDING: No actual duplicates found in database!")
		fmt.Println()
		fmt.Println("🔍 POSSIBLE EXPLANATIONS for earlier '3x duplicate' report:")
		fmt.Println()
		fmt.Println("1. ❌ FALSE POSITIVE from Query:")
		fmt.Println("   - Earlier query might have included soft-deleted records")
		fmt.Println("   - Or query had WHERE code IN ('5101') which returned records from")
		fmt.Println("     different queries concatenated together")
		fmt.Println()
		fmt.Println("2. 🔄 TEMPORARY STATE:")
		fmt.Println("   - Duplicates existed briefly during migration/seed")
		fmt.Println("   - FirstOrCreate logic cleaned them up automatically")
		fmt.Println("   - When you saw '3x', it was during a transaction rollback")
		fmt.Println()
		fmt.Println("3. 📺 FRONTEND RENDERING ISSUE:")
		fmt.Println("   - Same account rendered multiple times in UI")
		fmt.Println("   - JavaScript grouping/filtering bug")
		fmt.Println("   - React re-render caused duplicate display")
		fmt.Println()
		fmt.Println("4. 🔍 SEED.GO BEHAVIOR:")
		fmt.Println("   - SeedAccounts uses FirstOrCreate (line 67)")
		fmt.Println("   - This PREVENTS duplicates by design")
		fmt.Println("   - Even if called multiple times, no duplicates should occur")
		fmt.Println()
	} else if len(allDuplicates) > 0 {
		fmt.Println("⚠️  CONFIRMED: Actual duplicates exist in database")
		fmt.Println()
		fmt.Println("🔍 LIKELY CAUSES:")
		fmt.Println()
		fmt.Println("1. 🔄 Multiple Concurrent Seed Runs:")
		fmt.Println("   - Two processes called SeedAccounts simultaneously")
		fmt.Println("   - FirstOrCreate race condition occurred")
		fmt.Println("   - Both created the account before checking")
		fmt.Println()
		fmt.Println("2. ❌ Missing UNIQUE Constraint:")
		fmt.Println("   - Database allows duplicate codes")
		fmt.Println("   - No protection at database level")
		fmt.Println("   - Race conditions can bypass application logic")
		fmt.Println()
		fmt.Println("3. 📦 Data Import/Restore:")
		fmt.Println("   - Backup restored on top of existing data")
		fmt.Println("   - Manual SQL INSERT without duplicate check")
		fmt.Println()
		fmt.Println("4. 🐛 Bug in Past Version:")
		fmt.Println("   - Old version didn't use FirstOrCreate")
		fmt.Println("   - Created duplicates before code was fixed")
		fmt.Println()
	}

	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println("✅ ANALYSIS COMPLETE")
	fmt.Println("=" + string(make([]byte, 80)))
}

