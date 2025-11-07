package utils

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// APIResponse represents a standardized API response structure
type APIResponse struct {
	Status    string      `json:"status"`           // "success" or "error"
	Message   string      `json:"message"`          // Human readable message
	Data      interface{} `json:"data,omitempty"`   // Response data (only for success)
	Error     *APIError   `json:"error,omitempty"`  // Error details (only for errors)
	Meta      *APIMeta    `json:"meta,omitempty"`   // Additional metadata
	Timestamp time.Time   `json:"timestamp"`        // Response timestamp
}

// APIError represents detailed error information
type APIError struct {
	Code        string                 `json:"code"`                  // Error code (e.g., "VALIDATION_ERROR")
	Message     string                 `json:"message"`               // Error message
	Details     interface{}            `json:"details,omitempty"`     // Additional error details
	Field       string                 `json:"field,omitempty"`       // Field causing error (for validation)
	Validation  map[string]string      `json:"validation,omitempty"`  // Validation errors
	Context     map[string]interface{} `json:"context,omitempty"`     // Error context
	TraceID     string                 `json:"trace_id,omitempty"`    // Request trace ID
}

// APIMeta represents response metadata
type APIMeta struct {
	RequestID    string                 `json:"request_id,omitempty"`
	Version      string                 `json:"version,omitempty"`
	ProcessingTime string               `json:"processing_time,omitempty"`
	Pagination   *PaginationMeta        `json:"pagination,omitempty"`
	Additional   map[string]interface{} `json:"additional,omitempty"`
}

// PaginationMeta represents pagination information
type PaginationMeta struct {
	Page        int   `json:"page"`
	Limit       int   `json:"limit"`
	Total       int64 `json:"total"`
	TotalPages  int   `json:"total_pages"`
	HasNext     bool  `json:"has_next"`
	HasPrevious bool  `json:"has_previous"`
}

// Error codes constants
const (
	// Client Errors (4xx)
	ErrorCodeValidation       = "VALIDATION_ERROR"
	ErrorCodeNotFound         = "NOT_FOUND"
	ErrorCodeUnauthorized     = "UNAUTHORIZED"
	ErrorCodeForbidden        = "FORBIDDEN"
	ErrorCodeConflict         = "CONFLICT"
	ErrorCodeBadRequest       = "BAD_REQUEST"
	ErrorCodeMethodNotAllowed = "METHOD_NOT_ALLOWED"
	ErrorCodeTooManyRequests  = "TOO_MANY_REQUESTS"

	// Server Errors (5xx)
	ErrorCodeInternalServer = "INTERNAL_SERVER_ERROR"
	ErrorCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	ErrorCodeTimeout = "TIMEOUT"
	ErrorCodeDatabaseError = "DATABASE_ERROR"

	// Business Logic Errors
	ErrorCodeBusinessRule     = "BUSINESS_RULE_VIOLATION"
	ErrorCodeInsufficientFunds = "INSUFFICIENT_FUNDS"
	ErrorCodeInvalidStatus    = "INVALID_STATUS"
	ErrorCodeOperationNotAllowed = "OPERATION_NOT_ALLOWED"
	ErrorCodeDataIntegrity    = "DATA_INTEGRITY_ERROR"

	// Sales Specific Errors
	ErrorCodeSaleNotFound           = "SALE_NOT_FOUND"
	ErrorCodeInvalidSaleStatus      = "INVALID_SALE_STATUS"
	ErrorCodePaymentExceedsAmount   = "PAYMENT_EXCEEDS_AMOUNT"
	ErrorCodeNoOutstandingAmount    = "NO_OUTSTANDING_AMOUNT"
	ErrorCodeInventoryInsufficient  = "INVENTORY_INSUFFICIENT"
	ErrorCodeCustomerCreditLimit    = "CUSTOMER_CREDIT_LIMIT_EXCEEDED"
)

// ResponseBuilder provides a fluent interface for building API responses
type ResponseBuilder struct {
	response *APIResponse
}

// NewResponseBuilder creates a new response builder
func NewResponseBuilder() *ResponseBuilder {
	return &ResponseBuilder{
		response: &APIResponse{
			Timestamp: time.Now(),
		},
	}
}

// Success creates a success response
func (rb *ResponseBuilder) Success() *ResponseBuilder {
	rb.response.Status = "success"
	return rb
}

// Error creates an error response
func (rb *ResponseBuilder) Error() *ResponseBuilder {
	rb.response.Status = "error"
	return rb
}

// Message sets the response message
func (rb *ResponseBuilder) Message(message string) *ResponseBuilder {
	rb.response.Message = message
	return rb
}

// Data sets the response data
func (rb *ResponseBuilder) Data(data interface{}) *ResponseBuilder {
	rb.response.Data = data
	return rb
}

// ErrorCode sets the error code
func (rb *ResponseBuilder) ErrorCode(code string) *ResponseBuilder {
	if rb.response.Error == nil {
		rb.response.Error = &APIError{}
	}
	rb.response.Error.Code = code
	return rb
}

// ErrorMessage sets the error message
func (rb *ResponseBuilder) ErrorMessage(message string) *ResponseBuilder {
	if rb.response.Error == nil {
		rb.response.Error = &APIError{}
	}
	rb.response.Error.Message = message
	return rb
}

// ErrorDetails sets error details
func (rb *ResponseBuilder) ErrorDetails(details interface{}) *ResponseBuilder {
	if rb.response.Error == nil {
		rb.response.Error = &APIError{}
	}
	rb.response.Error.Details = details
	return rb
}

// ErrorField sets the field causing the error
func (rb *ResponseBuilder) ErrorField(field string) *ResponseBuilder {
	if rb.response.Error == nil {
		rb.response.Error = &APIError{}
	}
	rb.response.Error.Field = field
	return rb
}

// ValidationErrors sets validation errors
func (rb *ResponseBuilder) ValidationErrors(validationErrors map[string]string) *ResponseBuilder {
	if rb.response.Error == nil {
		rb.response.Error = &APIError{}
	}
	rb.response.Error.Validation = validationErrors
	return rb
}

// Context sets error context
func (rb *ResponseBuilder) Context(context map[string]interface{}) *ResponseBuilder {
	if rb.response.Error == nil {
		rb.response.Error = &APIError{}
	}
	rb.response.Error.Context = context
	return rb
}

// RequestID sets the request ID
func (rb *ResponseBuilder) RequestID(requestID string) *ResponseBuilder {
	if rb.response.Meta == nil {
		rb.response.Meta = &APIMeta{}
	}
	rb.response.Meta.RequestID = requestID
	return rb
}

// Version sets the API version
func (rb *ResponseBuilder) Version(version string) *ResponseBuilder {
	if rb.response.Meta == nil {
		rb.response.Meta = &APIMeta{}
	}
	rb.response.Meta.Version = version
	return rb
}

// ProcessingTime sets the processing time
func (rb *ResponseBuilder) ProcessingTime(duration time.Duration) *ResponseBuilder {
	if rb.response.Meta == nil {
		rb.response.Meta = &APIMeta{}
	}
	rb.response.Meta.ProcessingTime = duration.String()
	return rb
}

// Pagination sets pagination metadata
func (rb *ResponseBuilder) Pagination(page, limit int, total int64) *ResponseBuilder {
	if rb.response.Meta == nil {
		rb.response.Meta = &APIMeta{}
	}
	
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	if totalPages == 0 {
		totalPages = 1
	}
	
	rb.response.Meta.Pagination = &PaginationMeta{
		Page:        page,
		Limit:       limit,
		Total:       total,
		TotalPages:  totalPages,
		HasNext:     page < totalPages,
		HasPrevious: page > 1,
	}
	return rb
}

// AdditionalMeta sets additional metadata
func (rb *ResponseBuilder) AdditionalMeta(key string, value interface{}) *ResponseBuilder {
	if rb.response.Meta == nil {
		rb.response.Meta = &APIMeta{}
	}
	if rb.response.Meta.Additional == nil {
		rb.response.Meta.Additional = make(map[string]interface{})
	}
	rb.response.Meta.Additional[key] = value
	return rb
}

// Build returns the final response
func (rb *ResponseBuilder) Build() *APIResponse {
	return rb.response
}

// Send sends the response via gin context
func (rb *ResponseBuilder) Send(c *gin.Context, statusCode int) {
	// Add request ID from context if available
	if requestID, exists := c.Get("request_id"); exists {
		rb.RequestID(requestID.(string))
	}
	
	// Add version from context if available
	if version, exists := c.Get("api_version"); exists {
		rb.Version(version.(string))
	}
	
	c.JSON(statusCode, rb.response)
}

// Convenience functions for common responses

// SendSuccess sends a success response
func SendSuccess(c *gin.Context, message string, data interface{}) {
	NewResponseBuilder().
		Success().
		Message(message).
		Data(data).
		Send(c, http.StatusOK)
}

// SendCreated sends a created response
func SendCreated(c *gin.Context, message string, data interface{}) {
	NewResponseBuilder().
		Success().
		Message(message).
		Data(data).
		Send(c, http.StatusCreated)
}

// SendError sends an error response
func SendError(c *gin.Context, statusCode int, code, message string, details interface{}) {
	NewResponseBuilder().
		Error().
		Message(message).
		ErrorCode(code).
		ErrorMessage(message).
		ErrorDetails(details).
		Send(c, statusCode)
}

// SendValidationError sends a validation error response
func SendValidationError(c *gin.Context, message string, validationErrors map[string]string) {
	NewResponseBuilder().
		Error().
		Message(message).
		ErrorCode(ErrorCodeValidation).
		ErrorMessage(message).
		ValidationErrors(validationErrors).
		Send(c, http.StatusBadRequest)
}

// SendNotFound sends a not found error response
func SendNotFound(c *gin.Context, message string) {
	NewResponseBuilder().
		Error().
		Message(message).
		ErrorCode(ErrorCodeNotFound).
		ErrorMessage(message).
		Send(c, http.StatusNotFound)
}

// SendUnauthorized sends an unauthorized error response
func SendUnauthorized(c *gin.Context, message string) {
	NewResponseBuilder().
		Error().
		Message(message).
		ErrorCode(ErrorCodeUnauthorized).
		ErrorMessage(message).
		Send(c, http.StatusUnauthorized)
}

// SendForbidden sends a forbidden error response
func SendForbidden(c *gin.Context, message string) {
	NewResponseBuilder().
		Error().
		Message(message).
		ErrorCode(ErrorCodeForbidden).
		ErrorMessage(message).
		Send(c, http.StatusForbidden)
}

// SendConflict sends a conflict error response
func SendConflict(c *gin.Context, message string, details interface{}) {
	NewResponseBuilder().
		Error().
		Message(message).
		ErrorCode(ErrorCodeConflict).
		ErrorMessage(message).
		ErrorDetails(details).
		Send(c, http.StatusConflict)
}

// SendInternalError sends an internal server error response
func SendInternalError(c *gin.Context, message string, details interface{}) {
	NewResponseBuilder().
		Error().
		Message(message).
		ErrorCode(ErrorCodeInternalServer).
		ErrorMessage(message).
		ErrorDetails(details).
		Send(c, http.StatusInternalServerError)
}

// SendBusinessRuleError sends a business rule violation error
func SendBusinessRuleError(c *gin.Context, message string, context map[string]interface{}) {
	NewResponseBuilder().
		Error().
		Message(message).
		ErrorCode(ErrorCodeBusinessRule).
		ErrorMessage(message).
		Context(context).
		Send(c, http.StatusBadRequest)
}

// SendPaginatedSuccess sends a paginated success response
func SendPaginatedSuccess(c *gin.Context, message string, data interface{}, page, limit int, total int64) {
	NewResponseBuilder().
		Success().
		Message(message).
		Data(data).
		Pagination(page, limit, total).
		Send(c, http.StatusOK)
}

// Sales-specific convenience functions

// SendSaleNotFound sends sale not found error
func SendSaleNotFound(c *gin.Context, saleID uint) {
	NewResponseBuilder().
		Error().
		Message("Sale not found").
		ErrorCode(ErrorCodeSaleNotFound).
		ErrorMessage("The requested sale could not be found").
		Context(map[string]interface{}{
			"sale_id": saleID,
		}).
		Send(c, http.StatusNotFound)
}

// SendInvalidSaleStatus sends invalid sale status error
func SendInvalidSaleStatus(c *gin.Context, saleID uint, currentStatus, requestedStatus string) {
	NewResponseBuilder().
		Error().
		Message("Invalid sale status for this operation").
		ErrorCode(ErrorCodeInvalidSaleStatus).
		ErrorMessage("The sale status does not allow this operation").
		Context(map[string]interface{}{
			"sale_id":          saleID,
			"current_status":   currentStatus,
			"requested_status": requestedStatus,
		}).
		Send(c, http.StatusBadRequest)
}

// SendPaymentExceedsAmount sends payment exceeds amount error
func SendPaymentExceedsAmount(c *gin.Context, saleID uint, paymentAmount, outstandingAmount float64) {
	NewResponseBuilder().
		Error().
		Message("Payment amount exceeds outstanding amount").
		ErrorCode(ErrorCodePaymentExceedsAmount).
		ErrorMessage("The payment amount cannot exceed the outstanding amount").
		Context(map[string]interface{}{
			"sale_id":            saleID,
			"payment_amount":     paymentAmount,
			"outstanding_amount": outstandingAmount,
		}).
		Send(c, http.StatusBadRequest)
}

// SendNoOutstandingAmount sends no outstanding amount error
func SendNoOutstandingAmount(c *gin.Context, saleID uint) {
	NewResponseBuilder().
		Error().
		Message("No outstanding amount to pay").
		ErrorCode(ErrorCodeNoOutstandingAmount).
		ErrorMessage("This sale has no outstanding amount remaining").
		Context(map[string]interface{}{
			"sale_id": saleID,
		}).
		Send(c, http.StatusBadRequest)
}

// GetErrorFromCode returns HTTP status code for error code
func GetErrorFromCode(errorCode string) int {
	switch errorCode {
	case ErrorCodeValidation, ErrorCodeBadRequest, ErrorCodeBusinessRule,
		 ErrorCodeInvalidStatus, ErrorCodePaymentExceedsAmount, ErrorCodeNoOutstandingAmount:
		return http.StatusBadRequest
	case ErrorCodeNotFound, ErrorCodeSaleNotFound:
		return http.StatusNotFound
	case ErrorCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrorCodeForbidden:
		return http.StatusForbidden
	case ErrorCodeConflict:
		return http.StatusConflict
	case ErrorCodeMethodNotAllowed:
		return http.StatusMethodNotAllowed
	case ErrorCodeTooManyRequests:
		return http.StatusTooManyRequests
	case ErrorCodeServiceUnavailable:
		return http.StatusServiceUnavailable
	case ErrorCodeTimeout:
		return http.StatusGatewayTimeout
	case ErrorCodeInternalServer, ErrorCodeDatabaseError:
		fallthrough
	default:
		return http.StatusInternalServerError
	}
}

// Helper functions for parsing

// ParseIntWithDefault parses string to int with default value
func ParseIntWithDefault(str string, defaultValue int) int {
	if str == "" {
		return defaultValue
	}
	if val, err := strconv.Atoi(str); err == nil {
		return val
	}
	return defaultValue
}

// ParseDate parses string to time.Time
func ParseDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// Convenience response functions

// SuccessResponse creates a success response
func SuccessResponse(message string, data interface{}) gin.H {
	return gin.H{
		"status":  "success",
		"message": message,
		"data":    data,
	}
}

// CreateErrorResponse creates an error response (avoiding conflict with errors.go)
func CreateErrorResponse(message string, err error) gin.H {
	response := gin.H{
		"status":  "error",
		"message": message,
	}
	
	if err != nil {
		response["error"] = err.Error()
	}
	
	return response
}
