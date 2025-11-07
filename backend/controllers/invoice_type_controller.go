package controllers

import (
	"log"
	"strconv"
	"time"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"app-sistem-akuntansi/utils"

	"github.com/gin-gonic/gin"
)

type InvoiceTypeController struct {
	invoiceTypeService   *services.InvoiceTypeService
	invoiceNumberService *services.InvoiceNumberService
}

func NewInvoiceTypeController(invoiceTypeService *services.InvoiceTypeService, invoiceNumberService *services.InvoiceNumberService) *InvoiceTypeController {
	return &InvoiceTypeController{
		invoiceTypeService:   invoiceTypeService,
		invoiceNumberService: invoiceNumberService,
	}
}

// GetInvoiceTypes godoc
// @Summary List invoice types
// @Description Get all invoice types, with optional active_only=true to filter only active ones
// @Tags Invoice Types
// @Accept json
// @Produce json
// @Security Bearer
// @Param active_only query bool false "Filter only active types" default(false)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/invoice-types [get]
func (c *InvoiceTypeController) GetInvoiceTypes(ctx *gin.Context) {
	log.Printf("📋 Getting invoice types list")
	
	// Check for active_only query parameter
	activeOnly := ctx.Query("active_only") == "true"
	
	invoiceTypes, err := c.invoiceTypeService.GetInvoiceTypes(activeOnly)
	if err != nil {
		log.Printf("❌ Failed to get invoice types: %v", err)
		utils.SendInternalError(ctx, "Failed to retrieve invoice types", err.Error())
		return
	}

	log.Printf("✅ Retrieved %d invoice types", len(invoiceTypes))
	utils.SendSuccess(ctx, "Invoice types retrieved successfully", invoiceTypes)
}

// GetInvoiceType godoc
// @Summary Get invoice type
// @Description Get a single invoice type by ID
// @Tags Invoice Types
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Invoice Type ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /api/v1/invoice-types/{id} [get]
func (c *InvoiceTypeController) GetInvoiceType(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		log.Printf("❌ Invalid invoice type ID parameter: %v", err)
		utils.SendValidationError(ctx, "Invalid invoice type ID", map[string]string{
			"id": "Invoice type ID must be a valid positive number",
		})
		return
	}

	log.Printf("🔍 Getting invoice type details for ID: %d", id)
	
	invoiceType, err := c.invoiceTypeService.GetInvoiceTypeByID(uint(id))
	if err != nil {
		log.Printf("❌ Invoice type %d not found: %v", id, err)
		utils.SendNotFound(ctx, "Invoice type not found")
		return
	}

	log.Printf("✅ Retrieved invoice type %d details successfully", id)
	utils.SendSuccess(ctx, "Invoice type retrieved successfully", invoiceType)
}

// CreateInvoiceType godoc
// @Summary Create invoice type
// @Description Create a new invoice type
// @Tags Invoice Types
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.InvoiceTypeCreateRequest true "Invoice type data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/invoice-types [post]
func (c *InvoiceTypeController) CreateInvoiceType(ctx *gin.Context) {
	log.Printf("🆕 Creating new invoice type")
	
	var request models.InvoiceTypeCreateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		log.Printf("❌ Invalid invoice type creation request: %v", err)
		utils.SendValidationError(ctx, "Invalid invoice type data", map[string]string{
			"request": "Please check the request format and required fields",
		})
		return
	}

	// Get user ID from context
	userIDInterface, exists := ctx.Get("user_id")
	if !exists {
		log.Printf("❌ User authentication missing for invoice type creation")
		utils.SendUnauthorized(ctx, "User authentication required")
		return
	}
	userID, ok := userIDInterface.(uint)
	if !ok {
		log.Printf("❌ Invalid user ID type: %T", userIDInterface)
		utils.SendUnauthorized(ctx, "Invalid user authentication")
		return
	}

	log.Printf("📄 Creating invoice type '%s' (%s) by user %d", request.Name, request.Code, userID)
	
	invoiceType, err := c.invoiceTypeService.CreateInvoiceType(request, userID)
	if err != nil {
		log.Printf("❌ Failed to create invoice type: %v", err)
		utils.SendBusinessRuleError(ctx, "Failed to create invoice type", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Invoice type created successfully: %d", invoiceType.ID)
	utils.SendCreated(ctx, "Invoice type created successfully", invoiceType)
}

// UpdateInvoiceType godoc
// @Summary Update invoice type
// @Description Update an existing invoice type
// @Tags Invoice Types
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Invoice Type ID"
// @Param request body models.InvoiceTypeUpdateRequest true "Invoice type update data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/invoice-types/{id} [put]
func (c *InvoiceTypeController) UpdateInvoiceType(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.SendValidationError(ctx, "Invalid invoice type ID", map[string]string{
			"id": "Invoice type ID must be a valid positive number",
		})
		return
	}

	log.Printf("📝 Updating invoice type %d", id)
	
	var request models.InvoiceTypeUpdateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		log.Printf("❌ Invalid invoice type update request: %v", err)
		utils.SendValidationError(ctx, "Invalid update data", map[string]string{
			"request": "Please check the request format",
		})
		return
	}

	invoiceType, err := c.invoiceTypeService.UpdateInvoiceType(uint(id), request)
	if err != nil {
		log.Printf("❌ Failed to update invoice type %d: %v", id, err)
		utils.SendBusinessRuleError(ctx, "Failed to update invoice type", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Invoice type %d updated successfully", id)
	utils.SendSuccess(ctx, "Invoice type updated successfully", invoiceType)
}

// DeleteInvoiceType godoc
// @Summary Delete invoice type
// @Description Delete an invoice type (only if not used)
// @Tags Invoice Types
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Invoice Type ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/invoice-types/{id} [delete]
func (c *InvoiceTypeController) DeleteInvoiceType(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.SendValidationError(ctx, "Invalid invoice type ID", map[string]string{
			"id": "Invoice type ID must be a valid positive number",
		})
		return
	}

	log.Printf("🗑️ Deleting invoice type %d", id)
	
	if err := c.invoiceTypeService.DeleteInvoiceType(uint(id)); err != nil {
		log.Printf("❌ Failed to delete invoice type %d: %v", id, err)
		utils.SendBusinessRuleError(ctx, "Failed to delete invoice type", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Invoice type %d deleted successfully", id)
	utils.SendSuccess(ctx, "Invoice type deleted successfully", nil)
}

// ToggleInvoiceType godoc
// @Summary Toggle active status
// @Description Toggle the active status of an invoice type
// @Tags Invoice Types
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Invoice Type ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/invoice-types/{id}/toggle [post]
func (c *InvoiceTypeController) ToggleInvoiceType(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.SendValidationError(ctx, "Invalid invoice type ID", map[string]string{
			"id": "Invoice type ID must be a valid positive number",
		})
		return
	}

	log.Printf("🔄 Toggling invoice type %d status", id)
	
	invoiceType, err := c.invoiceTypeService.ToggleInvoiceType(uint(id))
	if err != nil {
		log.Printf("❌ Failed to toggle invoice type %d status: %v", id, err)
		utils.SendBusinessRuleError(ctx, "Failed to toggle invoice type status", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	status := "activated"
	if !invoiceType.IsActive {
		status = "deactivated"
	}

	log.Printf("✅ Invoice type %d %s successfully", id, status)
	utils.SendSuccess(ctx, "Invoice type "+status+" successfully", invoiceType)
}

// GetActiveInvoiceTypes godoc
// @Summary List active invoice types
// @Description Get only active invoice types (for dropdowns)
// @Tags Invoice Types
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/invoice-types/active [get]
func (c *InvoiceTypeController) GetActiveInvoiceTypes(ctx *gin.Context) {
	log.Printf("📋 Getting active invoice types for dropdown")
	
	invoiceTypes, err := c.invoiceTypeService.GetActiveInvoiceTypes()
	if err != nil {
		log.Printf("❌ Failed to get active invoice types: %v", err)
		utils.SendInternalError(ctx, "Failed to retrieve active invoice types", err.Error())
		return
	}

	log.Printf("✅ Retrieved %d active invoice types", len(invoiceTypes))
	utils.SendSuccess(ctx, "Active invoice types retrieved successfully", invoiceTypes)
}

// PreviewInvoiceNumber godoc
// @Summary Preview next invoice number
// @Description Preview what the next invoice number would be for a given type and date
// @Tags Invoice Types
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.InvoiceNumberRequest true "Invoice number preview request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/invoice-types/preview-number [post]
func (c *InvoiceTypeController) PreviewInvoiceNumber(ctx *gin.Context) {
	var request models.InvoiceNumberRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		utils.SendValidationError(ctx, "Invalid request data", map[string]string{
			"request": "Please provide invoice_type_id and date",
		})
		return
	}

	log.Printf("🔍 Previewing invoice number for type %d, date %s", request.InvoiceTypeID, request.Date.Format("2006-01-02"))
	
	preview, err := c.invoiceNumberService.PreviewInvoiceNumber(request.InvoiceTypeID, request.Date)
	if err != nil {
		log.Printf("❌ Failed to preview invoice number: %v", err)
		utils.SendBusinessRuleError(ctx, "Failed to preview invoice number", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Invoice number preview generated: %s", preview.InvoiceNumber)
	utils.SendSuccess(ctx, "Invoice number preview generated successfully", preview)
}

// PreviewInvoiceNumberByID godoc
// @Summary Preview next invoice number by ID
// @Description Preview the next invoice number using path param and optional date query (?date=YYYY-MM-DD)
// @Tags Invoice Types
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Invoice Type ID"
// @Param date query string false "Custom date (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/invoice-types/{id}/preview [get]
func (c *InvoiceTypeController) PreviewInvoiceNumberByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil || id == 0 {
		utils.SendValidationError(ctx, "Invalid invoice type ID", map[string]string{
			"id": "Invoice type ID must be a valid positive number",
		})
		return
	}

	dateStr := ctx.Query("date")
	var date time.Time
	if dateStr == "" {
		date = time.Now()
	} else {
		parsed, perr := time.Parse("2006-01-02", dateStr)
		if perr != nil {
			utils.SendValidationError(ctx, "Invalid date format", map[string]string{
				"date": "Use YYYY-MM-DD format",
			})
			return
		}
		date = parsed
	}

	log.Printf("🔍 Previewing invoice number for type %d via GET, date %s", id, date.Format("2006-01-02"))
	preview, svcErr := c.invoiceNumberService.PreviewInvoiceNumber(uint(id), date)
	if svcErr != nil {
		log.Printf("❌ Failed to preview invoice number: %v", svcErr)
		utils.SendBusinessRuleError(ctx, "Failed to preview invoice number", map[string]interface{}{
			"details": svcErr.Error(),
		})
		return
	}

	log.Printf("✅ Invoice number preview generated: %s", preview.InvoiceNumber)
	utils.SendSuccess(ctx, "Invoice number preview generated successfully", preview)
}

// GetCounterHistory godoc
// @Summary Get counter history
// @Description Get counter history for a specific invoice type
// @Tags Invoice Types
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Invoice Type ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/invoice-types/{id}/counter-history [get]
func (c *InvoiceTypeController) GetCounterHistory(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.SendValidationError(ctx, "Invalid invoice type ID", map[string]string{
			"id": "Invoice type ID must be a valid positive number",
		})
		return
	}

	log.Printf("📊 Getting counter history for invoice type %d", id)
	
	history, err := c.invoiceNumberService.GetCounterHistory(uint(id))
	if err != nil {
		log.Printf("❌ Failed to get counter history: %v", err)
		utils.SendInternalError(ctx, "Failed to retrieve counter history", err.Error())
		return
	}

	log.Printf("✅ Retrieved counter history with %d entries", len(history))
	utils.SendSuccess(ctx, "Counter history retrieved successfully", history)
}

// ResetCounterRequest represents the request to reset invoice counter
type ResetCounterRequest struct {
	Year    int `json:"year" binding:"required,min=2020,max=2050" example:"2024"`
	Counter int `json:"counter" binding:"required,min=0" example:"0"`
}

// ResetCounterForYear godoc
// @Summary Reset invoice counter
// @Description Reset counter for a specific invoice type and year
// @Tags Invoice Types
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Invoice Type ID"
// @Param request body ResetCounterRequest true "Year and new counter value"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/invoice-types/{id}/reset-counter [post]
func (c *InvoiceTypeController) ResetCounterForYear(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.SendValidationError(ctx, "Invalid invoice type ID", map[string]string{
			"id": "Invoice type ID must be a valid positive number",
		})
		return
	}

	var request struct {
		Year     int `json:"year" binding:"required,min=2020,max=2050"`
		Counter  int `json:"counter" binding:"required,min=0"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		log.Printf("❌ Invalid counter reset request: %v", err)
		utils.SendValidationError(ctx, "Invalid request data", map[string]string{
			"year":    "Year must be between 2020 and 2050",
			"counter": "Counter must be a non-negative number",
		})
		return
	}

	log.Printf("🔄 Resetting counter for invoice type %d, year %d to %d", id, request.Year, request.Counter)
	
	err = c.invoiceNumberService.ResetCounterForYear(uint(id), request.Year, request.Counter)
	if err != nil {
		log.Printf("❌ Failed to reset counter: %v", err)
		utils.SendBusinessRuleError(ctx, "Failed to reset counter", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	// Get preview of next invoice number
	preview, err := c.invoiceNumberService.PreviewInvoiceNumber(uint(id), time.Now())
	if err != nil {
		log.Printf("⚠️ Counter reset successful but failed to generate preview: %v", err)
		utils.SendSuccess(ctx, "Counter reset successfully", map[string]interface{}{
			"invoice_type_id": uint(id),
			"year":            request.Year,
			"new_counter":     request.Counter,
		})
		return
	}

	log.Printf("✅ Counter reset successfully. Next invoice: %s", preview.InvoiceNumber)
	utils.SendSuccess(ctx, "Counter reset successfully", map[string]interface{}{
		"invoice_type_id":      uint(id),
		"year":                 request.Year,
		"new_counter":          request.Counter,
		"next_invoice_preview": preview,
	})
}
