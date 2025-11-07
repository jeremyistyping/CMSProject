package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// FormatRupiah formats a float64 value to Indonesian Rupiah format
// Example: 1234567.89 -> "Rp 1.234.567,89"
func FormatRupiah(amount float64) string {
	// Handle negative values
	isNegative := amount < 0
	if isNegative {
		amount = -amount
	}

	// Format with 2 decimal places
	amountStr := fmt.Sprintf("%.2f", amount)
	parts := strings.Split(amountStr, ".")
	
	integerPart := parts[0]
	decimalPart := parts[1]
	
	// Add thousand separators (dots) to integer part
	formattedInteger := addThousandSeparators(integerPart)
	
	// Construct final format with comma as decimal separator
	result := fmt.Sprintf("Rp %s,%s", formattedInteger, decimalPart)
	
	if isNegative {
		result = "-" + result
	}
	
	return result
}

// FormatRupiahWithoutDecimals formats currency without decimal places if amount is whole number
// Example: 1234567.00 -> "Rp 1.234.567", 1234567.50 -> "Rp 1.234.567,50"
func FormatRupiahWithoutDecimals(amount float64) string {
	// Check if it's a whole number
	if amount == float64(int64(amount)) {
		return FormatRupiahInteger(int64(amount))
	}
	return FormatRupiah(amount)
}

// FormatRupiahInteger formats an integer to Indonesian Rupiah format
// Example: 1234567 -> "Rp 1.234.567"
func FormatRupiahInteger(amount int64) string {
	// Handle negative values
	isNegative := amount < 0
	if isNegative {
		amount = -amount
	}
	
	amountStr := strconv.FormatInt(amount, 10)
	formattedAmount := addThousandSeparators(amountStr)
	
	result := fmt.Sprintf("Rp %s", formattedAmount)
	
	if isNegative {
		result = "-" + result
	}
	
	return result
}

// addThousandSeparators adds dots as thousand separators
// Example: "1234567" -> "1.234.567"
func addThousandSeparators(numberStr string) string {
	// Reverse the string for easier processing
	runes := []rune(numberStr)
	n := len(runes)
	
	var result []rune
	for i := 0; i < n; i++ {
		if i > 0 && i%3 == 0 {
			result = append(result, '.')
		}
		result = append(result, runes[n-1-i])
	}
	
	// Reverse back to get correct order
	for i := 0; i < len(result)/2; i++ {
		result[i], result[len(result)-1-i] = result[len(result)-1-i], result[i]
	}
	
	return string(result)
}

// ParseRupiah parses Indonesian Rupiah format to float64
// Example: "Rp 1.234.567,89" -> 1234567.89
func ParseRupiah(currencyStr string) (float64, error) {
	// Remove "Rp " prefix and spaces
	cleanStr := strings.TrimSpace(currencyStr)
	cleanStr = strings.TrimPrefix(cleanStr, "Rp")
	cleanStr = strings.TrimSpace(cleanStr)
	
	// Handle negative values
	isNegative := strings.HasPrefix(cleanStr, "-")
	if isNegative {
		cleanStr = strings.TrimPrefix(cleanStr, "-")
		cleanStr = strings.TrimSpace(cleanStr)
	}
	
	// Replace Indonesian format with standard format
	// Replace dots (thousand separators) with empty string
	// Replace comma (decimal separator) with dot
	cleanStr = strings.ReplaceAll(cleanStr, ".", "")
	cleanStr = strings.ReplaceAll(cleanStr, ",", ".")
	
	amount, err := strconv.ParseFloat(cleanStr, 64)
	if err != nil {
		return 0, err
	}
	
	if isNegative {
		amount = -amount
	}
	
	return amount, nil
}

// FormatRupiahCompact formats currency in compact form for notifications
// Example: 1234567.89 -> "Rp 1,23 juta"
func FormatRupiahCompact(amount float64) string {
	abs := amount
	if abs < 0 {
		abs = -abs
	}
	
	var formatted string
	if abs >= 1000000000 { // Miliar
		formatted = fmt.Sprintf("Rp %.1f miliar", amount/1000000000)
	} else if abs >= 1000000 { // Juta
		formatted = fmt.Sprintf("Rp %.1f juta", amount/1000000)
	} else if abs >= 1000 { // Ribu
		formatted = fmt.Sprintf("Rp %.0f ribu", amount/1000)
	} else {
		formatted = FormatRupiahWithoutDecimals(amount)
	}
	
	// Clean up .0 in compact format
	formatted = strings.ReplaceAll(formatted, ".0 ", " ")
	
	return formatted
}
