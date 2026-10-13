package domain

import (
	"fmt"
	"strings"
)

// CurrencySymbol maps standard ISO 4217 currency codes to their familiar symbols.
// For known currencies with unambiguous glyphs, it returns the symbol (e.g. $, €, £, ¥, ₹).
// For other valid ISO codes, it safely falls back to the code with a trailing space (e.g. "CHF ", "XYZ ").
func CurrencySymbol(code string) string {
	trimmed := strings.ToUpper(strings.TrimSpace(code))
	switch trimmed {
	case "USD", "":
		return "$"
	case "EUR":
		return "€"
	case "GBP":
		return "£"
	case "JPY", "CNY":
		return "¥"
	case "CAD":
		return "CA$"
	case "AUD":
		return "A$"
	case "CHF":
		return "CHF "
	case "INR":
		return "₹"
	case "BRL":
		return "R$"
	case "SEK", "NOK", "DKK":
		return "kr "
	default:
		return trimmed + " "
	}
}

// FormatCost formats a monetary amount with the appropriate currency symbol and 2 decimal places.
func FormatCost(amount float64, currency string) string {
	symbol := CurrencySymbol(currency)
	return fmt.Sprintf("%s%.2f", symbol, amount)
}
