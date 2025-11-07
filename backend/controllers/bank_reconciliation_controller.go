package controllers

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type BankReconciliationController struct {
	reconciliationService *services.BankReconciliationService
}

func NewBankReconciliationController(reconciliationService *services.BankReconciliationService) *BankReconciliationController {
	return &BankReconciliationController{
		reconciliationService: reconciliationService,
	}
}

// ========== SNAPSHOT ENDPOINTS ==========

// GenerateSnapshot godoc
// @Summary Generate bank statement snapshot
// @Description Create a frozen snapshot of bank transactions for a specific period
// @Tags BankReconciliation
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body services.CreateSnapshotRequest true "Snapshot request"
// @Success 201 {object} models.BankReconciliationSnapshot
// @Router /api/bank-reconciliation/snapshots [post]
func (c *BankReconciliationController) GenerateSnapshot(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var request services.CreateSnapshotRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	snapshot, err := c.reconciliationService.GenerateSnapshot(request, userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to generate snapshot",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Snapshot generated successfully",
		"data":    snapshot,
	})
}

// GetSnapshots godoc
// @Summary Get all snapshots for a cash bank account
// @Description Retrieve all snapshots for a specific cash bank account
// @Tags BankReconciliation
// @Accept json
// @Produce json
// @Security Bearer
// @Param cash_bank_id query int true "Cash Bank ID"
// @Success 200 {array} models.BankReconciliationSnapshot
// @Router /api/bank-reconciliation/snapshots [get]
func (c *BankReconciliationController) GetSnapshots(ctx *gin.Context) {
	cashBankIDStr := ctx.Query("cash_bank_id")
	if cashBankIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "cash_bank_id parameter is required",
		})
		return
	}

	cashBankID, err := strconv.ParseUint(cashBankIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid cash_bank_id",
		})
		return
	}

	snapshots, err := c.reconciliationService.GetSnapshots(uint(cashBankID))
	if err != nil {
		// Log detailed error
		log.Printf("❌ Error retrieving snapshots for cash_bank_id=%d: %v", cashBankID, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve snapshots",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    snapshots,
	})
}

// GetSnapshotByID godoc
// @Summary Get snapshot by ID
// @Description Retrieve snapshot details with transactions
// @Tags BankReconciliation
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Snapshot ID"
// @Success 200 {object} models.BankReconciliationSnapshot
// @Router /api/bank-reconciliation/snapshots/{id} [get]
func (c *BankReconciliationController) GetSnapshotByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid snapshot ID",
		})
		return
	}

	snapshot, err := c.reconciliationService.GetSnapshotByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Snapshot not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    snapshot,
	})
}

// LockSnapshot godoc
// @Summary Lock snapshot to prevent modifications
// @Description Lock a snapshot after period close to ensure data integrity
// @Tags BankReconciliation
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Snapshot ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/bank-reconciliation/snapshots/{id}/lock [post]
func (c *BankReconciliationController) LockSnapshot(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid snapshot ID",
		})
		return
	}

	if err := c.reconciliationService.LockSnapshot(uint(id), userID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to lock snapshot",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Snapshot locked successfully",
	})
}

// ========== RECONCILIATION ENDPOINTS ==========

// PerformReconciliation godoc
// @Summary Perform bank reconciliation
// @Description Compare base snapshot with current data to detect changes
// @Tags BankReconciliation
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body services.CreateReconciliationRequest true "Reconciliation request"
// @Success 201 {object} models.BankReconciliation
// @Router /api/bank-reconciliation/reconcile [post]
func (c *BankReconciliationController) PerformReconciliation(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var request services.CreateReconciliationRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	reconciliation, err := c.reconciliationService.PerformReconciliation(request, userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to perform reconciliation",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Reconciliation completed",
		"data":    reconciliation,
	})
}

// GetReconciliations godoc
// @Summary Get all reconciliations for a cash bank account
// @Description Retrieve all reconciliation records for a specific cash bank account
// @Tags BankReconciliation
// @Accept json
// @Produce json
// @Security Bearer
// @Param cash_bank_id query int true "Cash Bank ID"
// @Success 200 {array} models.BankReconciliation
// @Router /api/bank-reconciliation/reconciliations [get]
func (c *BankReconciliationController) GetReconciliations(ctx *gin.Context) {
	cashBankIDStr := ctx.Query("cash_bank_id")
	if cashBankIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "cash_bank_id parameter is required",
		})
		return
	}

	cashBankID, err := strconv.ParseUint(cashBankIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid cash_bank_id",
		})
		return
	}

	reconciliations, err := c.reconciliationService.GetReconciliations(uint(cashBankID))
	if err != nil {
		// Log detailed error
		log.Printf("❌ Error retrieving reconciliations for cash_bank_id=%d: %v", cashBankID, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve reconciliations",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    reconciliations,
	})
}

// GetReconciliationByID godoc
// @Summary Get reconciliation by ID
// @Description Retrieve detailed reconciliation report with differences
// @Tags BankReconciliation
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Reconciliation ID"
// @Success 200 {object} models.BankReconciliation
// @Router /api/bank-reconciliation/reconciliations/{id} [get]
func (c *BankReconciliationController) GetReconciliationByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid reconciliation ID",
		})
		return
	}

	reconciliation, err := c.reconciliationService.GetReconciliationByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Reconciliation not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    reconciliation,
	})
}

// ApproveReconciliation godoc
// @Summary Approve reconciliation
// @Description Approve a reconciliation report (manager/supervisor only)
// @Tags BankReconciliation
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Reconciliation ID"
// @Param request body map[string]string true "Approval notes"
// @Success 200 {object} map[string]interface{}
// @Router /api/bank-reconciliation/reconciliations/{id}/approve [post]
func (c *BankReconciliationController) ApproveReconciliation(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid reconciliation ID",
		})
		return
	}

	var request struct {
		Notes string `json:"notes"`
	}
	if err := ctx.ShouldBindJSON(&request); err != nil {
		// Notes are optional, so we can continue even if binding fails
		request.Notes = ""
	}

	if err := c.reconciliationService.ApproveReconciliation(uint(id), userID, request.Notes); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to approve reconciliation",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Reconciliation approved successfully",
	})
}

// RejectReconciliation godoc
// @Summary Reject reconciliation
// @Description Reject a reconciliation report (manager/supervisor only)
// @Tags BankReconciliation
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Reconciliation ID"
// @Param request body map[string]string true "Rejection notes (required)"
// @Success 200 {object} map[string]interface{}
// @Router /api/bank-reconciliation/reconciliations/{id}/reject [post]
func (c *BankReconciliationController) RejectReconciliation(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid reconciliation ID",
		})
		return
	}

	var request struct {
		Notes string `json:"notes" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Rejection notes are required",
		})
		return
	}

	if err := c.reconciliationService.RejectReconciliation(uint(id), userID, request.Notes); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to reject reconciliation",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Reconciliation rejected",
	})
}

// ========== AUDIT TRAIL ENDPOINTS ==========

// GetAuditTrail godoc
// @Summary Get audit trail for cash bank account
// @Description Retrieve complete audit trail of all changes
// @Tags BankReconciliation
// @Accept json
// @Produce json
// @Security Bearer
// @Param cash_bank_id query int true "Cash Bank ID"
// @Param limit query int false "Limit number of records"
// @Success 200 {array} models.CashBankAuditTrail
// @Router /api/bank-reconciliation/audit-trail [get]
func (c *BankReconciliationController) GetAuditTrail(ctx *gin.Context) {
	cashBankIDStr := ctx.Query("cash_bank_id")
	if cashBankIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "cash_bank_id parameter is required",
		})
		return
	}

	cashBankID, err := strconv.ParseUint(cashBankIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid cash_bank_id",
		})
		return
	}

	limitStr := ctx.DefaultQuery("limit", "100")
	limit, _ := strconv.Atoi(limitStr)

	logs, err := c.reconciliationService.GetAuditTrail(uint(cashBankID), limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve audit trail",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
	})
}

// LogAuditTrail godoc
// @Summary Log audit trail entry
// @Description Create audit trail entry for cash bank changes
// @Tags BankReconciliation
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.CashBankAuditTrail true "Audit log entry"
// @Success 201 {object} map[string]interface{}
// @Router /api/bank-reconciliation/audit-trail [post]
func (c *BankReconciliationController) LogAuditTrail(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var log models.CashBankAuditTrail
	if err := ctx.ShouldBindJSON(&log); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Set user ID from context
	log.UserID = userID

	// Get IP and User Agent
	log.IPAddress = ctx.ClientIP()
	log.UserAgent = ctx.Request.UserAgent()

	if err := c.reconciliationService.LogAuditTrail(log); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to log audit trail",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Audit trail logged successfully",
	})
}
