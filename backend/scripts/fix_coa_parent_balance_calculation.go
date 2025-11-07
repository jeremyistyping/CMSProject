package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Account struct {
	ID       uint    `gorm:"primaryKey"`
	Code     string  `gorm:"not null;size:20;index"`
	Name     string  `gorm:"not null;size:100"`
	Type     string  `gorm:"not null;size:20"`
	ParentID *uint   `gorm:"index"`
	Level    int     `gorm:"default:1"`
	IsHeader bool    `gorm:"default:false"`
	IsActive bool    `gorm:"default:true"`
	Balance  float64 `gorm:"type:decimal(20,2);default:0"`
}

func main() {
	fmt.Println("🔧 COA Parent Balance Calculation Fix")
	fmt.Println("=====================================")

	// Database connection
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost/sistem_akuntansi?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	ctx := context.Background()

	// 1. Show current COA structure and balances
	fmt.Println("\n1. Current COA Structure:")
	showCOAStructure(db, ctx)

	// 2. Fix parent account balances based on children
	fmt.Println("\n2. Fixing Parent Account Balances...")
	err = fixParentAccountBalances(db, ctx)
	if err != nil {
		log.Fatal("Failed to fix parent balances:", err)
	}

	// 3. Show corrected COA structure
	fmt.Println("\n3. Corrected COA Structure:")
	showCOAStructure(db, ctx)

	// 4. Verify balance consistency
	fmt.Println("\n4. Verifying Balance Consistency...")
	verifyBalanceConsistency(db, ctx)

	fmt.Println("\n✅ COA Parent Balance Fix Completed!")
}

func showCOAStructure(db *gorm.DB, ctx context.Context) {
	var accounts []Account
	err := db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("code").
		Find(&accounts).Error
	if err != nil {
		log.Printf("Error fetching accounts: %v", err)
		return
	}

	// Group accounts by level for better display
	levelMap := make(map[int][]Account)
	for _, acc := range accounts {
		levelMap[acc.Level] = append(levelMap[acc.Level], acc)
	}

	// Display accounts by level
	for level := 1; level <= 4; level++ {
		if accounts, exists := levelMap[level]; exists {
			fmt.Printf("\nLevel %d:\n", level)
			for _, acc := range accounts {
				indent := ""
				for i := 0; i < level-1; i++ {
					indent += "  "
				}
				parentInfo := ""
				if acc.ParentID != nil {
					parentInfo = fmt.Sprintf(" (Parent: %d)", *acc.ParentID)
				}
				fmt.Printf("%s%s %s - %s: Rp %,.0f%s\n", 
					indent, acc.Code, acc.Name, acc.Type, acc.Balance, parentInfo)
			}
		}
	}
}

func fixParentAccountBalances(db *gorm.DB, ctx context.Context) error {
	// Get all parent accounts (accounts that have children)
	var parentAccounts []Account
	err := db.WithContext(ctx).
		Where("id IN (SELECT DISTINCT parent_id FROM accounts WHERE parent_id IS NOT NULL AND deleted_at IS NULL)").
		Find(&parentAccounts).Error
	if err != nil {
		return fmt.Errorf("failed to get parent accounts: %w", err)
	}

	fmt.Printf("Found %d parent accounts to update\n", len(parentAccounts))

	// Update each parent account balance
	for _, parent := range parentAccounts {
		// Calculate sum of children balances
		var childrenSum float64
		err = db.WithContext(ctx).
			Model(&Account{}).
			Where("parent_id = ? AND deleted_at IS NULL", parent.ID).
			Select("COALESCE(SUM(balance), 0)").
			Scan(&childrenSum).Error
		if err != nil {
			log.Printf("Error calculating children sum for parent %s (%d): %v", parent.Code, parent.ID, err)
			continue
		}

		// Update parent balance
		oldBalance := parent.Balance
		err = db.WithContext(ctx).
			Model(&Account{}).
			Where("id = ?", parent.ID).
			Update("balance", childrenSum).Error
		if err != nil {
			log.Printf("Error updating parent %s (%d): %v", parent.Code, parent.ID, err)
			continue
		}

		fmt.Printf("  ✅ %s %s: Rp %,.0f → Rp %,.0f\n", 
			parent.Code, parent.Name, oldBalance, childrenSum)
	}

	return nil
}

func verifyBalanceConsistency(db *gorm.DB, ctx context.Context) {
	// Check for inconsistent parent-child relationships
	var inconsistencies []struct {
		ParentCode    string  `json:"parent_code"`
		ParentName    string  `json:"parent_name"`
		ParentBalance float64 `json:"parent_balance"`
		ChildrenSum   float64 `json:"children_sum"`
		Difference    float64 `json:"difference"`
	}

	err := db.WithContext(ctx).Raw(`
		SELECT 
			p.code as parent_code,
			p.name as parent_name,
			p.balance as parent_balance,
			COALESCE(SUM(c.balance), 0) as children_sum,
			p.balance - COALESCE(SUM(c.balance), 0) as difference
		FROM accounts p
		LEFT JOIN accounts c ON c.parent_id = p.id AND c.deleted_at IS NULL
		WHERE p.deleted_at IS NULL 
		  AND p.id IN (SELECT DISTINCT parent_id FROM accounts WHERE parent_id IS NOT NULL AND deleted_at IS NULL)
		GROUP BY p.id, p.code, p.name, p.balance
		HAVING ABS(p.balance - COALESCE(SUM(c.balance), 0)) > 0.01
		ORDER BY p.code
	`).Scan(&inconsistencies).Error

	if err != nil {
		log.Printf("Error checking consistency: %v", err)
		return
	}

	if len(inconsistencies) == 0 {
		fmt.Println("  ✅ All parent-child balances are consistent!")
	} else {
		fmt.Printf("  ⚠️  Found %d inconsistencies:\n", len(inconsistencies))
		for _, inc := range inconsistencies {
			fmt.Printf("    %s %s: Parent=%.2f, Children=%.2f, Diff=%.2f\n", 
				inc.ParentCode, inc.ParentName, inc.ParentBalance, inc.ChildrenSum, inc.Difference)
		}
	}
}
