package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"app-sistem-akuntansi/utils"
	"gorm.io/gorm"
)

// TransactionSafetyService provides enhanced transaction safety features
type TransactionSafetyService struct {
	db     *gorm.DB
	logger *utils.JournalLogger
}

// NewTransactionSafetyService creates a new transaction safety service
func NewTransactionSafetyService(db *gorm.DB) *TransactionSafetyService {
	return &TransactionSafetyService{
		db:     db,
		logger: utils.NewJournalLogger(db),
	}
}

// TransactionConfig defines configuration for safe transactions
type TransactionConfig struct {
	Timeout         time.Duration
	RetryCount      int
	IsolationLevel  sql.IsolationLevel
	ReadOnly        bool
	DeferConstraints bool
}

// DefaultTransactionConfig returns default safe transaction configuration
func DefaultTransactionConfig() TransactionConfig {
	return TransactionConfig{
		Timeout:         30 * time.Second,
		RetryCount:      3,
		IsolationLevel:  sql.LevelReadCommitted,
		ReadOnly:        false,
		DeferConstraints: false,
	}
}

// JournalTransactionConfig returns configuration optimized for journal operations
func JournalTransactionConfig() TransactionConfig {
	return TransactionConfig{
		Timeout:         2 * time.Minute, // Longer timeout for journal operations
		RetryCount:      5,
		IsolationLevel:  sql.LevelReadCommitted,
		ReadOnly:        false,
		DeferConstraints: true, // Defer constraints for complex journal entries
	}
}

// ReportingTransactionConfig returns configuration for read-only reporting operations
func ReportingTransactionConfig() TransactionConfig {
	return TransactionConfig{
		Timeout:         5 * time.Minute, // Longer timeout for complex reports
		RetryCount:      2,
		IsolationLevel:  sql.LevelRepeatableRead,
		ReadOnly:        true,
		DeferConstraints: false,
	}
}

// ExecuteInSafeTransaction executes a function within a safe transaction context
func (tss *TransactionSafetyService) ExecuteInSafeTransaction(
	ctx context.Context,
	config TransactionConfig,
	fn func(tx *gorm.DB) error,
) error {
	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	userID, _ := utils.GetUserIDFromContext(ctx)
	
	// Log transaction start
	tss.logger.LogProcessingInfo(ctx, "Starting safe transaction", map[string]interface{}{
		"timeout_seconds":  config.Timeout.Seconds(),
		"retry_count":      config.RetryCount,
		"isolation_level":  config.IsolationLevel.String(),
		"read_only":        config.ReadOnly,
		"defer_constraints": config.DeferConstraints,
		"user_id":          userID,
	})

	var lastError error
	
	for attempt := 1; attempt <= config.RetryCount; attempt++ {
		err := tss.executeTransactionAttempt(timeoutCtx, config, fn, attempt)
		
		if err == nil {
			// Success
			tss.logger.LogProcessingInfo(ctx, "Safe transaction completed successfully", map[string]interface{}{
				"attempt":     attempt,
				"total_attempts": config.RetryCount,
			})
			return nil
		}
		
		lastError = err
		
		// Check if error is retryable
		if !tss.isRetryableError(err) {
			tss.logger.LogValidationError(ctx, nil, "non_retryable_transaction_error", err)
			break
		}
		
		// Check timeout
		select {
		case <-timeoutCtx.Done():
			tss.logger.LogValidationError(ctx, nil, "transaction_timeout", timeoutCtx.Err())
			return fmt.Errorf("transaction timeout after %v: %v", config.Timeout, timeoutCtx.Err())
		default:
		}
		
		if attempt < config.RetryCount {
			// Calculate backoff delay (exponential backoff with jitter)
			backoffDelay := time.Duration(attempt) * time.Second
			tss.logger.LogWarning(ctx, fmt.Sprintf("Transaction attempt %d failed, retrying in %v", attempt, backoffDelay), map[string]interface{}{
				"attempt": attempt,
				"error":   err.Error(),
				"backoff_delay": backoffDelay.String(),
			})
			
			select {
			case <-timeoutCtx.Done():
				return fmt.Errorf("transaction timeout during retry backoff: %v", timeoutCtx.Err())
			case <-time.After(backoffDelay):
			}
		}
	}
	
	// All attempts failed
	tss.logger.LogValidationError(ctx, nil, "transaction_all_attempts_failed", lastError)
	return fmt.Errorf("transaction failed after %d attempts: %v", config.RetryCount, lastError)
}

// executeTransactionAttempt executes a single transaction attempt
func (tss *TransactionSafetyService) executeTransactionAttempt(
	ctx context.Context,
	config TransactionConfig,
	fn func(tx *gorm.DB) error,
	attempt int,
) error {
	// Begin transaction with timeout context
	tx := tss.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %v", tx.Error)
	}

	// Set isolation level if specified
	if config.IsolationLevel != sql.LevelDefault {
		isolationQuery := tss.getIsolationLevelQuery(config.IsolationLevel)
		if isolationQuery != "" {
			if err := tx.Exec(isolationQuery).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to set isolation level: %v", err)
			}
		}
	}

	// Set read-only mode if specified
	if config.ReadOnly {
		if err := tx.Exec("SET TRANSACTION READ ONLY").Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to set read-only mode: %v", err)
		}
	}

	// Defer constraints if specified
	if config.DeferConstraints {
		if err := tx.Exec("SET CONSTRAINTS ALL DEFERRED").Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to defer constraints: %v", err)
		}
	}

	// Execute the transaction function
	err := fn(tx)
	
	if err != nil {
		// Rollback on error
		if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
			tss.logger.LogValidationError(ctx, nil, "rollback_failed", rollbackErr)
			return fmt.Errorf("transaction failed and rollback failed: original error: %v, rollback error: %v", err, rollbackErr)
		}
		return err
	}

	// Commit transaction
	if commitErr := tx.Commit().Error; commitErr != nil {
		// Commit failed, try to rollback
		if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
			tss.logger.LogValidationError(ctx, nil, "commit_and_rollback_failed", fmt.Errorf("commit error: %v, rollback error: %v", commitErr, rollbackErr))
		}
		return fmt.Errorf("failed to commit transaction: %v", commitErr)
	}

	return nil
}

// isRetryableError determines if an error is retryable
func (tss *TransactionSafetyService) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errorStr := err.Error()
	
	// Common retryable errors
	retryableErrors := []string{
		"deadlock detected",
		"could not serialize access",
		"connection refused",
		"connection reset",
		"timeout",
		"temporary failure",
		"duplicate key value violates unique constraint",
	}

	for _, retryableErr := range retryableErrors {
		if fmt.Sprintf("%v", errorStr) == retryableErr {
			return true
		}
	}

	return false
}

// getIsolationLevelQuery returns the SQL query to set isolation level
func (tss *TransactionSafetyService) getIsolationLevelQuery(level sql.IsolationLevel) string {
	switch level {
	case sql.LevelReadUncommitted:
		return "SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED"
	case sql.LevelReadCommitted:
		return "SET TRANSACTION ISOLATION LEVEL READ COMMITTED"
	case sql.LevelRepeatableRead:
		return "SET TRANSACTION ISOLATION LEVEL REPEATABLE READ"
	case sql.LevelSerializable:
		return "SET TRANSACTION ISOLATION LEVEL SERIALIZABLE"
	default:
		return ""
	}
}

// ExecuteJournalTransaction executes a journal-specific safe transaction
func (tss *TransactionSafetyService) ExecuteJournalTransaction(
	ctx context.Context,
	fn func(tx *gorm.DB) error,
) error {
	config := JournalTransactionConfig()
	return tss.ExecuteInSafeTransaction(ctx, config, fn)
}

// ExecuteReportingTransaction executes a reporting-specific safe transaction
func (tss *TransactionSafetyService) ExecuteReportingTransaction(
	ctx context.Context,
	fn func(tx *gorm.DB) error,
) error {
	config := ReportingTransactionConfig()
	return tss.ExecuteInSafeTransaction(ctx, config, fn)
}

// ExecuteBatchTransaction executes a batch processing safe transaction
func (tss *TransactionSafetyService) ExecuteBatchTransaction(
	ctx context.Context,
	batchSize int,
	fn func(tx *gorm.DB) error,
) error {
	// Adjust timeout based on batch size
	baseTimeout := 30 * time.Second
	timeoutMultiplier := time.Duration(batchSize/100 + 1) // +30s for every 100 items
	adjustedTimeout := baseTimeout * timeoutMultiplier
	
	// Cap at 10 minutes
	if adjustedTimeout > 10*time.Minute {
		adjustedTimeout = 10 * time.Minute
	}

	config := TransactionConfig{
		Timeout:         adjustedTimeout,
		RetryCount:      3,
		IsolationLevel:  sql.LevelReadCommitted,
		ReadOnly:        false,
		DeferConstraints: true,
	}

	tss.logger.LogProcessingInfo(ctx, "Starting batch transaction", map[string]interface{}{
		"batch_size":       batchSize,
		"adjusted_timeout": adjustedTimeout.String(),
	})

	return tss.ExecuteInSafeTransaction(ctx, config, fn)
}

// TransactionStats holds transaction execution statistics
type TransactionStats struct {
	TotalTransactions      int64         `json:"total_transactions"`
	SuccessfulTransactions int64         `json:"successful_transactions"`
	FailedTransactions     int64         `json:"failed_transactions"`
	RetriedTransactions    int64         `json:"retried_transactions"`
	AverageExecutionTime   time.Duration `json:"average_execution_time"`
	TotalExecutionTime     time.Duration `json:"total_execution_time"`
}

// GetTransactionStats returns transaction execution statistics
func (tss *TransactionSafetyService) GetTransactionStats(ctx context.Context, since time.Time) (*TransactionStats, error) {
	// This would typically be implemented with metrics collection
	// For now, return a basic implementation
	stats := &TransactionStats{
		TotalTransactions:      0,
		SuccessfulTransactions: 0,
		FailedTransactions:     0,
		RetriedTransactions:    0,
		AverageExecutionTime:   0,
		TotalExecutionTime:     0,
	}

	// TODO: Implement actual statistics collection from metrics store
	tss.logger.LogProcessingInfo(ctx, "Transaction stats requested", map[string]interface{}{
		"since": since,
	})

	return stats, nil
}