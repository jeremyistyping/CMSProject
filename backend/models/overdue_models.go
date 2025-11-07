package models

import (
	"time"
	"gorm.io/gorm"
)

// ReminderLog tracks when reminders were sent
type ReminderLog struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	SaleID      uint      `json:"sale_id" gorm:"not null;index"`
	DaysBefore  int       `json:"days_before" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at"`
	
	// Relations
	Sale        *Sale     `json:"sale,omitempty" gorm:"foreignKey:SaleID"`
}

// OverdueRecord tracks overdue invoices
type OverdueRecord struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	SaleID            uint      `json:"sale_id" gorm:"not null;index"`
	OverdueDate       time.Time `json:"overdue_date" gorm:"not null"`
	DaysOverdue       int       `json:"days_overdue" gorm:"not null"`
	OutstandingAmount float64   `json:"outstanding_amount" gorm:"type:decimal(15,2);not null"`
	Status            string    `json:"status" gorm:"type:varchar(50);not null;default:'ACTIVE'"` // ACTIVE, RESOLVED, WRITTEN_OFF
	ResolvedAt        *time.Time `json:"resolved_at,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	
	// Relations
	Sale              *Sale     `json:"sale,omitempty" gorm:"foreignKey:SaleID"`
}

// InterestCharge tracks interest charges on overdue invoices
type InterestCharge struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	SaleID       uint      `json:"sale_id" gorm:"not null;index"`
	ChargeDate   time.Time `json:"charge_date" gorm:"not null"`
	DaysOverdue  int       `json:"days_overdue" gorm:"not null"`
	Principal    float64   `json:"principal" gorm:"type:decimal(15,2);not null"`
	InterestRate float64   `json:"interest_rate" gorm:"type:decimal(5,2);not null"` // Annual percentage
	Amount       float64   `json:"amount" gorm:"type:decimal(15,2);not null"`
	Status       string    `json:"status" gorm:"type:varchar(50);not null;default:'APPLIED'"` // APPLIED, REVERSED, WAIVED
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	
	// Relations
	Sale         *Sale     `json:"sale,omitempty" gorm:"foreignKey:SaleID"`
}

// CollectionTask tracks collection efforts
type CollectionTask struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	SaleID       uint       `json:"sale_id" gorm:"not null;index"`
	Level        string     `json:"level" gorm:"type:varchar(50);not null"` // GENTLE, MODERATE, ACTIVE, INTENSIVE, AGGRESSIVE, LEGAL
	Action       string     `json:"action" gorm:"type:text;not null"`
	DaysOverdue  int        `json:"days_overdue" gorm:"not null"`
	Priority     string     `json:"priority" gorm:"type:varchar(20);not null"` // LOW, MEDIUM, HIGH, CRITICAL
	Status       string     `json:"status" gorm:"type:varchar(50);not null;default:'PENDING'"` // PENDING, IN_PROGRESS, COMPLETED, FAILED
	AssignedTo   *uint      `json:"assigned_to,omitempty" gorm:"index"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	Notes        string     `json:"notes" gorm:"type:text"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	
	// Relations
	Sale         *Sale      `json:"sale,omitempty" gorm:"foreignKey:SaleID"`
	AssignedUser *User      `json:"assigned_user,omitempty" gorm:"foreignKey:AssignedTo"`
}

// WriteOffSuggestion suggests invoices for write-off
type WriteOffSuggestion struct {
	ID                uint       `json:"id" gorm:"primaryKey"`
	SaleID            uint       `json:"sale_id" gorm:"not null;index"`
	OutstandingAmount float64    `json:"outstanding_amount" gorm:"type:decimal(15,2);not null"`
	DaysOverdue       int        `json:"days_overdue" gorm:"not null"`
	Reason            string     `json:"reason" gorm:"type:text;not null"`
	Status            string     `json:"status" gorm:"type:varchar(50);not null;default:'PENDING_APPROVAL'"` // PENDING_APPROVAL, APPROVED, REJECTED, WRITTEN_OFF
	Priority          string     `json:"priority" gorm:"type:varchar(20);not null"` // LOW, MEDIUM, HIGH
	ApprovedBy        *uint      `json:"approved_by,omitempty" gorm:"index"`
	ApprovedAt        *time.Time `json:"approved_at,omitempty"`
	WrittenOffAt      *time.Time `json:"written_off_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	
	// Relations
	Sale              *Sale      `json:"sale,omitempty" gorm:"foreignKey:SaleID"`
	Approver          *User      `json:"approver,omitempty" gorm:"foreignKey:ApprovedBy"`
}

// SaleCancellation tracks sale cancellations
type SaleCancellation struct {
	ID               uint       `json:"id" gorm:"primaryKey"`
	SaleID           uint       `json:"sale_id" gorm:"not null;index"`
	CancellationDate time.Time  `json:"cancellation_date" gorm:"not null"`
	Reason           string     `json:"reason" gorm:"type:text;not null"`
	UserID           uint       `json:"user_id" gorm:"not null;index"`
	OriginalAmount   float64    `json:"original_amount" gorm:"type:decimal(15,2);not null"`
	RefundAmount     float64    `json:"refund_amount" gorm:"type:decimal(15,2);default:0"`
	Status           string     `json:"status" gorm:"type:varchar(50);not null;default:'PENDING'"` // PENDING, APPROVED, COMPLETED
	ApprovedBy       *uint      `json:"approved_by,omitempty" gorm:"index"`
	ApprovedAt       *time.Time `json:"approved_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	
	// Relations
	Sale             *Sale      `json:"sale,omitempty" gorm:"foreignKey:SaleID"`
	User             *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Approver         *User      `json:"approver,omitempty" gorm:"foreignKey:ApprovedBy"`
}

// CreditNote represents credit notes issued for cancellations/returns
type CreditNote struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	SaleID           uint      `json:"sale_id" gorm:"not null;index"`
	CreditNoteNumber string    `json:"credit_note_number" gorm:"type:varchar(100);not null;unique"`
	Amount           float64   `json:"amount" gorm:"type:decimal(15,2);not null"`
	Reason           string    `json:"reason" gorm:"type:text;not null"`
	Date             time.Time `json:"date" gorm:"not null"`
	Status           string    `json:"status" gorm:"type:varchar(50);not null;default:'ISSUED'"` // ISSUED, APPLIED, CANCELLED
	AppliedToSaleID  *uint     `json:"applied_to_sale_id,omitempty" gorm:"index"` // If credit note applied to another sale
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	
	// Relations
	Sale             *Sale     `json:"sale,omitempty" gorm:"foreignKey:SaleID"`
	AppliedToSale    *Sale     `json:"applied_to_sale,omitempty" gorm:"foreignKey:AppliedToSaleID"`
}

// PaymentReminder tracks automated payment reminders
type PaymentReminder struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	SaleID      uint      `json:"sale_id" gorm:"not null;index"`
	ReminderType string   `json:"reminder_type" gorm:"type:varchar(50);not null"` // EMAIL, SMS, PHONE, LETTER
	DaysBefore  int       `json:"days_before" gorm:"not null"`
	SentAt      time.Time `json:"sent_at" gorm:"not null"`
	Status      string    `json:"status" gorm:"type:varchar(50);not null"` // SENT, DELIVERED, FAILED, READ
	CreatedAt   time.Time `json:"created_at"`
	
	// Relations
	Sale        *Sale     `json:"sale,omitempty" gorm:"foreignKey:SaleID"`
}

// OverdueAnalytics provides analytics for overdue management
type OverdueAnalytics struct {
	TotalOverdueAmount    float64 `json:"total_overdue_amount"`
	TotalOverdueCount     int64   `json:"total_overdue_count"`
	AverageOverdueDays    float64 `json:"average_overdue_days"`
	WorstOverdueDays      int     `json:"worst_overdue_days"`
	TotalInterestCharged  float64 `json:"total_interest_charged"`
	CollectionEfficiency  float64 `json:"collection_efficiency"` // Percentage of overdue collected
	WriteOffRate          float64 `json:"write_off_rate"` // Percentage of total sales written off
}

// OverdueAging represents aging analysis of overdue amounts
type OverdueAging struct {
	Period       string  `json:"period"`        // "0-30", "31-60", "61-90", "91-120", "120+"
	Count        int64   `json:"count"`
	Amount       float64 `json:"amount"`
	Percentage   float64 `json:"percentage"`
}

// Hooks for automatic timestamp updates
func (r *ReminderLog) BeforeCreate(tx *gorm.DB) error {
	r.CreatedAt = time.Now()
	return nil
}

func (o *OverdueRecord) BeforeCreate(tx *gorm.DB) error {
	o.CreatedAt = time.Now()
	o.UpdatedAt = time.Now()
	return nil
}

func (o *OverdueRecord) BeforeUpdate(tx *gorm.DB) error {
	o.UpdatedAt = time.Now()
	return nil
}

func (i *InterestCharge) BeforeCreate(tx *gorm.DB) error {
	i.CreatedAt = time.Now()
	i.UpdatedAt = time.Now()
	return nil
}

func (i *InterestCharge) BeforeUpdate(tx *gorm.DB) error {
	i.UpdatedAt = time.Now()
	return nil
}

func (c *CollectionTask) BeforeCreate(tx *gorm.DB) error {
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	return nil
}

func (c *CollectionTask) BeforeUpdate(tx *gorm.DB) error {
	c.UpdatedAt = time.Now()
	return nil
}

func (w *WriteOffSuggestion) BeforeCreate(tx *gorm.DB) error {
	w.CreatedAt = time.Now()
	w.UpdatedAt = time.Now()
	return nil
}

func (w *WriteOffSuggestion) BeforeUpdate(tx *gorm.DB) error {
	w.UpdatedAt = time.Now()
	return nil
}

func (s *SaleCancellation) BeforeCreate(tx *gorm.DB) error {
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()
	return nil
}

func (s *SaleCancellation) BeforeUpdate(tx *gorm.DB) error {
	s.UpdatedAt = time.Now()
	return nil
}

func (c *CreditNote) BeforeCreate(tx *gorm.DB) error {
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	return nil
}

func (c *CreditNote) BeforeUpdate(tx *gorm.DB) error {
	c.UpdatedAt = time.Now()
	return nil
}

func (p *PaymentReminder) BeforeCreate(tx *gorm.DB) error {
	p.CreatedAt = time.Now()
	return nil
}
