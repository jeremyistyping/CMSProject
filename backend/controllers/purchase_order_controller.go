package controllers

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PurchaseOrderController struct {
	poService    *services.PurchaseOrderService
	poPDFService *services.POPDFService
}

func NewPurchaseOrderController(poService *services.PurchaseOrderService, poPDFService *services.POPDFService) *PurchaseOrderController {
	return &PurchaseOrderController{
		poService:    poService,
		poPDFService: poPDFService,
	}
}

// CreateFromPR creates a PO from an approved PR
// @Summary Create PO from PR
// @Tags Purchase Orders
// @Accept json
// @Produce json
// @Param request body models.CreatePORequest true "Create PO Request"
// @Success 201 {object} models.PurchaseOrder
// @Router /api/purchase-orders/from-pr [post]
func (c *PurchaseOrderController) CreateFromPR(ctx *gin.Context) {
	var req models.CreatePORequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	po, err := c.poService.CreateFromPR(&req, userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, po)
}

// GetAll retrieves all purchase orders
// @Summary Get all purchase orders
// @Tags Purchase Orders
// @Produce json
// @Param project_id query int false "Filter by project ID"
// @Param status query string false "Filter by status"
// @Param vendor_id query int false "Filter by vendor ID"
// @Success 200 {array} models.PurchaseOrder
// @Router /api/purchase-orders [get]
func (c *PurchaseOrderController) GetAll(ctx *gin.Context) {
	filter := make(map[string]interface{})

	if projectID := ctx.Query("project_id"); projectID != "" {
		if id, err := strconv.ParseUint(projectID, 10, 32); err == nil {
			filter["project_id"] = uint(id)
		}
	}
	if status := ctx.Query("status"); status != "" {
		filter["status"] = status
	}
	if vendorID := ctx.Query("vendor_id"); vendorID != "" {
		if id, err := strconv.ParseUint(vendorID, 10, 32); err == nil {
			filter["vendor_id"] = uint(id)
		}
	}

	pos, err := c.poService.GetAll(filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, pos)
}

// GetByID retrieves a purchase order by ID
// @Summary Get purchase order by ID
// @Tags Purchase Orders
// @Produce json
// @Param id path int true "Purchase Order ID"
// @Success 200 {object} models.PurchaseOrder
// @Router /api/purchase-orders/{id} [get]
func (c *PurchaseOrderController) GetByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	po, err := c.poService.GetByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Purchase order not found"})
		return
	}

	ctx.JSON(http.StatusOK, po)
}

// GetByPRID retrieves purchase orders by PR ID
// @Summary Get purchase orders by PR ID
// @Tags Purchase Orders
// @Produce json
// @Param pr_id path int true "Purchase Request ID"
// @Success 200 {array} models.PurchaseOrder
// @Router /api/purchase-orders/by-pr/{pr_id} [get]
func (c *PurchaseOrderController) GetByPRID(ctx *gin.Context) {
	prID, err := strconv.ParseUint(ctx.Param("pr_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid PR ID"})
		return
	}

	pos, err := c.poService.GetByPRID(uint(prID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, pos)
}

// SendPO marks a PO as sent
// @Summary Send PO to vendor
// @Tags Purchase Orders
// @Param id path int true "Purchase Order ID"
// @Success 200 {object} map[string]string
// @Router /api/purchase-orders/{id}/send [post]
func (c *PurchaseOrderController) SendPO(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := c.poService.SendPO(uint(id)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "PO sent successfully"})
}

// Delete deletes a PO
// @Summary Delete purchase order
// @Tags Purchase Orders
// @Param id path int true "Purchase Order ID"
// @Success 200 {object} map[string]string
// @Router /api/purchase-orders/{id} [delete]
func (c *PurchaseOrderController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := c.poService.Delete(uint(id)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "PO deleted successfully"})
}

// CreateGoodsReceipt creates a goods receipt for a PO
// @Summary Create goods receipt
// @Tags Purchase Orders
// @Accept json
// @Produce json
// @Param request body models.CreateGRRequest true "Create GR Request"
// @Success 201 {object} models.GoodsReceipt
// @Router /api/purchase-orders/goods-receipt [post]
func (c *PurchaseOrderController) CreateGoodsReceipt(ctx *gin.Context) {
	var req models.CreateGRRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	gr, err := c.poService.CreateGoodsReceipt(&req, userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gr)
}

// GetGoodsReceiptsByPOID retrieves goods receipts for a PO
// @Summary Get goods receipts by PO ID
// @Tags Purchase Orders
// @Produce json
// @Param id path int true "Purchase Order ID"
// @Success 200 {array} models.GoodsReceipt
// @Router /api/purchase-orders/{id}/goods-receipts [get]
func (c *PurchaseOrderController) GetGoodsReceiptsByPOID(ctx *gin.Context) {
	poID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid PO ID"})
		return
	}

	grs, err := c.poService.GetGoodsReceiptsByPOID(uint(poID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, grs)
}

// DownloadPOPDF generates and downloads PO as PDF
// @Summary Download PO as PDF
// @Tags Purchase Orders
// @Produce application/pdf
// @Param id path int true "Purchase Order ID"
// @Success 200 {file} binary
// @Router /api/purchase-orders/{id}/pdf [get]
func (c *PurchaseOrderController) DownloadPOPDF(ctx *gin.Context) {
	poID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid PO ID"})
		return
	}

	// Get PO to get the code for filename
	po, err := c.poService.GetByID(uint(poID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Purchase order not found"})
		return
	}

	pdfBytes, err := c.poPDFService.GeneratePOPDF(uint(poID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Header("Content-Type", "application/pdf")
	ctx.Header("Content-Disposition", "attachment; filename=PO_"+po.Code+".pdf")
	ctx.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// DownloadGRPDF generates and downloads Goods Receipt as PDF
// @Summary Download Goods Receipt as PDF
// @Tags Purchase Orders
// @Produce application/pdf
// @Param id path int true "Purchase Order ID"
// @Success 200 {file} binary
// @Router /api/purchase-orders/{id}/goods-receipt-pdf [get]
func (c *PurchaseOrderController) DownloadGRPDF(ctx *gin.Context) {
	poID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid PO ID"})
		return
	}

	// Get PO to get the code for filename
	po, err := c.poService.GetByID(uint(poID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Purchase order not found"})
		return
	}

	pdfBytes, err := c.poPDFService.GenerateGRPDF(uint(poID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Header("Content-Type", "application/pdf")
	ctx.Header("Content-Disposition", "attachment; filename=GR_"+po.Code+".pdf")
	ctx.Data(http.StatusOK, "application/pdf", pdfBytes)
}
