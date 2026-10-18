package bigquery_test

import (
	"testing"

	"github.com/julienbreux/agy-cost-board/internal/bigquery"
	"github.com/julienbreux/agy-cost-board/internal/domain"
)

func TestDemoDataProvider(t *testing.T) {
	provider := bigquery.NewDemoDataProvider()
	ctx := t.Context()

	t.Run("FetchTelemetryLogs generates logs for requested window", func(t *testing.T) {
		logs, err := provider.FetchTelemetryLogs(ctx, 30)
		if err != nil {
			t.Fatalf("unexpected error fetching demo logs: %v", err)
		}

		if len(logs) == 0 {
			t.Fatalf("expected telemetry logs to be generated, got 0")
		}

		// Verify distinct users and models exist
		userMap := make(map[string]bool)
		modelMap := make(map[string]bool)
		effortMap := make(map[string]bool)
		clientMap := make(map[string]bool)
		quotaMap := make(map[string]bool)
		var totalTokens, totalCachedTokens int64

		for _, l := range logs {
			userMap[l.UserID] = true
			modelMap[l.Model] = true
			effortMap[l.EffortLevel] = true
			clientMap[l.ClientName] = true
			quotaMap[l.QuotaType] = true
			totalTokens += l.TotalTokens
			totalCachedTokens += l.CachedTokens
		}

		if len(userMap) < 4 {
			t.Errorf("expected at least 4 active demo users, got %d", len(userMap))
		}
		if !modelMap["gemini-4.0-pro"] {
			t.Errorf("expected gemini-4.0-pro model in telemetry logs")
		}
		if totalTokens <= 0 {
			t.Errorf("expected positive total token volume, got %d", totalTokens)
		}
		if totalCachedTokens <= 0 {
			t.Errorf("expected positive cached token volume in demo logs, got %d", totalCachedTokens)
		}
		if len(effortMap) < 2 {
			t.Errorf("expected multiple effort levels in demo logs, got %v", effortMap)
		}
		if len(clientMap) < 2 {
			t.Errorf("expected multiple client names in demo logs, got %v", clientMap)
		}
		if !quotaMap["REGULAR_QUOTA"] || !quotaMap["OVERAGE_QUOTA"] {
			t.Errorf("expected both REGULAR_QUOTA and OVERAGE_QUOTA in demo logs, got %v", quotaMap)
		}
	})

	t.Run("FetchBilledCosts generates matching billing entries", func(t *testing.T) {
		costs, err := provider.FetchBilledCosts(ctx, 30)
		if err != nil {
			t.Fatalf("unexpected error fetching demo billing: %v", err)
		}

		if len(costs) == 0 {
			t.Fatalf("expected billing entries to be generated, got 0")
		}

		var totalNetCost float64
		for _, c := range costs {
			totalNetCost += c.NetCost
			if c.Currency != "EUR" {
				t.Errorf("expected EUR currency, got %s", c.Currency)
			}
		}

		if totalNetCost <= 0 {
			t.Errorf("expected positive total billed cost, got %f", totalNetCost)
		}
	})

	t.Run("FetchLicenseSeats returns assigned and dormant seats", func(t *testing.T) {
		seats, quota, err := provider.FetchLicenseSeats(ctx, 30)
		if err != nil {
			t.Fatalf("unexpected error fetching license seats: %v", err)
		}

		if quota <= 0 {
			t.Errorf("expected positive seat quota, got %d", quota)
		}

		var dormantCount int
		for _, s := range seats {
			if s.Status == domain.SeatStatusDormant {
				dormantCount++
			}
		}

		if dormantCount == 0 {
			t.Errorf("expected at least one dormant seat in demo data")
		}
	})
}
