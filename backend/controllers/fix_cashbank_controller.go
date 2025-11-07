package controllers

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FixCashBankController struct {
	db                *gorm.DB
	cashBankService   *services.CashBankService
}

func NewFixCashBankController(db *gorm.DB, cashBankService *services.CashBankService) *FixCashBankController {
	return &FixCashBankController{
		db:              db,
		cashBankService: cashBankService,
	}
}

// FixCashBankGLLinks godoc
// @Summary Fix cash bank GL account links
// @Description Create GL accounts and link them to cash/bank accounts that are missing GL account connections
// @Tags Admin
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} models.APIResponse
// @Router /api/admin/fix-cashbank-gl-links [post]
func (c *FixCashBankController) FixCashBankGLLinks(ctx *gin.Context) {
	// Get all cash banks
	var cashBanks []models.CashBank
	if err := c.db.Preload("Account").Find(&cashBanks).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch cash banks",
			"details": err.Error(),
		})
		return
	}

	// Find accounts that need fixing
	var accountsToFix []models.CashBank
	var results []gin.H
	
	for _, cb := range cashBanks {
		// Check if account_id is missing or points to deleted/non-existent account
		if cb.AccountID == 0 || cb.Account.ID == 0 {
			accountsToFix = append(accountsToFix, cb)
		}
	}

	if len(accountsToFix) == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "All cash/bank accounts already have proper GL account links!",
			"fixed_count": 0,
			"accounts": []gin.H{},
		})
		return
	}

	// Start transaction
	tx := c.db.Begin()

	for _, cb := range accountsToFix {
		result := gin.H{
			"cash_bank_id":   cb.ID,
			"cash_bank_name": cb.Name,
			"type":           cb.Type,
			"status":         "",
			"gl_account_id":  0,
			"gl_account_code": "",
		}

		// Generate GL account code
		var code string
		if cb.Type == "CASH" {
			code = fmt.Sprintf("1100-%03d", time.Now().Unix()%1000)
		} else {
			code = fmt.Sprintf("1110-%03d", time.Now().Unix()%1000)
		}

		// Check if GL account with this name already exists
		var existingAccount models.Account
		err := tx.Where("name = ? AND type = 'ASSET' AND category = 'CURRENT_ASSET'", cb.Name).First(&existingAccount).Error
		
		var glAccountID uint
		if err == gorm.ErrRecordNotFound {
			// Create new GL account
			glAccount := models.Account{
				Code:        code,
				Name:        cb.Name,
				Type:        "ASSET",
				Category:    "CURRENT_ASSET",
				Level:       3,
				IsHeader:    false,
				IsActive:    true,
				Description: fmt.Sprintf("Auto-created GL account for %s account: %s", cb.Type, cb.Name),
			}

			// Find parent account (1100 - Current Assets)
			var parentAccount models.Account
			if err := tx.Where("code = ?", "1100").First(&parentAccount).Error; err == nil {
				glAccount.ParentID = &parentAccount.ID
			}

			if err := tx.Create(&glAccount).Error; err != nil {
				tx.Rollback()
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to create GL account",
					"details": err.Error(),
					"account": cb.Name,
				})
				return
			}
			glAccountID = glAccount.ID
			result["status"] = "created_new_gl_account"
			result["gl_account_code"] = code
		} else if err != nil {
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Error checking for existing GL account",
				"details": err.Error(),
				"account": cb.Name,
			})
			return
		} else {
			glAccountID = existingAccount.ID
			result["status"] = "using_existing_gl_account"
			result["gl_account_code"] = existingAccount.Code
		}

		// Update cash bank account to link to GL account
		if err := tx.Model(&cb).Update("account_id", glAccountID).Error; err != nil {
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to update cash bank account",
				"details": err.Error(),
				"account": cb.Name,
			})
			return
		}

		result["gl_account_id"] = glAccountID
		results = append(results, result)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to commit transaction",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":     true,
		"message":     fmt.Sprintf("Successfully fixed %d cash/bank accounts", len(accountsToFix)),
		"fixed_count": len(accountsToFix),
		"accounts":    results,
	})
}

// CheckCashBankGLLinks godoc
// @Summary Check cash bank GL account links status
// @Description Check which cash/bank accounts are missing GL account connections
// @Tags Admin
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} models.APIResponse
// @Router /api/admin/check-cashbank-gl-links [get]
func (c *FixCashBankController) CheckCashBankGLLinks(ctx *gin.Context) {
	type AccountStatus struct {
		CashBankID   uint   `json:"cash_bank_id"`
		CashBankCode string `json:"cash_bank_code"`
		CashBankName string `json:"cash_bank_name"`
		Type         string `json:"type"`
		AccountID    uint   `json:"account_id"`
		AccountCode  string `json:"account_code"`
		AccountName  string `json:"account_name"`
		Status       string `json:"status"`
	}

	var statuses []AccountStatus
	
	query := `
		SELECT 
			cb.id as cash_bank_id,
			cb.code as cash_bank_code,
			cb.name as cash_bank_name,
			cb.type,
			COALESCE(cb.account_id, 0) as account_id,
			COALESCE(a.code, '') as account_code,
			COALESCE(a.name, '') as account_name,
			CASE 
				WHEN cb.account_id IS NULL THEN 'missing_gl_account'
				WHEN a.id IS NULL THEN 'invalid_gl_account'
				ELSE 'linked_properly'
			END as status
		FROM cash_banks cb 
		LEFT JOIN accounts a ON cb.account_id = a.id AND a.deleted_at IS NULL
		WHERE cb.deleted_at IS NULL
		ORDER BY cb.id
	`

	if err := c.db.Raw(query).Scan(&statuses).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to check cash bank GL links",
			"details": err.Error(),
		})
		return
	}

	// Count statuses
	var linkedCount, missingCount, invalidCount int
	var problematicAccounts []AccountStatus

	for _, status := range statuses {
		switch status.Status {
		case "linked_properly":
			linkedCount++
		case "missing_gl_account":
			missingCount++
			problematicAccounts = append(problematicAccounts, status)
		case "invalid_gl_account":
			invalidCount++
			problematicAccounts = append(problematicAccounts, status)
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"summary": gin.H{
			"total_accounts":        len(statuses),
			"linked_properly":       linkedCount,
			"missing_gl_accounts":   missingCount,
			"invalid_gl_accounts":   invalidCount,
			"needs_fixing":          len(problematicAccounts),
		},
		"all_accounts":          statuses,
		"problematic_accounts":  problematicAccounts,
	})
}
