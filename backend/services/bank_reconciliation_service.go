package services

import (
	"app-sistem-akuntansi/models"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type BankReconciliationService struct {
	db *gorm.DB
}

func NewBankReconciliationService(db *gorm.DB) *BankReconciliationService {
	return &BankReconciliationService{db: db}
}

// ========== SNAPSHOT OPERATIONS ==========

// CreateSnapshotRequest - Request untuk membuat snapshot
type CreateSnapshotRequest struct {
	CashBankID uint   `json:"cash_bank_id" binding:"required"`
	Period     string `json:"period" binding:"required"` // Format: YYYY-MM
	Notes      string `json:"notes"`
}

// GenerateSnapshot - Membuat snapshot transaksi untuk periode tertentu
func (s *BankReconciliationService) GenerateSnapshot(req CreateSnapshotRequest, userID uint) (*models.BankReconciliationSnapshot, error) {
	// Validate cash bank exists
	var cashBank models.CashBank
	if err := s.db.First(&cashBank, req.CashBankID).Error; err != nil {
		return nil, errors.New("cash bank account not found")
	}

	// Parse period
	periodTime, err := time.Parse("2006-01", req.Period)
	if err != nil {
		return nil, errors.New("invalid period format, use YYYY-MM")
	}

	// Get transactions for the period
	startDate := periodTime
	endDate := periodTime.AddDate(0, 1, 0).Add(-time.Second)

	var transactions []models.CashBankTransaction
	if err := s.db.Where("cash_bank_id = ? AND transaction_date >= ? AND transaction_date <= ?",
		req.CashBankID, startDate, endDate).
		Order("transaction_date ASC, id ASC").
		Find(&transactions).Error; err != nil {
		return nil, err
	}

	// Calculate balances
	var openingBalance, closingBalance, totalDebit, totalCredit float64
	
	// Get opening balance (balance before period start)
	if err := s.db.Raw(`
		SELECT COALESCE(balance_after, 0) as balance
		FROM cash_bank_transactions
		WHERE cash_bank_id = ? AND transaction_date < ?
		ORDER BY transaction_date DESC, id DESC
		LIMIT 1
	`, req.CashBankID, startDate).Scan(&openingBalance).Error; err != nil {
		// If no previous transactions, opening balance is 0
		openingBalance = 0
	}

	// Calculate totals
	for _, tx := range transactions {
		if tx.Amount > 0 {
			totalDebit += tx.Amount
		} else {
			totalCredit += -tx.Amount
		}
	}

	// Closing balance is opening + debit - credit
	closingBalance = openingBalance + totalDebit - totalCredit

	// Create snapshot
	snapshot := models.BankReconciliationSnapshot{
		CashBankID:       req.CashBankID,
		Period:           req.Period,
		SnapshotDate:     time.Now(),
		GeneratedBy:      userID,
		OpeningBalance:   openingBalance,
		ClosingBalance:   closingBalance,
		TotalDebit:       totalDebit,
		TotalCredit:      totalCredit,
		TransactionCount: len(transactions),
		Notes:            req.Notes,
		Status:           "ACTIVE",
	}

	// Begin transaction
	tx := s.db.Begin()

	// Save snapshot
	if err := tx.Create(&snapshot).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Save transaction snapshots
	for _, transaction := range transactions {
		snapshotTx := models.ReconciliationTransactionSnapshot{
			SnapshotID:      snapshot.ID,
			TransactionID:   transaction.ID,
			TransactionDate: transaction.TransactionDate,
			ReferenceType:   transaction.ReferenceType,
			ReferenceID:     transaction.ReferenceID,
			Amount:          transaction.Amount,
			BalanceAfter:    transaction.BalanceAfter,
			Notes:           transaction.Notes,
			CreatedBy:       userID,
		}

		// Determine debit/credit
		if transaction.Amount > 0 {
			snapshotTx.DebitAmount = transaction.Amount
			snapshotTx.CreditAmount = 0
		} else {
			snapshotTx.DebitAmount = 0
			snapshotTx.CreditAmount = -transaction.Amount
		}

		if err := tx.Create(&snapshotTx).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// Generate hash for integrity
	hash, err := s.generateSnapshotHash(&snapshot, transactions)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	snapshot.DataHash = hash

	if err := tx.Model(&snapshot).Update("data_hash", hash).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Load relations
	s.db.Preload("CashBank").Preload("GeneratedByUser").Preload("Transactions").First(&snapshot, snapshot.ID)

	return &snapshot, nil
}

// generateSnapshotHash - Generate SHA-256 hash untuk snapshot integrity
func (s *BankReconciliationService) generateSnapshotHash(snapshot *models.BankReconciliationSnapshot, transactions []models.CashBankTransaction) (string, error) {
	// Create a composite structure for hashing
	hashData := struct {
		CashBankID       uint
		Period           string
		OpeningBalance   float64
		ClosingBalance   float64
		TotalDebit       float64
		TotalCredit      float64
		TransactionCount int
		Transactions     []models.CashBankTransaction
	}{
		CashBankID:       snapshot.CashBankID,
		Period:           snapshot.Period,
		OpeningBalance:   snapshot.OpeningBalance,
		ClosingBalance:   snapshot.ClosingBalance,
		TotalDebit:       snapshot.TotalDebit,
		TotalCredit:      snapshot.TotalCredit,
		TransactionCount: snapshot.TransactionCount,
		Transactions:     transactions,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(hashData)
	if err != nil {
		return "", err
	}

	// Generate SHA-256 hash
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:]), nil
}

// LockSnapshot - Lock snapshot untuk prevent modifications
func (s *BankReconciliationService) LockSnapshot(snapshotID uint, userID uint) error {
	now := time.Now()
	return s.db.Model(&models.BankReconciliationSnapshot{}).
		Where("id = ?", snapshotID).
		Updates(map[string]interface{}{
			"is_locked": true,
			"locked_at": now,
			"locked_by": userID,
		}).Error
}

// GetSnapshots - Get all snapshots for a cash bank account
func (s *BankReconciliationService) GetSnapshots(cashBankID uint) ([]models.BankReconciliationSnapshot, error) {
	var snapshots []models.BankReconciliationSnapshot
	err := s.db.Where("cash_bank_id = ?", cashBankID).
		Preload("CashBank").
		Preload("GeneratedByUser").
		Order("snapshot_date DESC").
		Find(&snapshots).Error
	return snapshots, err
}

// GetSnapshotByID - Get snapshot by ID with all details
func (s *BankReconciliationService) GetSnapshotByID(id uint) (*models.BankReconciliationSnapshot, error) {
	var snapshot models.BankReconciliationSnapshot
	err := s.db.Preload("CashBank").
		Preload("GeneratedByUser").
		Preload("Transactions").
		First(&snapshot, id).Error
	return &snapshot, err
}

// ========== RECONCILIATION OPERATIONS ==========

// CreateReconciliationRequest - Request untuk membuat rekonsiliasi
type CreateReconciliationRequest struct {
	CashBankID           uint   `json:"cash_bank_id" binding:"required"`
	Period               string `json:"period" binding:"required"`
	BaseSnapshotID       uint   `json:"base_snapshot_id" binding:"required"`
	ComparisonSnapshotID *uint  `json:"comparison_snapshot_id"` // Optional, akan di-generate jika null
	Notes                string `json:"notes"`
}

// PerformReconciliation - Melakukan rekonsiliasi antara base snapshot dan current data
func (s *BankReconciliationService) PerformReconciliation(req CreateReconciliationRequest, userID uint) (*models.BankReconciliation, error) {
	// Get base snapshot
	baseSnapshot, err := s.GetSnapshotByID(req.BaseSnapshotID)
	if err != nil {
		return nil, errors.New("base snapshot not found")
	}

	// If comparison snapshot ID not provided, generate new snapshot
	var comparisonSnapshot *models.BankReconciliationSnapshot
	if req.ComparisonSnapshotID == nil {
		// Generate new snapshot for current state
		newSnapshot, err := s.GenerateSnapshot(CreateSnapshotRequest{
			CashBankID: req.CashBankID,
			Period:     req.Period,
			Notes:      "Auto-generated for reconciliation",
		}, userID)
		if err != nil {
			return nil, err
		}
		comparisonSnapshot = newSnapshot
		req.ComparisonSnapshotID = &newSnapshot.ID
	} else {
		comparisonSnapshot, err = s.GetSnapshotByID(*req.ComparisonSnapshotID)
		if err != nil {
			return nil, errors.New("comparison snapshot not found")
		}
	}

	// Compare snapshots
	differences, err := s.compareSnapshots(baseSnapshot, comparisonSnapshot)
	if err != nil {
		return nil, err
	}

	// Calculate variance
	variance := comparisonSnapshot.ClosingBalance - baseSnapshot.ClosingBalance
	transactionVariance := comparisonSnapshot.TransactionCount - baseSnapshot.TransactionCount
	isBalanced := variance == 0 && transactionVariance == 0 && len(differences) == 0

	// Create reconciliation record
	reconciliation := models.BankReconciliation{
		CashBankID:              req.CashBankID,
		Period:                  req.Period,
		BaseSnapshotID:          req.BaseSnapshotID,
		ComparisonSnapshotID:    req.ComparisonSnapshotID,
		ReconciliationDate:      time.Now(),
		ReconciliationBy:        userID,
		BaseBalance:             baseSnapshot.ClosingBalance,
		CurrentBalance:          comparisonSnapshot.ClosingBalance,
		Variance:                variance,
		BaseTransactionCount:    baseSnapshot.TransactionCount,
		CurrentTransactionCount: comparisonSnapshot.TransactionCount,
		TransactionVariance:     transactionVariance,
		MissingTransactions:     s.countDifferencesByType(differences, "MISSING"),
		AddedTransactions:       s.countDifferencesByType(differences, "ADDED"),
		ModifiedTransactions:    s.countDifferencesByType(differences, "MODIFIED"),
		Status:                  "PENDING",
		IsBalanced:              isBalanced,
		Notes:                   req.Notes,
	}

	// Begin transaction
	tx := s.db.Begin()

	// Save reconciliation
	if err := tx.Create(&reconciliation).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Save differences
	for _, diff := range differences {
		diff.ReconciliationID = reconciliation.ID
		if err := tx.Create(&diff).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Load relations
	s.db.Preload("CashBank").
		Preload("BaseSnapshot").
		Preload("ComparisonSnapshot").
		Preload("ReconciliationByUser").
		Preload("Differences").
		First(&reconciliation, reconciliation.ID)

	return &reconciliation, nil
}

// compareSnapshots - Bandingkan 2 snapshot dan temukan perbedaan
func (s *BankReconciliationService) compareSnapshots(base, comparison *models.BankReconciliationSnapshot) ([]models.ReconciliationDifference, error) {
	var differences []models.ReconciliationDifference

	// Create maps for quick lookup
	baseMap := make(map[uint]models.ReconciliationTransactionSnapshot)
	comparisonMap := make(map[uint]models.ReconciliationTransactionSnapshot)

	for _, tx := range base.Transactions {
		baseMap[tx.TransactionID] = tx
	}

	for _, tx := range comparison.Transactions {
		comparisonMap[tx.TransactionID] = tx
	}

	// Find missing transactions (in base but not in comparison)
	for txID, baseTx := range baseMap {
		if _, exists := comparisonMap[txID]; !exists {
			differences = append(differences, models.ReconciliationDifference{
				DifferenceType:   "MISSING",
				Severity:         "HIGH",
				BaseTransactionID: &baseTx.TransactionID,
				Description:      fmt.Sprintf("Transaction %d missing in current data", txID),
				Status:           "PENDING",
			})
		}
	}

	// Find added transactions (in comparison but not in base)
	for txID, compTx := range comparisonMap {
		if _, exists := baseMap[txID]; !exists {
			differences = append(differences, models.ReconciliationDifference{
				DifferenceType:       "ADDED",
				Severity:             "MEDIUM",
				CurrentTransactionID: &compTx.TransactionID,
				Description:          fmt.Sprintf("Transaction %d added after snapshot", txID),
				Status:               "PENDING",
			})
		}
	}

	// Find modified transactions
	for txID, baseTx := range baseMap {
		if compTx, exists := comparisonMap[txID]; exists {
			// Check if amount changed
			if baseTx.Amount != compTx.Amount {
				baseTxID := baseTx.TransactionID
				compTxID := compTx.TransactionID
				differences = append(differences, models.ReconciliationDifference{
					DifferenceType:       "AMOUNT_CHANGE",
					Severity:             "CRITICAL",
					BaseTransactionID:    &baseTxID,
					CurrentTransactionID: &compTxID,
					Field:                "amount",
					OldValue:             fmt.Sprintf("%.2f", baseTx.Amount),
					NewValue:             fmt.Sprintf("%.2f", compTx.Amount),
					AmountDifference:     compTx.Amount - baseTx.Amount,
					Description:          fmt.Sprintf("Amount changed for transaction %d", txID),
					Status:               "PENDING",
				})
			}

			// Check if date changed
			if !baseTx.TransactionDate.Equal(compTx.TransactionDate) {
				baseTxID := baseTx.TransactionID
				compTxID := compTx.TransactionID
				differences = append(differences, models.ReconciliationDifference{
					DifferenceType:       "DATE_CHANGE",
					Severity:             "HIGH",
					BaseTransactionID:    &baseTxID,
					CurrentTransactionID: &compTxID,
					Field:                "transaction_date",
					OldValue:             baseTx.TransactionDate.Format("2006-01-02 15:04:05"),
					NewValue:             compTx.TransactionDate.Format("2006-01-02 15:04:05"),
					Description:          fmt.Sprintf("Transaction date changed for transaction %d", txID),
					Status:               "PENDING",
				})
			}
		}
	}

	return differences, nil
}

// countDifferencesByType - Hitung jumlah perbedaan berdasarkan type
func (s *BankReconciliationService) countDifferencesByType(differences []models.ReconciliationDifference, diffType string) int {
	count := 0
	for _, diff := range differences {
		if diff.DifferenceType == diffType {
			count++
		}
	}
	return count
}

// GetReconciliations - Get all reconciliations for a cash bank account
func (s *BankReconciliationService) GetReconciliations(cashBankID uint) ([]models.BankReconciliation, error) {
	var reconciliations []models.BankReconciliation
	err := s.db.Where("cash_bank_id = ?", cashBankID).
		Preload("CashBank").
		Preload("BaseSnapshot").
		Preload("ComparisonSnapshot").
		Preload("ReconciliationByUser").
		Preload("ReviewedByUser").
		Order("reconciliation_date DESC").
		Find(&reconciliations).Error
	return reconciliations, err
}

// GetReconciliationByID - Get reconciliation by ID with all details
func (s *BankReconciliationService) GetReconciliationByID(id uint) (*models.BankReconciliation, error) {
	var reconciliation models.BankReconciliation
	err := s.db.Preload("CashBank").
		Preload("BaseSnapshot").
		Preload("BaseSnapshot.Transactions").
		Preload("ComparisonSnapshot").
		Preload("ComparisonSnapshot.Transactions").
		Preload("ReconciliationByUser").
		Preload("ReviewedByUser").
		Preload("Differences").
		First(&reconciliation, id).Error
	return &reconciliation, err
}

// ApproveReconciliation - Approve reconciliation
func (s *BankReconciliationService) ApproveReconciliation(id uint, userID uint, notes string) error {
	now := time.Now()
	return s.db.Model(&models.BankReconciliation{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        "APPROVED",
			"reviewed_by":   userID,
			"reviewed_at":   now,
			"review_notes":  notes,
			"balance_confirmed": true,
		}).Error
}

// RejectReconciliation - Reject reconciliation
func (s *BankReconciliationService) RejectReconciliation(id uint, userID uint, notes string) error {
	now := time.Now()
	return s.db.Model(&models.BankReconciliation{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       "REJECTED",
			"reviewed_by":  userID,
			"reviewed_at":  now,
			"review_notes": notes,
		}).Error
}

// ========== AUDIT TRAIL OPERATIONS ==========

// LogAuditTrail - Log audit trail untuk perubahan cash bank
func (s *BankReconciliationService) LogAuditTrail(log models.CashBankAuditTrail) error {
	return s.db.Create(&log).Error
}

// GetAuditTrail - Get audit trail for cash bank account
func (s *BankReconciliationService) GetAuditTrail(cashBankID uint, limit int) ([]models.CashBankAuditTrail, error) {
	var logs []models.CashBankAuditTrail
	query := s.db.Where("cash_bank_id = ?", cashBankID).
		Preload("User").
		Preload("ApprovedByUser").
		Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&logs).Error
	return logs, err
}
