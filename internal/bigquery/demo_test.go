package bigquery_test

import (
	"context"
	"testing"

	"github.com/julienbreux/agy-ge-board/internal/bigquery"
	"github.com/julienbreux/agy-ge-board/internal/domain"
)

func TestDemoDataProvider(t *testing.T) {
	provider := bigquery.NewDemoDataProvider()
	ctx := context.Background()

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
		var totalTokens int64

		for _, l := range logs {
			userMap[l.UserID] = true
			modelMap[l.Model] = true
			totalTokens += l.TotalTokens
		}

		if len(userMap) < 4 {
			t.Errorf("expected at least 4 active demo users, got %d", len(userMap))
		}
		if !modelMap["gemini-1.5-pro"] {
			t.Errorf("expected gemini-1.5-pro model in telemetry logs")
		}
		if totalTokens <= 0 {
			t.Errorf("expected positive total token volume, got %d", totalTokens)
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
			if c.Currency != "USD" {
				t.Errorf("expected USD currency, got %s", c.Currency)
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
