package main

import (
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

// COADisplayAccount represents an account with display-friendly balance
type COADisplayAccount struct {
	ID           uint    `json:"id"`
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	Category     string  `json:"category"`
	RawBalance   float64 `json:"raw_balance"`     // Balance dari database (asli)
	DisplayBalance float64 `json:"display_balance"` // Balance untuk ditampilkan (sudah dikoreksi)
	IsActive     bool    `json:"is_active"`
	ParentID     *uint   `json:"parent_id,omitempty"`
}

func main() {
	log.Printf("🔧 Creating COA Display Service with correct balance signs...")

	// Initialize database connection
	db := database.ConnectDB()

	log.Printf("\n📊 Testing COA Display Service:")
	accounts := getCOAForDisplay(db)
	
	log.Printf("\n📈 COA Accounts for Display (with corrected balances):")
	for _, acc := range accounts {
		log.Printf("   %s - %s (%s)", acc.Code, acc.Name, acc.Type)
		log.Printf("     Raw Balance: %.2f → Display Balance: %.2f", 
			acc.RawBalance, acc.DisplayBalance)
	}
}

// getCOAForDisplay returns accounts formatted for COA display with correct balance signs
func getCOAForDisplay(db *gorm.DB) []COADisplayAccount {
	var accounts []models.Account
	
	// Get active accounts with non-zero balances (what should appear in COA)
	db.Where("is_active = ? AND balance != 0", true).
		Order("code ASC").
		Find(&accounts)
		
	var displayAccounts []COADisplayAccount
	
	for _, acc := range accounts {
		displayAccount := COADisplayAccount{
			ID:         acc.ID,
			Code:       acc.Code,
			Name:       acc.Name,
			Type:       acc.Type,
			Category:   acc.Category,
			RawBalance: acc.Balance,
			DisplayBalance: getDisplayBalance(acc.Balance, acc.Type),
			IsActive:   acc.IsActive,
			ParentID:   acc.ParentID,
		}
		
		displayAccounts = append(displayAccounts, displayAccount)
	}
	
	return displayAccounts
}

// getDisplayBalance converts raw balance to display balance based on account type
func getDisplayBalance(rawBalance float64, accountType string) float64 {
	switch accountType {
	case "ASSET", "EXPENSE":
		// ASSET dan EXPENSE: tampilkan as-is
		// Positive balance = positive display
		return rawBalance
		
	case "REVENUE", "LIABILITY", "EQUITY":
		// REVENUE, LIABILITY, EQUITY: flip sign untuk display
		// Negative balance (normal for these) → positive display
		return rawBalance * -1
		
	default:
		// Default: tampilkan as-is
		return rawBalance
	}
}

// Example API endpoint implementation
func createCOAAPI() {
	log.Printf("\n💻 Example API Implementation:")
	log.Printf(`
// In your API handler (e.g., handlers/coa.go)
func GetCOAHandler(c *gin.Context) {
	db := database.ConnectDB()
	accounts := getCOAForDisplay(db)
	
	c.JSON(200, gin.H{
		"success": true,
		"data": accounts,
		"message": "COA retrieved successfully with correct balance display",
	})
}

// Frontend will receive:
{
  "success": true,
  "data": [
    {
      "code": "1101",
      "name": "Kas",
      "type": "ASSET", 
      "raw_balance": 1110000,
      "display_balance": 1110000,  // Positive (correct)
      "is_active": true
    },
    {
      "code": "4101", 
      "name": "Pendapatan Penjualan",
      "type": "REVENUE",
      "raw_balance": -1000000,
      "display_balance": 1000000,  // Positive (corrected from negative)
      "is_active": true
    },
    {
      "code": "2103",
      "name": "PPN Keluaran", 
      "type": "LIABILITY",
      "raw_balance": -110000,
      "display_balance": 110000,   // Positive (corrected from negative)
      "is_active": true
    }
  ]
}
`)
}