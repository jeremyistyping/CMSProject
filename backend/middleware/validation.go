package middleware

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ValidationManager struct {
	validator *validator.Validate
}

func NewValidationManager() *ValidationManager {
	validate := validator.New()
	
	// Register custom validation functions
	validate.RegisterValidation("password", validatePassword)
	validate.RegisterValidation("username", validateUsername)
	validate.RegisterValidation("phone_id", validateIndonesianPhone)
	validate.RegisterValidation("currency", validateCurrency)
	
	// Use JSON tag name for field names in error messages
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &ValidationManager{
		validator: validate,
	}
}

// ValidateStruct validates a struct and returns formatted error messages
func (vm *ValidationManager) ValidateStruct(s interface{}) map[string]string {
	errs := vm.validator.Struct(s)
	if errs == nil {
		return nil
	}

	validationErrors := make(map[string]string)
	
	for _, err := range errs.(validator.ValidationErrors) {
		field := err.Field()
		tag := err.Tag()
		
		switch tag {
		case "required":
			validationErrors[field] = field + " is required"
		case "email":
			validationErrors[field] = field + " must be a valid email address"
		case "min":
			validationErrors[field] = field + " must be at least " + err.Param() + " characters long"
		case "max":
			validationErrors[field] = field + " must be at most " + err.Param() + " characters long"
		case "eqfield":
			validationErrors[field] = field + " must match " + err.Param()
		case "oneof":
			validationErrors[field] = field + " must be one of: " + err.Param()
		case "alphanum":
			validationErrors[field] = field + " must contain only alphanumeric characters"
		case "e164":
			validationErrors[field] = field + " must be a valid phone number"
		case "password":
			validationErrors[field] = "Password must contain at least 8 characters, including uppercase, lowercase, number, and special character"
		case "username":
			validationErrors[field] = "Username must be 3-50 characters long and contain only letters, numbers, and underscores"
		case "phone_id":
			validationErrors[field] = "Phone number must be a valid Indonesian phone number"
		case "currency":
			validationErrors[field] = "Currency must be a valid 3-letter currency code"
		default:
			validationErrors[field] = field + " is invalid"
		}
	}

	return validationErrors
}

// ValidationMiddleware provides validation for request bodies
func (vm *ValidationManager) ValidationMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// This middleware will be used in conjunction with ShouldBindJSON
		// The actual validation happens in the controller when binding
		c.Next()
	})
}

// ValidateJSON validates JSON request body against a struct
func (vm *ValidationManager) ValidateJSON(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON format",
			"code":  "INVALID_JSON",
			"details": err.Error(),
		})
		return false
	}

	if validationErrors := vm.ValidateStruct(obj); validationErrors != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Validation failed",
			"code":  "VALIDATION_ERROR",
			"details": validationErrors,
		})
		return false
	}

	return true
}

// Custom validation functions

// validatePassword checks password strength
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	
	if len(password) < 8 {
		return false
	}

	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}

// validateUsername checks username format
func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	
	if len(username) < 3 || len(username) > 50 {
		return false
	}

	for _, char := range username {
		if !((char >= 'a' && char <= 'z') || 
			 (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || 
			 char == '_') {
			return false
		}
	}

	return true
}

// validateIndonesianPhone checks Indonesian phone number format
func validateIndonesianPhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	
	// Remove any non-digit characters
	digits := ""
	for _, char := range phone {
		if char >= '0' && char <= '9' {
			digits += string(char)
		}
	}

	// Check Indonesian phone number patterns
	// Mobile: 08xx-xxxx-xxxx (10-13 digits starting with 08)
	// Landline: 021-xxxx-xxxx, 022-xxxx-xxxx, etc.
	
	if len(digits) >= 10 && len(digits) <= 15 {
		// Mobile numbers starting with 08
		if strings.HasPrefix(digits, "08") {
			return len(digits) >= 10 && len(digits) <= 13
		}
		// Landline numbers starting with 0
		if strings.HasPrefix(digits, "0") && len(digits) >= 10 {
			return true
		}
		// International format starting with 62
		if strings.HasPrefix(digits, "62") && len(digits) >= 11 {
			return true
		}
	}

	return false
}

// validateCurrency checks if currency code is valid
func validateCurrency(fl validator.FieldLevel) bool {
	currency := strings.ToUpper(fl.Field().String())
	
	validCurrencies := map[string]bool{
		"IDR": true, "USD": true, "EUR": true, "SGD": true,
		"MYR": true, "JPY": true, "GBP": true, "AUD": true,
		"CNY": true, "KRW": true, "THB": true, "VND": true,
	}

	return validCurrencies[currency]
}

// SanitizeInput removes potentially dangerous characters from input
func SanitizeInput(input string) string {
	// Remove HTML tags
	input = strings.ReplaceAll(input, "<", "&lt;")
	input = strings.ReplaceAll(input, ">", "&gt;")
	
	// Remove script tags
	input = strings.ReplaceAll(input, "<script", "&lt;script")
	input = strings.ReplaceAll(input, "</script>", "&lt;/script&gt;")
	
	// Remove javascript: protocol
	input = strings.ReplaceAll(input, "javascript:", "")
	
	// Remove SQL injection patterns
	sqlPatterns := []string{
		"'", "\"", ";", "--", "/*", "*/", "xp_", "sp_",
		"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "CREATE",
		"ALTER", "EXEC", "EXECUTE", "UNION", "SCRIPT",
	}
	
	for _, pattern := range sqlPatterns {
		input = strings.ReplaceAll(strings.ToLower(input), strings.ToLower(pattern), "")
	}

	return strings.TrimSpace(input)
}

// Input sanitization middleware
func (vm *ValidationManager) SanitizeInputMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize URL parameters
		for key, values := range c.Request.URL.Query() {
			for i, value := range values {
				c.Request.URL.Query()[key][i] = SanitizeInput(value)
			}
		}

		// Sanitize form data
		if c.Request.Form != nil {
			for key, values := range c.Request.Form {
				for i, value := range values {
					c.Request.Form[key][i] = SanitizeInput(value)
				}
			}
		}

		c.Next()
	}
}

// Password policy validation
type PasswordPolicy struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireNumber  bool
	RequireSpecial bool
	MaxLength      int
}

var DefaultPasswordPolicy = PasswordPolicy{
	MinLength:      8,
	RequireUpper:   true,
	RequireLower:   true,
	RequireNumber:  true,
	RequireSpecial: true,
	MaxLength:      128,
}

// ValidatePasswordPolicy validates password against policy
func ValidatePasswordPolicy(password string, policy PasswordPolicy) []string {
	var errors []string

	if len(password) < policy.MinLength {
		errors = append(errors, "Password must be at least " + string(rune(policy.MinLength)) + " characters long")
	}

	if len(password) > policy.MaxLength {
		errors = append(errors, "Password must be at most " + string(rune(policy.MaxLength)) + " characters long")
	}

	if policy.RequireUpper {
		hasUpper := false
		for _, char := range password {
			if char >= 'A' && char <= 'Z' {
				hasUpper = true
				break
			}
		}
		if !hasUpper {
			errors = append(errors, "Password must contain at least one uppercase letter")
		}
	}

	if policy.RequireLower {
		hasLower := false
		for _, char := range password {
			if char >= 'a' && char <= 'z' {
				hasLower = true
				break
			}
		}
		if !hasLower {
			errors = append(errors, "Password must contain at least one lowercase letter")
		}
	}

	if policy.RequireNumber {
		hasNumber := false
		for _, char := range password {
			if char >= '0' && char <= '9' {
				hasNumber = true
				break
			}
		}
		if !hasNumber {
			errors = append(errors, "Password must contain at least one number")
		}
	}

	if policy.RequireSpecial {
		hasSpecial := false
		specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
		for _, char := range password {
			if strings.ContainsRune(specialChars, char) {
				hasSpecial = true
				break
			}
		}
		if !hasSpecial {
			errors = append(errors, "Password must contain at least one special character")
		}
	}

	return errors
}

// Common validation rules for different entity types
func ValidateUserRegistration(user interface{}) map[string]string {
	vm := NewValidationManager()
	return vm.ValidateStruct(user)
}

func ValidateLoginRequest(req interface{}) map[string]string {
	vm := NewValidationManager()
	return vm.ValidateStruct(req)
}

func ValidateAccountData(account interface{}) map[string]string {
	vm := NewValidationManager()
	return vm.ValidateStruct(account)
}

func ValidateProductData(product interface{}) map[string]string {
	vm := NewValidationManager()
	return vm.ValidateStruct(product)
}

func ValidateTransactionData(transaction interface{}) map[string]string {
	vm := NewValidationManager()
	return vm.ValidateStruct(transaction)
}
