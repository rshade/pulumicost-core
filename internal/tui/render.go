package tui

import (
	"fmt"
	"strings"
)

// Money formatting constants.
const (
	// decimalPartsCount is the expected number of parts when splitting on decimal point.
	decimalPartsCount = 2
	// thousandsGroupSize is the number of digits in each thousands group.
	thousandsGroupSize = 3
)

// FormatMoney formats a monetary value with currency symbol and thousands separators.
// This is the primary function for displaying money values with full currency information.
//
// Usage:
//
//	FormatMoney(1234.56, "USD") // "$1,234.56 USD"
//	FormatMoney(-999.99, "EUR") // "-$999.99 EUR"
func FormatMoney(amount float64, currency string) string {
	formatted := FormatMoneyShort(amount)
	if currency == "" {
		return formatted
	}
	return fmt.Sprintf("%s %s", formatted, currency)
}

// FormatMoneyShort formats a monetary value with currency symbol but no currency name.
// Use this when you want just the dollar amount without specifying the currency type.
//
// Usage:
//
//	FormatMoneyShort(1234.56)  // "$1,234.56"
//	FormatMoneyShort(-999.99)  // "-$999.99"
//	FormatMoneyShort(0)        // "$0.00"
func FormatMoneyShort(amount float64) string {
	// Handle special cases
	if amount != amount { // NaN
		return "$0.00"
	}

	// Format with 2 decimal places
	formatted := fmt.Sprintf("%.2f", amount)

	// Handle negative numbers
	isNegative := strings.HasPrefix(formatted, "-")
	if isNegative {
		formatted = formatted[1:] // Remove the minus sign temporarily
	}

	// Add thousands separators
	parts := strings.Split(formatted, ".")
	if len(parts) != decimalPartsCount {
		result := "$" + formatted
		if isNegative {
			result = "-" + result
		}
		return result
	}

	// Add commas to integer part
	intPart := parts[0]
	if len(intPart) > thousandsGroupSize {
		var result []string
		for i, j := 0, len(intPart); i < j; i += thousandsGroupSize {
			end := j - i
			if end > thousandsGroupSize {
				end = j - i - thousandsGroupSize
			} else {
				end = 0
			}
			result = append([]string{intPart[end : j-i]}, result...)
		}
		intPart = strings.Join(result, ",")
	}

	result := fmt.Sprintf("$%s.%s", intPart, parts[1])
	if isNegative {
		result = "-" + result
	}
	return result
}

// FormatPercent formats a percentage value with one decimal place.
// Handles rounding and ensures consistent formatting for percentage displays.
//
// Usage:
//
//	FormatPercent(85.7)  // "85.7%"
//	FormatPercent(100)   // "100.0%"
//	FormatPercent(0.123) // "0.1%"
func FormatPercent(value float64) string {
	// Handle special cases
	if value != value { // NaN
		return "0.0%"
	}

	// Round to one decimal place
	rounded := fmt.Sprintf("%.1f", value)
	return rounded + "%"
}
