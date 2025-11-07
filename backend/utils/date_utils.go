package utils

import (
	"fmt"
	"strings"
	"time"
)

// Jakarta timezone for Indonesia
var JakartaTZ *time.Location

func init() {
	var err error
	JakartaTZ, err = time.LoadLocation("Asia/Jakarta")
	if err != nil {
		// Fallback to UTC+7 if timezone loading fails
		JakartaTZ = time.FixedZone("WIB", 7*3600)
	}
}

// DateUtils provides utilities for date handling in the application
type DateUtils struct{}

// NewDateUtils creates a new DateUtils instance
func NewDateUtils() *DateUtils {
	return &DateUtils{}
}

// ParseDateTimeWithTZ parses date string in YYYY-MM-DD format with Jakarta timezone
func (du *DateUtils) ParseDateTimeWithTZ(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("date string cannot be empty")
	}

	// Parse as start of day in Jakarta timezone
	t, err := time.ParseInLocation("2006-01-02", dateStr, JakartaTZ)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse date '%s': %v", dateStr, err)
	}
	return t, nil
}

// ParseEndDateTimeWithTZ parses date string as end of day in Jakarta timezone
func (du *DateUtils) ParseEndDateTimeWithTZ(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("date string cannot be empty")
	}

	// Parse as start of day first
	startOfDay, err := time.ParseInLocation("2006-01-02", dateStr, JakartaTZ)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse date '%s': %v", dateStr, err)
	}

	// Convert to end of day (23:59:59.999)
	endOfDay := startOfDay.Add(24*time.Hour - time.Nanosecond)
	return endOfDay, nil
}

// FormatDateRange formats date range for display with timezone awareness
func (du *DateUtils) FormatDateRange(start, end time.Time) string {
	startFormatted := start.In(JakartaTZ).Format("2006-01-02 15:04:05 MST")
	endFormatted := end.In(JakartaTZ).Format("2006-01-02 15:04:05 MST")
	return fmt.Sprintf("%s to %s", startFormatted, endFormatted)
}

// GetCurrentJakartaTime returns current time in Jakarta timezone
func (du *DateUtils) GetCurrentJakartaTime() time.Time {
	return time.Now().In(JakartaTZ)
}

// ValidateDateRange validates that end date is after start date
func (du *DateUtils) ValidateDateRange(start, end time.Time) error {
	if end.Before(start) {
		return fmt.Errorf("end date (%s) cannot be before start date (%s)", 
			end.Format("2006-01-02"), start.Format("2006-01-02"))
	}

	// Check for reasonable date range (not more than 5 years)
	maxDuration := 5 * 365 * 24 * time.Hour
	if end.Sub(start) > maxDuration {
		return fmt.Errorf("date range too large (max 5 years): %s to %s",
			start.Format("2006-01-02"), end.Format("2006-01-02"))
	}

	return nil
}

// GetPeriodBounds returns start and end bounds for a given period
func (du *DateUtils) GetPeriodBounds(date time.Time, groupBy string) (time.Time, time.Time) {
	jakartaDate := date.In(JakartaTZ)
	
	switch groupBy {
	case "day":
		start := time.Date(jakartaDate.Year(), jakartaDate.Month(), jakartaDate.Day(), 0, 0, 0, 0, JakartaTZ)
		end := start.Add(24*time.Hour - time.Nanosecond)
		return start, end

	case "week":
		// Start from Monday
		weekday := int(jakartaDate.Weekday())
		if weekday == 0 { // Sunday
			weekday = 7
		}
		start := time.Date(jakartaDate.Year(), jakartaDate.Month(), jakartaDate.Day(), 0, 0, 0, 0, JakartaTZ).AddDate(0, 0, -(weekday-1))
		end := start.AddDate(0, 0, 7).Add(-time.Nanosecond)
		return start, end

	case "month":
		start := time.Date(jakartaDate.Year(), jakartaDate.Month(), 1, 0, 0, 0, 0, JakartaTZ)
		end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
		return start, end

	case "quarter":
		month := jakartaDate.Month()
		quarterStart := ((int(month) - 1) / 3) * 3 + 1
		start := time.Date(jakartaDate.Year(), time.Month(quarterStart), 1, 0, 0, 0, 0, JakartaTZ)
		end := start.AddDate(0, 3, 0).Add(-time.Nanosecond)
		return start, end

	case "year":
		start := time.Date(jakartaDate.Year(), 1, 1, 0, 0, 0, 0, JakartaTZ)
		end := start.AddDate(1, 0, 0).Add(-time.Nanosecond)
		return start, end

	default:
		// Default to day
		start := time.Date(jakartaDate.Year(), jakartaDate.Month(), jakartaDate.Day(), 0, 0, 0, 0, JakartaTZ)
		end := start.Add(24*time.Hour - time.Nanosecond)
		return start, end
	}
}

// FormatPeriodWithTZ formats date according to groupBy parameter with timezone awareness
func (du *DateUtils) FormatPeriodWithTZ(date time.Time, groupBy string) string {
	jakartaDate := date.In(JakartaTZ)

	switch groupBy {
	case "day":
		return jakartaDate.Format("2006-01-02")
	case "week":
		// Get start of week (Monday)
		weekday := int(jakartaDate.Weekday())
		if weekday == 0 { // Sunday
			weekday = 7
		}
		startOfWeek := jakartaDate.AddDate(0, 0, -(weekday-1))
		return fmt.Sprintf("Week of %s", startOfWeek.Format("2006-01-02"))
	case "month":
		return jakartaDate.Format("2006-01")
	case "quarter":
		quarter := ((int(jakartaDate.Month()) - 1) / 3) + 1
		return fmt.Sprintf("%d-Q%d", jakartaDate.Year(), quarter)
	case "year":
		return jakartaDate.Format("2006")
	default:
		return jakartaDate.Format("2006-01-02")
	}
}

// CalculateDueDateFromPaymentTerms calculates due date based on payment terms
func (du *DateUtils) CalculateDueDateFromPaymentTerms(invoiceDate time.Time, paymentTerms string) time.Time {
	// Normalize payment terms to uppercase
	normalizedTerms := strings.ToUpper(strings.TrimSpace(paymentTerms))
	
	switch normalizedTerms {
	case "COD", "CASH_ON_DELIVERY":
		// Cash on Delivery - same day payment
		return invoiceDate
	case "EOM", "END_OF_MONTH":
		// End of Month - due on last day of current month
		year, month, _ := invoiceDate.Date()
		lastDayOfMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, invoiceDate.Location())
		return lastDayOfMonth
	case "2_10_NET_30", "2/10_NET_30":
		// 2/10, Net 30 - 2% discount if paid within 10 days, otherwise net 30 days
		return invoiceDate.AddDate(0, 0, 30)
	case "NET7", "NET_7":
		return invoiceDate.AddDate(0, 0, 7)
	case "NET15", "NET_15":
		return invoiceDate.AddDate(0, 0, 15)
	case "NET30", "NET_30":
		return invoiceDate.AddDate(0, 0, 30)
	case "NET45", "NET_45":
		return invoiceDate.AddDate(0, 0, 45)
	case "NET60", "NET_60":
		return invoiceDate.AddDate(0, 0, 60)
	case "NET90", "NET_90":
		return invoiceDate.AddDate(0, 0, 90)
	case "NET120", "NET_120":
		return invoiceDate.AddDate(0, 0, 120)
	default:
		// Default to NET30 if unknown term
		return invoiceDate.AddDate(0, 0, 30)
	}
}

// GetPaymentTermsDescription returns human-readable description of payment terms
func (du *DateUtils) GetPaymentTermsDescription(paymentTerms string, invoiceDate time.Time) string {
	normalizedTerms := strings.ToUpper(strings.TrimSpace(paymentTerms))
	dueDate := du.CalculateDueDateFromPaymentTerms(invoiceDate, paymentTerms)
	
	// Use Indonesian month names for clarity
	invoiceDateStr := du.FormatDateWithIndonesianMonth(invoiceDate)
	dueDateStr := du.FormatDateWithIndonesianMonth(dueDate)
	
	switch normalizedTerms {
	case "COD", "CASH_ON_DELIVERY":
		return fmt.Sprintf("Pembayaran tunai pada saat pengiriman (%s)", invoiceDateStr)
	case "EOM", "END_OF_MONTH":
		return fmt.Sprintf("Pembayaran pada akhir bulan (%s → %s)", invoiceDateStr, dueDateStr)
	case "2_10_NET_30", "2/10_NET_30":
		return fmt.Sprintf("2%% diskon jika dibayar dalam 10 hari, atau penuh dalam 30 hari (%s → %s)", invoiceDateStr, dueDateStr)
	case "NET7", "NET_7":
		return fmt.Sprintf("Pembayaran dalam 7 hari (%s → %s)", invoiceDateStr, dueDateStr)
	case "NET15", "NET_15":
		return fmt.Sprintf("Pembayaran dalam 15 hari (%s → %s)", invoiceDateStr, dueDateStr)
	case "NET30", "NET_30":
		return fmt.Sprintf("Pembayaran dalam 30 hari (%s → %s)", invoiceDateStr, dueDateStr)
	case "NET45", "NET_45":
		return fmt.Sprintf("Pembayaran dalam 45 hari (%s → %s)", invoiceDateStr, dueDateStr)
	case "NET60", "NET_60":
		return fmt.Sprintf("Pembayaran dalam 60 hari (%s → %s)", invoiceDateStr, dueDateStr)
	case "NET90", "NET_90":
		return fmt.Sprintf("Pembayaran dalam 90 hari (%s → %s)", invoiceDateStr, dueDateStr)
	case "NET120", "NET_120":
		return fmt.Sprintf("Pembayaran dalam 120 hari (%s → %s)", invoiceDateStr, dueDateStr)
	default:
		return fmt.Sprintf("Pembayaran dalam 30 hari (default) (%s → %s)", invoiceDateStr, dueDateStr)
	}
}

// ValidatePaymentTerms validates if the payment terms are valid
func (du *DateUtils) ValidatePaymentTerms(paymentTerms string) error {
	normalizedTerms := strings.ToUpper(strings.TrimSpace(paymentTerms))
	
	validTerms := map[string]bool{
		"COD":            true,
		"CASH_ON_DELIVERY": true,
		"EOM":            true,
		"END_OF_MONTH":   true,
		"2_10_NET_30":    true,
		"2/10_NET_30":    true,
		"NET7":           true,
		"NET_7":          true,
		"NET15":          true,
		"NET_15":         true,
		"NET30":          true,
		"NET_30":         true,
		"NET45":          true,
		"NET_45":         true,
		"NET60":          true,
		"NET_60":         true,
		"NET90":          true,
		"NET_90":         true,
		"NET120":         true,
		"NET_120":        true,
	}
	
	if !validTerms[normalizedTerms] {
		return fmt.Errorf("invalid payment terms: %s. Valid terms: COD, NET15, NET30, NET45, NET60, NET90, EOM", paymentTerms)
	}
	
	return nil
}

// FormatDateForIndonesia formats date in Indonesian format (DD/MM/YYYY)
func (du *DateUtils) FormatDateForIndonesia(date time.Time) string {
	return date.Format("02/01/2006")
}

// FormatDateWithIndonesianMonth formats date with Indonesian month names (DD Month YYYY)
func (du *DateUtils) FormatDateWithIndonesianMonth(date time.Time) string {
	monthNames := map[int]string{
		1:  "Januari",
		2:  "Februari",
		3:  "Maret",
		4:  "April",
		5:  "Mei",
		6:  "Juni",
		7:  "Juli",
		8:  "Agustus",
		9:  "September",
		10: "Oktober",
		11: "November",
		12: "Desember",
	}
	
	day := date.Day()
	month := int(date.Month())
	year := date.Year()
	
	return fmt.Sprintf("%d %s %d", day, monthNames[month], year)
}

// FormatDateForAPI formats date for API (YYYY-MM-DD)
func (du *DateUtils) FormatDateForAPI(date time.Time) string {
	return date.Format("2006-01-02")
}

// ParseIndonesianDate parses Indonesian date format (DD/MM/YYYY) to time.Time
func (du *DateUtils) ParseIndonesianDate(dateStr string) (time.Time, error) {
	return time.Parse("02/01/2006", dateStr)
}

// IsBusinessDay checks if the given date is a business day (Monday-Friday)
func (du *DateUtils) IsBusinessDay(date time.Time) bool {
	weekday := date.Weekday()
	return weekday != time.Saturday && weekday != time.Sunday
}

// AdjustToBusinessDay adjusts the due date to the next business day if it falls on weekend
func (du *DateUtils) AdjustToBusinessDay(date time.Time) time.Time {
	weekday := date.Weekday()
	switch weekday {
	case time.Saturday:
		// Move to Monday
		return date.AddDate(0, 0, 2)
	case time.Sunday:
		// Move to Monday  
		return date.AddDate(0, 0, 1)
	default:
		// It's a weekday, return as is
		return date
	}
}

// CalculateDaysOverdue calculates how many days an invoice is overdue
func (du *DateUtils) CalculateDaysOverdue(dueDate time.Time) int {
	now := time.Now().In(JakartaTZ)
	if now.Before(dueDate) {
		return 0 // Not overdue yet
	}
	
	duration := now.Sub(dueDate)
	return int(duration.Hours() / 24)
}

// GetPaymentTermsOptions returns available payment terms for UI
func (du *DateUtils) GetPaymentTermsOptions() []map[string]string {
	return []map[string]string{
		{"value": "COD", "label": "COD - Cash on Delivery"},
		{"value": "NET7", "label": "NET 7 - 7 Hari"},
		{"value": "NET15", "label": "NET 15 - 15 Hari"},
		{"value": "NET30", "label": "NET 30 - 30 Hari"},
		{"value": "NET45", "label": "NET 45 - 45 Hari"},
		{"value": "NET60", "label": "NET 60 - 60 Hari"},
		{"value": "NET90", "label": "NET 90 - 90 Hari"},
		{"value": "EOM", "label": "EOM - End of Month"},
	}
}
