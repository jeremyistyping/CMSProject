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
	fmt.Println("🔧 Fixing Account Hierarchy")
	fmt.Println("===========================")

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

	// Fix account hierarchy
	fmt.Println("\n1. Fixing Account Hierarchy...")
	err = fixAccountHierarchy(db, ctx)
	if err != nil {
		log.Fatal("Failed to fix account hierarchy:", err)
	}

	// Verify hierarchy
	fmt.Println("\n2. Verifying Account Hierarchy...")
	err = verifyAccountHierarchy(db, ctx)
	if err != nil {
		log.Fatal("Failed to verify account hierarchy:", err)
	}

	fmt.Println("\n✅ Account Hierarchy Fix Completed!")
}

func fixAccountHierarchy(db *gorm.DB, ctx context.Context) error {
	// Define correct parent-child relationships
	parentChildMap := map[string]string{
		"1100": "1000", // CURRENT ASSETS -> ASSETS
		"1101": "1100", // Kas -> CURRENT ASSETS
		"1102": "1100", // Bank -> CURRENT ASSETS
		"1200": "1100", // ACCOUNTS RECEIVABLE -> CURRENT ASSETS
		"1201": "1200", // Piutang Usaha -> ACCOUNTS RECEIVABLE
		"1301": "1100", // Persediaan Barang Dagangan -> CURRENT ASSETS
		"1500": "1000", // FIXED ASSETS -> ASSETS
		"1501": "1500", // Peralatan Kantor -> FIXED ASSETS
		"1502": "1500", // Kendaraan -> FIXED ASSETS
		"1503": "1500", // Bangunan -> FIXED ASSETS
		"1509": "1500", // TRUK -> FIXED ASSETS
		"2100": "2000", // CURRENT LIABILITIES -> LIABILITIES
		"2101": "2100", // Utang Usaha -> CURRENT LIABILITIES
		"1240": "1100", // PPN Masukan -> CURRENT ASSETS
		"2103": "2100", // PPN Keluaran -> CURRENT LIABILITIES
		"3101": "3000", // Modal Pemilik -> EQUITY
		"3201": "3000", // Laba Ditahan -> EQUITY
		"4101": "4000", // Pendapatan Penjualan -> REVENUE
		"4201": "4000", // Pendapatan Lain-lain -> REVENUE
		"4900": "4000", // Other Income -> REVENUE
		"5101": "5000", // Harga Pokok Penjualan -> EXPENSES
		"5201": "5000", // Beban Gaji -> EXPENSES
		"5202": "5000", // Beban Listrik -> EXPENSES
		"5203": "5000", // Beban Telepon -> EXPENSES
		"5204": "5000", // Beban Transportasi -> EXPENSES
		"5900": "5000", // General Expense -> EXPENSES
	}

	// Get account IDs
	accountMap := make(map[string]uint)
	var accounts []Account
	if err := db.Find(&accounts).Error; err != nil {
		return fmt.Errorf("failed to get accounts: %w", err)
	}

	for _, account := range accounts {
		accountMap[account.Code] = account.ID
	}

	// Update parent relationships
	for childCode, parentCode := range parentChildMap {
		if childID, childExists := accountMap[childCode]; childExists {
			if parentID, parentExists := accountMap[parentCode]; parentExists {
				if err := db.Model(&Account{}).Where("id = ?", childID).Update("parent_id", parentID).Error; err != nil {
					return fmt.Errorf("failed to update parent for %s: %w", childCode, err)
				}
				fmt.Printf("  ✅ %s -> %s (ID: %d -> %d)\n", childCode, parentCode, childID, parentID)
			} else {
				fmt.Printf("  ⚠️ Parent %s not found for child %s\n", parentCode, childCode)
			}
		} else {
			fmt.Printf("  ⚠️ Child %s not found\n", childCode)
		}
	}

	return nil
}

func verifyAccountHierarchy(db *gorm.DB, ctx context.Context) error {
	// Show current hierarchy
	var accounts []Account
	err := db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("code").
		Find(&accounts).Error
	if err != nil {
		return fmt.Errorf("failed to get accounts: %w", err)
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

	return nil
}
