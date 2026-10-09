package attribution_test

import (
	"context"
	"testing"
	"time"

	"github.com/julienbreux/agy-cost-board/internal/attribution"
	"github.com/julienbreux/agy-cost-board/internal/bigquery"
)

func TestAttributionEngine(t *testing.T) {
	demoProvider := bigquery.NewDemoDataProvider()
	engine := attribution.NewEngine(demoProvider, 5*time.Minute)
	ctx := context.Background()

	t.Run("GetAttributedCosts computes non-empty proportional allocations", func(t *testing.T) {
		costs, err := engine.GetAttributedCosts(ctx, 30, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(costs) == 0 {
			t.Fatalf("expected attributed cost entries, got 0")
		}

		// Verify every cost has positive allocated cost and share <= 1.0
		for _, c := range costs {
			if c.TokenShare < 0.0 || c.TokenShare > 1.0001 {
				t.Errorf("token share out of range [0, 1]: %f", c.TokenShare)
			}
			if c.AllocatedCost < 0 {
				t.Errorf("allocated cost cannot be negative: %f", c.AllocatedCost)
			}
			if c.UserID == "" {
				t.Errorf("user ID cannot be empty")
			}
		}
	})

	t.Run("GetAttributedCosts filters by model", func(t *testing.T) {
		filter := "gemini-1.5-pro"
		costs, err := engine.GetAttributedCosts(ctx, 30, filter)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for _, c := range costs {
			if c.Model != filter {
				t.Errorf("expected model %s, got %s", filter, c.Model)
			}
		}
	})

	t.Run("GetUserSummary computes detailed metrics for a user", func(t *testing.T) {
		summary, err := engine.GetUserSummary(ctx, "alex.turner@example.com", 30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if summary.UserID != "alex.turner@example.com" {
			t.Errorf("expected alex.turner@example.com, got %s", summary.UserID)
		}
		if summary.TotalTokens <= 0 {
			t.Errorf("expected positive total tokens, got %d", summary.TotalTokens)
		}
		if summary.TotalCost <= 0 {
			t.Errorf("expected positive total cost, got %f", summary.TotalCost)
		}
		if len(summary.ModelBreakdown) == 0 {
			t.Errorf("expected model breakdown entries")
		}
	})

	t.Run("GetLicenseGovernance returns active and dormant counts with savings", func(t *testing.T) {
		gov, err := engine.GetLicenseGovernance(ctx, 30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if gov.SeatQuota <= 0 {
			t.Errorf("expected positive seat quota, got %d", gov.SeatQuota)
		}
		if gov.AssignedSeats <= 0 {
			t.Errorf("expected assigned seats > 0, got %d", gov.AssignedSeats)
		}
		if gov.DormantSeats <= 0 {
			t.Errorf("expected dormant seats > 0 in demo data, got %d", gov.DormantSeats)
		}
		if gov.EstimatedMonthlySavings <= 0 {
			t.Errorf("expected positive monthly savings for dormant seats, got %f", gov.EstimatedMonthlySavings)
		}
	})

	t.Run("GetOverviewMetrics provides high-level KPIs and daily trends", func(t *testing.T) {
		overview, err := engine.GetOverviewMetrics(ctx, 30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if overview.TotalBilledCost <= 0 {
			t.Errorf("expected total billed cost > 0, got %f", overview.TotalBilledCost)
		}
		if overview.TotalTokens <= 0 {
			t.Errorf("expected total tokens > 0, got %d", overview.TotalTokens)
		}
		if len(overview.DailyTrends) == 0 {
			t.Errorf("expected daily spend trends")
		}
	})

	t.Run("TTL Cache hits return cached data", func(t *testing.T) {
		// First call populates cache
		overview1, err := engine.GetOverviewMetrics(ctx, 30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Second call should hit cache immediately
		overview2, err := engine.GetOverviewMetrics(ctx, 30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if overview1.TotalBilledCost != overview2.TotalBilledCost {
			t.Errorf("cache inconsistency: %f vs %f", overview1.TotalBilledCost, overview2.TotalBilledCost)
		}
	})
}
