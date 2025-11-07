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
	fmt.Println("INVESTIGATE DUPLICATE ACCOUNTS ROOT CAUSE")
	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println()

	// Step 1: Find all duplicate account codes
	fmt.Println("📊 STEP 1: Find All Duplicate Account Codes")
	fmt.Println(string(make([]byte, 80)))

	type DuplicateCode struct {
		Code  string
		Count int
	}

	var duplicates []DuplicateCode
	db.Raw(`
		SELECT code, COUNT(*) as count
		FROM accounts
		WHERE deleted_at IS NULL
		GROUP BY code
		HAVING COUNT(*) > 1
		ORDER BY count DESC, code
	`).Scan(&duplicates)

	fmt.Printf("Found %d duplicate account codes:\n\n", len(duplicates))

	for _, dup := range duplicates {
		fmt.Printf("  Code: %-10s → %d accounts\n", dup.Code, dup.Count)
	}
	fmt.Println()

	// Step 2: Detailed analysis of duplicate accounts
	fmt.Println("📊 STEP 2: Detailed Analysis of Duplicate Accounts")
	fmt.Println(string(make([]byte, 80)))

	type AccountDetail struct {
		ID          uint
		Code        string
		Name        string
		Type        string
		Balance     float64
		IsActive    bool
		CreatedAt   string
		UpdatedAt   string
		ParentID    *uint
		IsHeader    bool
		Level       int
		Description string
	}

	// Focus on the specific duplicates mentioned
	targetCodes := []string{"5101", "4101", "1301"}

	for _, code := range targetCodes {
		var accounts []AccountDetail
		db.Raw(`
			SELECT 
				id, code, name, type, balance, is_active, 
				created_at::text, updated_at::text,
				parent_id, is_header, level, description
			FROM accounts
			WHERE code = ?
			  AND deleted_at IS NULL
			ORDER BY id
		`, code).Scan(&accounts)

		fmt.Printf("\n🔍 Code: %s (%d duplicates)\n", code, len(accounts))
		fmt.Println(string(make([]byte, 80)))

		for i, acc := range accounts {
			fmt.Printf("\n%d. Account ID: %d\n", i+1, acc.ID)
			fmt.Printf("   Name:        %s\n", acc.Name)
			fmt.Printf("   Type:        %s\n", acc.Type)
			fmt.Printf("   Balance:     Rp %.2f\n", acc.Balance)
			fmt.Printf("   Active:      %v\n", acc.IsActive)
			fmt.Printf("   Is Header:   %v (Level: %d)\n", acc.IsHeader, acc.Level)
			if acc.ParentID != nil {
				fmt.Printf("   Parent ID:   %d\n", *acc.ParentID)
			} else {
				fmt.Printf("   Parent ID:   NULL (no parent)\n")
			}
			fmt.Printf("   Created At:  %s\n", acc.CreatedAt)
			fmt.Printf("   Updated At:  %s\n", acc.UpdatedAt)
			if acc.Description != "" {
				fmt.Printf("   Description: %s\n", acc.Description)
			}
		}
		fmt.Println()
	}

	// Step 3: Check which account is being used in transactions
	fmt.Println("\n📊 STEP 3: Check Which Accounts Are Used in Transactions")
	fmt.Println(string(make([]byte, 80)))

	for _, code := range targetCodes {
		type AccountUsage struct {
			AccountID    uint
			Name         string
			JournalCount int
			TotalDebit   float64
			TotalCredit  float64
		}

		var usage []AccountUsage
		db.Raw(`
			SELECT 
				a.id as account_id,
				a.name,
				COUNT(DISTINCT jl.id) as journal_count,
				COALESCE(SUM(jl.debit_amount), 0) as total_debit,
				COALESCE(SUM(jl.credit_amount), 0) as total_credit
			FROM accounts a
			LEFT JOIN unified_journal_lines jl ON jl.account_id = a.id
			WHERE a.code = ?
			  AND a.deleted_at IS NULL
			GROUP BY a.id, a.name
			ORDER BY a.id
		`, code).Scan(&usage)

		fmt.Printf("\n🔍 Code: %s - Transaction Usage\n", code)
		fmt.Println(string(make([]byte, 80)))

		for _, u := range usage {
			status := "❌ NOT USED"
			if u.JournalCount > 0 {
				status = "✅ ACTIVE (In Use)"
			}
			fmt.Printf("  ID: %-5d | %-40s | Entries: %4d | %s\n", 
				u.AccountID, u.Name, u.JournalCount, status)
			if u.JournalCount > 0 {
				fmt.Printf("           Dr: Rp %15.2f | Cr: Rp %15.2f\n", u.TotalDebit, u.TotalCredit)
			}
		}
	}
	fmt.Println()

	// Step 4: Check database constraints
	fmt.Println("\n📊 STEP 4: Check Database Constraints on 'code' Column")
	fmt.Println(string(make([]byte, 80)))

	type ConstraintInfo struct {
		ConstraintName string
		ConstraintType string
	}

	var constraints []ConstraintInfo
	db.Raw(`
		SELECT 
			tc.constraint_name,
			tc.constraint_type
		FROM information_schema.table_constraints tc
		JOIN information_schema.constraint_column_usage ccu 
			ON tc.constraint_name = ccu.constraint_name
		WHERE tc.table_name = 'accounts'
		  AND ccu.column_name = 'code'
		  AND tc.constraint_type IN ('UNIQUE', 'PRIMARY KEY')
	`).Scan(&constraints)

	if len(constraints) == 0 {
		fmt.Println("❌ NO UNIQUE CONSTRAINT on 'code' column!")
		fmt.Println("   This allows duplicate account codes to be created.")
		fmt.Println()
		fmt.Println("   💡 RECOMMENDED FIX:")
		fmt.Println("      1. Clean up duplicate accounts first")
		fmt.Println("      2. Add unique constraint: ALTER TABLE accounts ADD CONSTRAINT accounts_code_unique UNIQUE (code);")
	} else {
		fmt.Println("✅ Found constraints:")
		for _, c := range constraints {
			fmt.Printf("   - %s (%s)\n", c.ConstraintName, c.ConstraintType)
		}
	}
	fmt.Println()

	// Step 5: Check for backup tables
	fmt.Println("\n📊 STEP 5: Check Account Backup Tables")
	fmt.Println(string(make([]byte, 80)))

	type BackupTable struct {
		TableName string
		RowCount  int
	}

	var backups []BackupTable
	db.Raw(`
		SELECT 
			t.table_name,
			(SELECT COUNT(*) FROM ` + "`\" || t.table_name || \"`" + `) as row_count
		FROM information_schema.tables t
		WHERE t.table_schema = 'public'
		  AND t.table_name LIKE 'accounts%backup%'
		ORDER BY t.table_name
	`).Scan(&backups)

	if len(backups) > 0 {
		fmt.Printf("Found %d backup tables:\n\n", len(backups))
		for _, b := range backups {
			fmt.Printf("  📦 %-40s → %d rows\n", b.TableName, b.RowCount)
		}
		fmt.Println()
		fmt.Println("⚠️  Possible cause: Data might have been restored from backup")
		fmt.Println("   or migration script ran multiple times")
	} else {
		fmt.Println("No backup tables found")
	}
	fmt.Println()

	// Step 6: Timeline analysis
	fmt.Println("\n📊 STEP 6: Creation Timeline (When Were Duplicates Created?)")
	fmt.Println(string(make([]byte, 80)))

	for _, code := range targetCodes {
		type Timeline struct {
			ID        uint
			Name      string
			CreatedAt string
			TimeDiff  string
		}

		var timeline []Timeline
		db.Raw(`
			SELECT 
				id,
				name,
				created_at::text,
				(created_at - LAG(created_at) OVER (ORDER BY created_at))::text as time_diff
			FROM accounts
			WHERE code = ?
			  AND deleted_at IS NULL
			ORDER BY created_at
		`, code).Scan(&timeline)

		fmt.Printf("\n🔍 Code: %s - Creation Timeline\n", code)
		for i, t := range timeline {
			timeDiffStr := "FIRST"
			if t.TimeDiff != "" {
				timeDiffStr = t.TimeDiff + " after previous"
			}
			fmt.Printf("  %d. ID: %-5d | Created: %s | %s\n", 
				i+1, t.ID, t.CreatedAt, timeDiffStr)
			fmt.Printf("     Name: %s\n", t.Name)
		}
	}
	fmt.Println()

	// Summary
	fmt.Println("\n" + string(make([]byte, 80)))
	fmt.Println("📋 SUMMARY & ROOT CAUSE ANALYSIS")
	fmt.Println(string(make([]byte, 80)))
	fmt.Println()
	fmt.Println("🔍 Possible Causes:")
	fmt.Println("   1. ❌ No UNIQUE constraint on 'code' column")
	fmt.Println("   2. 🔄 Migration/seed scripts ran multiple times")
	fmt.Println("   3. 📦 Data restored from backup with existing data")
	fmt.Println("   4. 🔧 Manual SQL insertions without checking for duplicates")
	fmt.Println("   5. 🐛 Bug in account creation logic (no duplicate check)")
	fmt.Println()
	fmt.Println("✅ Solution Steps:")
	fmt.Println("   1. Identify which account is actively used (has transactions)")
	fmt.Println("   2. Soft delete or merge unused duplicate accounts")
	fmt.Println("   3. Add UNIQUE constraint to prevent future duplicates")
	fmt.Println("   4. Update application code to check for duplicates before insert")
	fmt.Println()

	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println("✅ INVESTIGATION COMPLETE")
	fmt.Println("=" + string(make([]byte, 80)))
}

