package domain_test

import (
	"strings"
	"testing"

	"github.com/julienbreux/agy-cost-board/internal/domain"
)

func TestCurrencySymbol(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{"USD", "$"},
		{"usd", "$"},
		{"", "$"},
		{"EUR", "€"},
		{"eur", "€"},
		{"GBP", "£"},
		{"JPY", "¥"},
		{"CAD", "CA$"},
		{"AUD", "A$"},
		{"CHF", "CHF "},
		{"INR", "₹"},
		{"BRL", "R$"},
		{"SEK", "kr "},
		{"NOK", "kr "},
		{"DKK", "kr "},
		{"XYZ", "XYZ "},
	}

	for _, tc := range tests {
		t.Run(tc.code, func(t *testing.T) {
			got := domain.CurrencySymbol(tc.code)
			if got != tc.expected {
				t.Errorf("CurrencySymbol(%q) = %q; want %q", tc.code, got, tc.expected)
			}
		})
	}
}

func TestFormatCost(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		currency string
		expected string
	}{
		{"USD standard", 123.45, "USD", "$123.45"},
		{"USD zero", 0.0, "USD", "$0.00"},
		{"USD default empty", 45.0, "", "$45.00"},
		{"EUR symbol", 89.99, "EUR", "€89.99"},
		{"GBP symbol", 15.5, "GBP", "£15.50"},
		{"JPY symbol", 1500.0, "JPY", "¥1500.00"},
		{"CAD symbol", 42.10, "CAD", "CA$42.10"},
		{"Unknown ISO code prefix", 77.25, "XYZ", "XYZ 77.25"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := domain.FormatCost(tc.amount, tc.currency)
			if got != tc.expected {
				t.Errorf("FormatCost(%f, %q) = %q; want %q", tc.amount, tc.currency, got, tc.expected)
			}
		})
	}
}

func TestGenerateOptimizationTips_Currency(t *testing.T) {
	t.Run("Dormant license tip uses dynamic currency symbol", func(t *testing.T) {
		tips := domain.GenerateOptimizationTips("user@google.com", nil, 0, domain.SeatStatusDormant, "EUR")
		if len(tips) == 0 {
			t.Fatalf("expected tips")
		}
		var foundDormant bool
		for _, tip := range tips {
			if tip.ID == "dormant-license" {
				foundDormant = true
				if !strings.Contains(tip.Description, "€45/mo") {
					t.Errorf("expected tip description to contain €45/mo, got: %s", tip.Description)
				}
			}
		}
		if !foundDormant {
			t.Errorf("dormant-license tip not found")
		}
	})
}
