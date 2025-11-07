package models

import "time"

// Standard API Response Models for Swagger Documentation

// APIResponse represents the standard API response structure
type APIResponse struct {
	Status  string      `json:"status" example:"success"`
	Message string      `json:"message" example:"Operation successful"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse represents error response structure
type ErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Something went wrong"`
	Error   string `json:"error,omitempty" example:"detailed error message"`
}

// ValidationErrorResponse represents validation error response structure
type ValidationErrorResponse struct {
	Status  string            `json:"status" example:"error"`
	Message string            `json:"message" example:"Validation failed"`
	Errors  map[string]string `json:"errors"`
}


// LoginData represents login data in response
type LoginData struct {
	Token        string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User         User   `json:"user"`
}

// UserResponse represents user data in responses
type UserResponse struct {
	ID        uint       `json:"id" example:"1"`
	Name      string     `json:"name" example:"John Doe"` // For compatibility with tax_account_settings
	Username  string     `json:"username" example:"john_doe"`
	Email     string     `json:"email" example:"john@example.com"`
	FullName  string     `json:"full_name" example:"John Doe"`
	Role      string     `json:"role" example:"admin"`
	IsActive  bool       `json:"is_active" example:"true"`
	CreatedAt time.Time  `json:"created_at" example:"2024-01-15T09:30:00Z"`
	UpdatedAt time.Time  `json:"updated_at" example:"2024-01-15T09:30:00Z"`
}

// ProductResponse represents product data in responses
type ProductResponse struct {
	ID            uint    `json:"id" example:"1"`
	Name          string  `json:"name" example:"Laptop Dell XPS 13"`
	Code          string  `json:"code" example:"LP-DELL-001"`
	CategoryID    uint    `json:"category_id" example:"1"`
	UnitID        uint    `json:"unit_id" example:"1"`
	Description   string  `json:"description" example:"High performance laptop for business"`
	PurchasePrice float64 `json:"purchase_price" example:"15000000"`
	SellingPrice  float64 `json:"selling_price" example:"18000000"`
	Stock         int     `json:"stock" example:"25"`
	MinStock      int     `json:"min_stock" example:"5"`
	IsActive      bool    `json:"is_active" example:"true"`
	CreatedAt     time.Time `json:"created_at" example:"2024-01-15T09:30:00Z"`
	UpdatedAt     time.Time `json:"updated_at" example:"2024-01-15T09:30:00Z"`
}

// ContactResponse represents contact data in responses
type ContactResponse struct {
	ID           uint      `json:"id" example:"1"`
	Name         string    `json:"name" example:"PT. Supplier ABC"`
	Code         string    `json:"code" example:"SUP-001"`
	Type         string    `json:"type" example:"supplier"`
	Email        string    `json:"email" example:"contact@supplierabc.com"`
	Phone        string    `json:"phone" example:"021-1234567"`
	Address      string    `json:"address" example:"Jl. Sudirman No. 123, Jakarta"`
	TaxID        string    `json:"tax_id" example:"01.234.567.8-901.000"`
	IsActive     bool      `json:"is_active" example:"true"`
	CreatedAt    time.Time `json:"created_at" example:"2024-01-15T09:30:00Z"`
	UpdatedAt    time.Time `json:"updated_at" example:"2024-01-15T09:30:00Z"`
}

// SaleResponse represents sales transaction data in responses
type SaleResponse struct {
	ID               uint      `json:"id" example:"1"`
	Code             string    `json:"code" example:"SO-2024-001"`
	CustomerID       uint      `json:"customer_id" example:"1"`
	Date             time.Time `json:"date" example:"2024-01-15T09:30:00Z"`
	DueDate          time.Time `json:"due_date" example:"2024-01-30T09:30:00Z"`
	Subtotal         float64   `json:"subtotal" example:"17500000"`
	DiscountPercent  float64   `json:"discount_percent" example:"5"`
	DiscountAmount   float64   `json:"discount_amount" example:"875000"`
	TaxPercent       float64   `json:"tax_percent" example:"11"`
	TaxAmount        float64   `json:"tax_amount" example:"1828750"`
	Total            float64   `json:"total" example:"18453750"`
	Status           string    `json:"status" example:"confirmed"`
	Notes            string    `json:"notes" example:"Urgent delivery required"`
	CreatedBy        uint      `json:"created_by" example:"1"`
	CreatedAt        time.Time `json:"created_at" example:"2024-01-15T09:30:00Z"`
	UpdatedAt        time.Time `json:"updated_at" example:"2024-01-15T09:30:00Z"`
}

// PurchaseResponse represents purchase transaction data in responses
type PurchaseResponse struct {
	ID               uint      `json:"id" example:"1"`
	Code             string    `json:"code" example:"PO-2024-001"`
	SupplierID       uint      `json:"supplier_id" example:"1"`
	Date             time.Time `json:"date" example:"2024-01-15T09:30:00Z"`
	DueDate          time.Time `json:"due_date" example:"2024-01-30T09:30:00Z"`
	Subtotal         float64   `json:"subtotal" example:"15000000"`
	DiscountPercent  float64   `json:"discount_percent" example:"3"`
	DiscountAmount   float64   `json:"discount_amount" example:"450000"`
	TaxPercent       float64   `json:"tax_percent" example:"11"`
	TaxAmount        float64   `json:"tax_amount" example:"1600500"`
	Total            float64   `json:"total" example:"16150500"`
	Status           string    `json:"status" example:"approved"`
	Notes            string    `json:"notes" example:"Standard purchase order"`
	CreatedBy        uint      `json:"created_by" example:"1"`
	ApprovalStatus   string    `json:"approval_status" example:"approved"`
	CreatedAt        time.Time `json:"created_at" example:"2024-01-15T09:30:00Z"`
	UpdatedAt        time.Time `json:"updated_at" example:"2024-01-15T09:30:00Z"`
}

// AccountResponse represents chart of accounts data in responses
type AccountResponse struct {
	ID          uint      `json:"id" example:"1"`
	Code        string    `json:"code" example:"1100-001"`
	Name        string    `json:"name" example:"Kas Kecil"`
	Type        string    `json:"type" example:"asset"`
	Category    string    `json:"category" example:"current_asset"`
	ParentID    *uint     `json:"parent_id,omitempty" example:"1"`
	Level       int       `json:"level" example:"3"`
	IsHeader    bool      `json:"is_header" example:"false"`
	Balance     float64   `json:"balance" example:"5000000"`
	Description string    `json:"description" example:"Kas untuk keperluan operasional harian"`
	IsActive    bool      `json:"is_active" example:"true"`
	CreatedAt   time.Time `json:"created_at" example:"2024-01-15T09:30:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2024-01-15T09:30:00Z"`
}

// PaymentResponse represents payment data in responses  
type PaymentResponse struct {
	ID              uint      `json:"id" example:"1"`
	Code            string    `json:"code" example:"PMT-2024-001"`
	Type            string    `json:"type" example:"sales_payment"`
	ReferenceID     uint      `json:"reference_id" example:"1"`
	ContactID       uint      `json:"contact_id" example:"1"`
	CashBankID      uint      `json:"cash_bank_id" example:"1"`
	Amount          float64   `json:"amount" example:"5000000"`
	PaymentMethod   string    `json:"payment_method" example:"transfer"`
	PaymentDate     time.Time `json:"payment_date" example:"2024-01-15T09:30:00Z"`
	Notes           string    `json:"notes" example:"Pembayaran parsial"`
	CreatedBy       uint      `json:"created_by" example:"1"`
	CreatedAt       time.Time `json:"created_at" example:"2024-01-15T09:30:00Z"`
	UpdatedAt       time.Time `json:"updated_at" example:"2024-01-15T09:30:00Z"`
}

// DashboardResponse represents dashboard summary data
type DashboardResponse struct {
	TotalSales      float64 `json:"total_sales" example:"125000000"`
	TotalPurchases  float64 `json:"total_purchases" example:"80000000"`
	TotalPayments   float64 `json:"total_payments" example:"95000000"`
	LowStockCount   int     `json:"low_stock_count" example:"5"`
	PendingApprovals int    `json:"pending_approvals" example:"3"`
}

// PaginatedResponse represents paginated data response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination struct {
		CurrentPage int   `json:"current_page" example:"1"`
		TotalPages  int   `json:"total_pages" example:"10"`
		TotalItems  int64 `json:"total_items" example:"95"`
		PerPage     int   `json:"per_page" example:"10"`
	} `json:"pagination"`
}