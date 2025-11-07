package services

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

// AccountHierarchyValidator handles validation of account hierarchies and prevents inconsistent setups
type AccountHierarchyValidator struct {
	db *gorm.DB
}

// HierarchyValidationResult represents the result of hierarchy validation
type HierarchyValidationResult struct {
	IsValid              bool                      `json:"is_valid"`
	Errors               []HierarchyError          `json:"errors"`
	Warnings             []HierarchyWarning        `json:"warnings"`
	ValidatedAccounts    int                       `json:"validated_accounts"`
	CircularReferences   []CircularReference       `json:"circular_references"`
	OrphanedAccounts     []OrphanedAccount         `json:"orphaned_accounts"`
	TypeMismatches       []TypeMismatch            `json:"type_mismatches"`
	DepthViolations      []DepthViolation          `json:"depth_violations"`
	HeaderBalanceIssues  []HeaderBalanceIssue      `json:"header_balance_issues"`
}

// HierarchyError represents a hierarchy validation error
type HierarchyError struct {
	Type        string `json:"type"`
	Message     string `json:"message"`
	AccountID   uint   `json:"account_id"`
	AccountCode string `json:"account_code"`
	AccountName string `json:"account_name"`
	Severity    string `json:"severity"`
}

// HierarchyWarning represents a hierarchy validation warning
type HierarchyWarning struct {
	Type        string `json:"type"`
	Message     string `json:"message"`
	AccountID   uint   `json:"account_id"`
	AccountCode string `json:"account_code"`
	AccountName string `json:"account_name"`
}

// CircularReference represents a circular reference in account hierarchy
type CircularReference struct {
	AccountID   uint   `json:"account_id"`
	AccountCode string `json:"account_code"`
	AccountName string `json:"account_name"`
	Path        []uint `json:"path"`
	PathCodes   []string `json:"path_codes"`
}

// OrphanedAccount represents an account with invalid parent reference
type OrphanedAccount struct {
	AccountID     uint   `json:"account_id"`
	AccountCode   string `json:"account_code"`
	AccountName   string `json:"account_name"`
	InvalidParentID uint `json:"invalid_parent_id"`
}

// TypeMismatch represents account type mismatch between parent and child
type TypeMismatch struct {
	ChildID      uint   `json:"child_id"`
	ChildCode    string `json:"child_code"`
	ChildName    string `json:"child_name"`
	ChildType    string `json:"child_type"`
	ParentID     uint   `json:"parent_id"`
	ParentCode   string `json:"parent_code"`
	ParentName   string `json:"parent_name"`
	ParentType   string `json:"parent_type"`
}

// DepthViolation represents violation of maximum hierarchy depth
type DepthViolation struct {
	AccountID     uint   `json:"account_id"`
	AccountCode   string `json:"account_code"`
	AccountName   string `json:"account_name"`
	CurrentDepth  int    `json:"current_depth"`
	MaxAllowedDepth int  `json:"max_allowed_depth"`
}

// HeaderBalanceIssue represents header account with non-zero balance
type HeaderBalanceIssue struct {
	AccountID      uint    `json:"account_id"`
	AccountCode    string  `json:"account_code"`
	AccountName    string  `json:"account_name"`
	CurrentBalance float64 `json:"current_balance"`
	ChildrenSum    float64 `json:"children_sum"`
	Difference     float64 `json:"difference"`
}

// HierarchyValidationConfig configuration for hierarchy validation
type HierarchyValidationConfig struct {
	MaxDepth                  int     `json:"max_depth"`
	AllowDifferentTypes       bool    `json:"allow_different_types"`
	AllowHeaderWithBalance    bool    `json:"allow_header_with_balance"`
	BalanceToleranceThreshold float64 `json:"balance_tolerance_threshold"`
	ValidateCircularReference bool    `json:"validate_circular_reference"`
	ValidateOrphanedAccounts  bool    `json:"validate_orphaned_accounts"`
	ValidateTypeMismatch      bool    `json:"validate_type_mismatch"`
	ValidateDepthLimit        bool    `json:"validate_depth_limit"`
	ValidateHeaderBalances    bool    `json:"validate_header_balances"`
}

// NewAccountHierarchyValidator creates a new account hierarchy validator
func NewAccountHierarchyValidator(db *gorm.DB) *AccountHierarchyValidator {
	return &AccountHierarchyValidator{
		db: db,
	}
}

// DefaultHierarchyValidationConfig returns default validation configuration
func DefaultHierarchyValidationConfig() *HierarchyValidationConfig {
	return &HierarchyValidationConfig{
		MaxDepth:                  5,    // Maximum 5 levels deep
		AllowDifferentTypes:       false, // Children must have same type as parent
		AllowHeaderWithBalance:    false, // Header accounts should not have direct balances
		BalanceToleranceThreshold: 0.01,  // 1 cent tolerance for balance differences
		ValidateCircularReference: true,
		ValidateOrphanedAccounts:  true,
		ValidateTypeMismatch:      true,
		ValidateDepthLimit:        true,
		ValidateHeaderBalances:    true,
	}
}

// ValidateAccountHierarchy performs comprehensive account hierarchy validation
func (v *AccountHierarchyValidator) ValidateAccountHierarchy(config *HierarchyValidationConfig) (*HierarchyValidationResult, error) {
	if config == nil {
		config = DefaultHierarchyValidationConfig()
	}

	log.Println("🔍 Starting comprehensive account hierarchy validation...")

	result := &HierarchyValidationResult{
		IsValid:             true,
		Errors:              []HierarchyError{},
		Warnings:            []HierarchyWarning{},
		CircularReferences:  []CircularReference{},
		OrphanedAccounts:    []OrphanedAccount{},
		TypeMismatches:      []TypeMismatch{},
		DepthViolations:     []DepthViolation{},
		HeaderBalanceIssues: []HeaderBalanceIssue{},
	}

	// Get all accounts
	var accounts []models.Account
	if err := v.db.Where("deleted_at IS NULL").Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch accounts: %w", err)
	}

	result.ValidatedAccounts = len(accounts)

	// Run validation checks
	if config.ValidateCircularReference {
		if err := v.checkCircularReferences(accounts, result); err != nil {
			return nil, fmt.Errorf("circular reference validation failed: %w", err)
		}
	}

	if config.ValidateOrphanedAccounts {
		if err := v.checkOrphanedAccounts(accounts, result); err != nil {
			return nil, fmt.Errorf("orphaned accounts validation failed: %w", err)
		}
	}

	if config.ValidateTypeMismatch {
		if err := v.checkTypeMismatches(accounts, result, config); err != nil {
			return nil, fmt.Errorf("type mismatch validation failed: %w", err)
		}
	}

	if config.ValidateDepthLimit {
		if err := v.checkDepthViolations(accounts, result, config); err != nil {
			return nil, fmt.Errorf("depth validation failed: %w", err)
		}
	}

	if config.ValidateHeaderBalances {
		if err := v.checkHeaderBalances(accounts, result, config); err != nil {
			return nil, fmt.Errorf("header balance validation failed: %w", err)
		}
	}

	// Determine overall validity
	result.IsValid = len(result.Errors) == 0

	// Log results
	v.logValidationResults(result)

	return result, nil
}

// checkCircularReferences checks for circular references in the hierarchy
func (v *AccountHierarchyValidator) checkCircularReferences(accounts []models.Account, result *HierarchyValidationResult) error {
	log.Println("  🔄 Checking for circular references...")

	accountMap := make(map[uint]*models.Account)
	for i := range accounts {
		accountMap[accounts[i].ID] = &accounts[i]
	}

	visited := make(map[uint]bool)
	recursionStack := make(map[uint]bool)

	var dfsCheck func(uint, []uint, []string) bool
	dfsCheck = func(accountID uint, path []uint, pathCodes []string) bool {
		if recursionStack[accountID] {
			// Found circular reference
			circularRef := CircularReference{
				AccountID:   accountID,
				AccountCode: accountMap[accountID].Code,
				AccountName: accountMap[accountID].Name,
				Path:        append(path, accountID),
				PathCodes:   append(pathCodes, accountMap[accountID].Code),
			}
			result.CircularReferences = append(result.CircularReferences, circularRef)

			error := HierarchyError{
				Type:        "circular_reference",
				Message:     fmt.Sprintf("Circular reference detected: %s", strings.Join(circularRef.PathCodes, " -> ")),
				AccountID:   accountID,
				AccountCode: accountMap[accountID].Code,
				AccountName: accountMap[accountID].Name,
				Severity:    "critical",
			}
			result.Errors = append(result.Errors, error)
			return true
		}

		if visited[accountID] {
			return false
		}

		visited[accountID] = true
		recursionStack[accountID] = true

		account := accountMap[accountID]
		if account != nil && account.ParentID != nil {
			if dfsCheck(*account.ParentID, append(path, accountID), append(pathCodes, account.Code)) {
				return true
			}
		}

		recursionStack[accountID] = false
		return false
	}

	for _, account := range accounts {
		if !visited[account.ID] {
			dfsCheck(account.ID, []uint{}, []string{})
		}
	}

	log.Printf("    Found %d circular references", len(result.CircularReferences))
	return nil
}

// checkOrphanedAccounts checks for accounts with invalid parent references
func (v *AccountHierarchyValidator) checkOrphanedAccounts(accounts []models.Account, result *HierarchyValidationResult) error {
	log.Println("  👻 Checking for orphaned accounts...")

	validAccountIDs := make(map[uint]bool)
	for _, account := range accounts {
		validAccountIDs[account.ID] = true
	}

	for _, account := range accounts {
		if account.ParentID != nil && !validAccountIDs[*account.ParentID] {
			orphaned := OrphanedAccount{
				AccountID:       account.ID,
				AccountCode:     account.Code,
				AccountName:     account.Name,
				InvalidParentID: *account.ParentID,
			}
			result.OrphanedAccounts = append(result.OrphanedAccounts, orphaned)

			error := HierarchyError{
				Type:        "orphaned_account",
				Message:     fmt.Sprintf("Account has invalid parent ID %d", *account.ParentID),
				AccountID:   account.ID,
				AccountCode: account.Code,
				AccountName: account.Name,
				Severity:    "high",
			}
			result.Errors = append(result.Errors, error)
		}
	}

	log.Printf("    Found %d orphaned accounts", len(result.OrphanedAccounts))
	return nil
}

// checkTypeMismatches checks for account type mismatches between parent and child
func (v *AccountHierarchyValidator) checkTypeMismatches(accounts []models.Account, result *HierarchyValidationResult, config *HierarchyValidationConfig) error {
	if config.AllowDifferentTypes {
		return nil
	}

	log.Println("  🏷️ Checking for type mismatches...")

	accountMap := make(map[uint]*models.Account)
	for i := range accounts {
		accountMap[accounts[i].ID] = &accounts[i]
	}

	for _, account := range accounts {
		if account.ParentID != nil {
			parent := accountMap[*account.ParentID]
			if parent != nil && parent.Type != account.Type {
				mismatch := TypeMismatch{
					ChildID:    account.ID,
					ChildCode:  account.Code,
					ChildName:  account.Name,
					ChildType:  account.Type,
					ParentID:   parent.ID,
					ParentCode: parent.Code,
					ParentName: parent.Name,
					ParentType: parent.Type,
				}
				result.TypeMismatches = append(result.TypeMismatches, mismatch)

				error := HierarchyError{
					Type:        "type_mismatch",
					Message:     fmt.Sprintf("Account type %s doesn't match parent type %s", account.Type, parent.Type),
					AccountID:   account.ID,
					AccountCode: account.Code,
					AccountName: account.Name,
					Severity:    "medium",
				}
				result.Errors = append(result.Errors, error)
			}
		}
	}

	log.Printf("    Found %d type mismatches", len(result.TypeMismatches))
	return nil
}

// checkDepthViolations checks for accounts that exceed maximum depth
func (v *AccountHierarchyValidator) checkDepthViolations(accounts []models.Account, result *HierarchyValidationResult, config *HierarchyValidationConfig) error {
	log.Printf("  📏 Checking for depth violations (max depth: %d)...", config.MaxDepth)

	accountMap := make(map[uint]*models.Account)
	for i := range accounts {
		accountMap[accounts[i].ID] = &accounts[i]
	}

	var calculateDepth func(uint, map[uint]int) int
	calculateDepth = func(accountID uint, depthCache map[uint]int) int {
		if depth, exists := depthCache[accountID]; exists {
			return depth
		}

		account := accountMap[accountID]
		if account == nil || account.ParentID == nil {
			depthCache[accountID] = 0
			return 0
		}

		parentDepth := calculateDepth(*account.ParentID, depthCache)
		depth := parentDepth + 1
		depthCache[accountID] = depth
		return depth
	}

	depthCache := make(map[uint]int)
	for _, account := range accounts {
		depth := calculateDepth(account.ID, depthCache)
		if depth > config.MaxDepth {
			violation := DepthViolation{
				AccountID:       account.ID,
				AccountCode:     account.Code,
				AccountName:     account.Name,
				CurrentDepth:    depth,
				MaxAllowedDepth: config.MaxDepth,
			}
			result.DepthViolations = append(result.DepthViolations, violation)

			error := HierarchyError{
				Type:        "depth_violation",
				Message:     fmt.Sprintf("Account depth %d exceeds maximum allowed depth %d", depth, config.MaxDepth),
				AccountID:   account.ID,
				AccountCode: account.Code,
				AccountName: account.Name,
				Severity:    "medium",
			}
			result.Errors = append(result.Errors, error)
		}
	}

	log.Printf("    Found %d depth violations", len(result.DepthViolations))
	return nil
}

// checkHeaderBalances checks header accounts for balance consistency
func (v *AccountHierarchyValidator) checkHeaderBalances(accounts []models.Account, result *HierarchyValidationResult, config *HierarchyValidationConfig) error {
	log.Println("  ⚖️ Checking header account balances...")

	headerAccounts := make([]models.Account, 0)
	for _, account := range accounts {
		if account.IsHeader {
			headerAccounts = append(headerAccounts, account)
		}
	}

	for _, header := range headerAccounts {
		var childrenSum float64
		if err := v.db.Model(&models.Account{}).
			Where("parent_id = ? AND deleted_at IS NULL", header.ID).
			Select("COALESCE(SUM(balance), 0)").
			Scan(&childrenSum).Error; err != nil {
			return fmt.Errorf("failed to calculate children sum for account %s: %w", header.Code, err)
		}

		difference := header.Balance - childrenSum

		// Check if header has balance when it shouldn't
		if !config.AllowHeaderWithBalance && header.Balance != 0 {
			warning := HierarchyWarning{
				Type:        "header_with_balance",
				Message:     fmt.Sprintf("Header account has non-zero balance: %.2f", header.Balance),
				AccountID:   header.ID,
				AccountCode: header.Code,
				AccountName: header.Name,
			}
			result.Warnings = append(result.Warnings, warning)
		}

		// Check if header balance doesn't match children sum
		if difference > config.BalanceToleranceThreshold || difference < -config.BalanceToleranceThreshold {
			issue := HeaderBalanceIssue{
				AccountID:      header.ID,
				AccountCode:    header.Code,
				AccountName:    header.Name,
				CurrentBalance: header.Balance,
				ChildrenSum:    childrenSum,
				Difference:     difference,
			}
			result.HeaderBalanceIssues = append(result.HeaderBalanceIssues, issue)

			error := HierarchyError{
				Type:        "header_balance_mismatch",
				Message:     fmt.Sprintf("Header balance %.2f doesn't match children sum %.2f (diff: %.2f)", header.Balance, childrenSum, difference),
				AccountID:   header.ID,
				AccountCode: header.Code,
				AccountName: header.Name,
				Severity:    "medium",
			}
			result.Errors = append(result.Errors, error)
		}
	}

	log.Printf("    Found %d header balance issues", len(result.HeaderBalanceIssues))
	return nil
}

// logValidationResults logs the validation results
func (v *AccountHierarchyValidator) logValidationResults(result *HierarchyValidationResult) {
	log.Printf("🏁 Account hierarchy validation completed:")
	log.Printf("   📊 Accounts validated: %d", result.ValidatedAccounts)
	log.Printf("   ✅ Valid: %t", result.IsValid)
	log.Printf("   ❌ Errors: %d", len(result.Errors))
	log.Printf("   ⚠️ Warnings: %d", len(result.Warnings))

	if len(result.CircularReferences) > 0 {
		log.Printf("   🔄 Circular references: %d", len(result.CircularReferences))
	}
	if len(result.OrphanedAccounts) > 0 {
		log.Printf("   👻 Orphaned accounts: %d", len(result.OrphanedAccounts))
	}
	if len(result.TypeMismatches) > 0 {
		log.Printf("   🏷️ Type mismatches: %d", len(result.TypeMismatches))
	}
	if len(result.DepthViolations) > 0 {
		log.Printf("   📏 Depth violations: %d", len(result.DepthViolations))
	}
	if len(result.HeaderBalanceIssues) > 0 {
		log.Printf("   ⚖️ Header balance issues: %d", len(result.HeaderBalanceIssues))
	}
}

// ValidateAccountUpdate validates an account update for hierarchy consistency
func (v *AccountHierarchyValidator) ValidateAccountUpdate(accountID uint, newParentID *uint, newType string) error {
	// Get the account
	var account models.Account
	if err := v.db.Where("id = ? AND deleted_at IS NULL", accountID).First(&account).Error; err != nil {
		return fmt.Errorf("account not found: %w", err)
	}

	// Check for circular reference if parent is changing
	if newParentID != nil && (account.ParentID == nil || *account.ParentID != *newParentID) {
		if err := v.checkWouldCreateCircularReference(accountID, *newParentID); err != nil {
			return err
		}
	}

	// Check type consistency if parent exists
	if newParentID != nil {
		var parent models.Account
		if err := v.db.Where("id = ? AND deleted_at IS NULL", *newParentID).First(&parent).Error; err != nil {
			return fmt.Errorf("parent account not found: %w", err)
		}

		if parent.Type != newType {
			return fmt.Errorf("account type %s doesn't match parent type %s", newType, parent.Type)
		}
	}

	return nil
}

// checkWouldCreateCircularReference checks if setting a parent would create a circular reference
func (v *AccountHierarchyValidator) checkWouldCreateCircularReference(accountID uint, parentID uint) error {
	if accountID == parentID {
		return errors.New("account cannot be its own parent")
	}

	// Check if the proposed parent is actually a descendant of this account
	var checkAncestry func(uint) bool
	checkAncestry = func(currentID uint) bool {
		if currentID == accountID {
			return true
		}

		var account models.Account
		if err := v.db.Where("id = ? AND deleted_at IS NULL", currentID).First(&account).Error; err != nil {
			return false
		}

		if account.ParentID != nil {
			return checkAncestry(*account.ParentID)
		}
		return false
	}

	if checkAncestry(parentID) {
		return fmt.Errorf("setting parent would create circular reference")
	}

	return nil
}

// FixHierarchyIssues attempts to fix detected hierarchy issues
func (v *AccountHierarchyValidator) FixHierarchyIssues(result *HierarchyValidationResult, autoFix bool) error {
	if result.IsValid {
		log.Println("✅ No hierarchy issues to fix")
		return nil
	}

	log.Printf("🔧 Attempting to fix %d hierarchy issues...", len(result.Errors))
	fixedCount := 0

	// Fix orphaned accounts by removing invalid parent references
	for _, orphaned := range result.OrphanedAccounts {
		if autoFix {
			if err := v.db.Model(&models.Account{}).
				Where("id = ?", orphaned.AccountID).
				Update("parent_id", nil).Error; err != nil {
				log.Printf("❌ Failed to fix orphaned account %s: %v", orphaned.AccountCode, err)
			} else {
				log.Printf("✅ Fixed orphaned account %s by removing invalid parent reference", orphaned.AccountCode)
				fixedCount++
			}
		} else {
			log.Printf("📝 Would fix orphaned account %s by removing invalid parent reference", orphaned.AccountCode)
		}
	}

	// Fix header balance issues
	for _, issue := range result.HeaderBalanceIssues {
		if autoFix {
			if err := v.db.Model(&models.Account{}).
				Where("id = ?", issue.AccountID).
				Update("balance", issue.ChildrenSum).Error; err != nil {
				log.Printf("❌ Failed to fix header balance for %s: %v", issue.AccountCode, err)
			} else {
				log.Printf("✅ Fixed header balance for %s: %.2f -> %.2f", issue.AccountCode, issue.CurrentBalance, issue.ChildrenSum)
				fixedCount++
			}
		} else {
			log.Printf("📝 Would fix header balance for %s: %.2f -> %.2f", issue.AccountCode, issue.CurrentBalance, issue.ChildrenSum)
		}
	}

	if autoFix {
		log.Printf("🎉 Fixed %d hierarchy issues", fixedCount)
	} else {
		log.Printf("📋 Would fix %d hierarchy issues (dry run)", fixedCount)
	}

	return nil
}