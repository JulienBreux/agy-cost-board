package attribution_test

import (
	"testing"
	"time"

	"github.com/julienbreux/agy-cost-board/internal/attribution"
	"github.com/julienbreux/agy-cost-board/internal/bigquery"
	"github.com/julienbreux/agy-cost-board/internal/domain"
)

func TestAttributionEngine(t *testing.T) {
	demoProvider := bigquery.NewDemoDataProvider()
	engine := attribution.NewEngine(demoProvider, 5*time.Minute)
	ctx := t.Context()

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
		summary, err := engine.GetUserSummary(ctx, "alex.turner@google.com", 30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if summary.UserID != "alex.turner@google.com" {
			t.Errorf("expected alex.turner@google.com, got %s", summary.UserID)
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

	t.Run("GetUserActivity returns sorted user logs with cost estimates", func(t *testing.T) {
		logs, err := engine.GetUserActivity(ctx, "alex.turner@google.com", 30, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(logs) == 0 {
			t.Fatalf("expected logs for alex.turner@google.com")
		}
		if len(logs) > 10 {
			t.Errorf("expected at most 10 logs, got %d", len(logs))
		}
		for i := 1; i < len(logs); i++ {
			if logs[i].Timestamp.After(logs[i-1].Timestamp) {
				t.Errorf("logs not sorted descending by timestamp: %v after %v", logs[i].Timestamp, logs[i-1].Timestamp)
			}
		}
	})

	t.Run("GetUserConsumptionDriving computes budget, velocity, and daily trends", func(t *testing.T) {
		driving, err := engine.GetUserConsumptionDriving(ctx, "alex.turner@google.com", 30, 200.0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if driving.UserID != "alex.turner@google.com" {
			t.Errorf("expected user alex.turner@google.com, got %s", driving.UserID)
		}
		if driving.TotalSpendInWindow <= 0 {
			t.Errorf("expected total spend > 0, got %f", driving.TotalSpendInWindow)
		}
		if driving.DailyBurnRate <= 0 {
			t.Errorf("expected daily burn rate > 0, got %f", driving.DailyBurnRate)
		}
		if driving.MonthlyBudget != 200.0 {
			t.Errorf("expected monthly budget 200.0, got %f", driving.MonthlyBudget)
		}
		if len(driving.DailyTrends) == 0 {
			t.Errorf("expected daily trends for user")
		}
		if len(driving.Recommendations) == 0 {
			t.Errorf("expected recommendations for user")
		}
		if driving.Currency != "EUR" {
			t.Errorf("expected driving currency EUR, got %s", driving.Currency)
		}
	})

	t.Run("Native currency propagation from billed costs", func(t *testing.T) {
		costs, err := engine.GetAttributedCosts(ctx, 30, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(costs) == 0 {
			t.Fatalf("expected costs")
		}
		if costs[0].Currency != "EUR" {
			t.Errorf("expected cost currency EUR, got %s", costs[0].Currency)
		}

		summary, err := engine.GetUserSummary(ctx, "alex.turner@google.com", 30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if summary.Currency != "EUR" {
			t.Errorf("expected summary currency EUR, got %s", summary.Currency)
		}

		overview, err := engine.GetOverviewMetrics(ctx, 30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if overview.Currency != "EUR" {
			t.Errorf("expected overview currency EUR, got %s", overview.Currency)
		}
	})

	t.Run("GetLeaderboard computes ranked developer entries with spend share, efficiency, and model mix", func(t *testing.T) {
		entries, err := engine.GetLeaderboard(ctx, 30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) == 0 {
			t.Fatalf("expected non-empty leaderboard entries")
		}

		var totalSpendShare float64
		for i, entry := range entries {
			expectedRank := i + 1
			if entry.Rank != expectedRank {
				t.Errorf("expected rank %d, got %d", expectedRank, entry.Rank)
			}
			if entry.UserID == "" {
				t.Errorf("expected non-empty UserID at rank %d", entry.Rank)
			}
			if entry.TotalCost <= 0 {
				t.Errorf("expected positive TotalCost for %s, got %f", entry.UserID, entry.TotalCost)
			}
			if entry.TotalTokens <= 0 {
				t.Errorf("expected positive TotalTokens for %s, got %d", entry.UserID, entry.TotalTokens)
			}
			if entry.ActiveDays <= 0 {
				t.Errorf("expected positive ActiveDays for %s, got %d", entry.UserID, entry.ActiveDays)
			}
			if entry.CostPer1KTokens <= 0 {
				t.Errorf("expected positive CostPer1KTokens for %s, got %f", entry.UserID, entry.CostPer1KTokens)
			}
			if entry.TopModel == "" {
				t.Errorf("expected TopModel to be set for %s", entry.UserID)
			}
			if len(entry.ModelBreakdown) == 0 {
				t.Errorf("expected ModelBreakdown to be non-empty for %s", entry.UserID)
			}
			if entry.Currency != "EUR" {
				t.Errorf("expected Currency EUR for %s, got %s", entry.UserID, entry.Currency)
			}
			if i > 0 && entry.TotalCost > entries[i-1].TotalCost {
				t.Errorf("leaderboard not sorted by TotalCost descending: rank %d (%f) > rank %d (%f)",
					entry.Rank, entry.TotalCost, entries[i-1].Rank, entries[i-1].TotalCost)
			}
			csvRow := entry.ToCSVRow()
			if len(csvRow) != 9 {
				t.Errorf("expected CSV row length 9, got %d", len(csvRow))
			}
			totalSpendShare += entry.ShareOfSpend
		}

		if totalSpendShare < 99.0 || totalSpendShare > 101.0 {
			t.Errorf("expected total spend share around 100%%, got %.2f%%", totalSpendShare)
		}
	})

	t.Run("Populates WTU, Included vs PAYG, CacheHitRate, TokenCostShares, ModelEffortStats, and ClientBreakdown", func(t *testing.T) {
		overview, err := engine.GetOverviewMetrics(ctx, 30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if overview.TotalWTU <= 0 {
			t.Errorf("expected TotalWTU > 0, got %d", overview.TotalWTU)
		}
		if overview.IncludedCost <= 0 {
			t.Errorf("expected IncludedCost > 0, got %f", overview.IncludedCost)
		}
		if overview.PAYGCost <= 0 {
			t.Errorf("expected PAYGCost > 0, got %f", overview.PAYGCost)
		}
		if overview.TotalCachedTokens <= 0 {
			t.Errorf("expected TotalCachedTokens > 0, got %d", overview.TotalCachedTokens)
		}
		if overview.CacheHitRate <= 0 {
			t.Errorf("expected CacheHitRate > 0, got %f", overview.CacheHitRate)
		}
		if len(overview.TokenCostShares) != 4 {
			t.Errorf("expected 4 TokenCostShares categories, got %d", len(overview.TokenCostShares))
		}
		if len(overview.ModelEffortStats) == 0 {
			t.Errorf("expected non-empty ModelEffortStats")
		}
		if len(overview.ClientBreakdown) == 0 {
			t.Errorf("expected non-empty ClientBreakdown")
		}

		entries, err := engine.GetLeaderboard(ctx, 30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if entries[0].WTU <= 0 {
			t.Errorf("expected leaderboard top entry WTU > 0, got %d", entries[0].WTU)
		}
		if entries[0].CacheHitRate <= 0 {
			t.Errorf("expected leaderboard top entry CacheHitRate > 0, got %f", entries[0].CacheHitRate)
		}
		if entries[0].TopEffort == "" {
			t.Errorf("expected leaderboard top entry TopEffort to be set")
		}

		driving, err := engine.GetUserConsumptionDriving(ctx, "alex.turner@google.com", 30, 200.0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if driving.TotalWTU <= 0 {
			t.Errorf("expected user driving TotalWTU > 0, got %d", driving.TotalWTU)
		}
		if driving.CacheHitRate <= 0 {
			t.Errorf("expected user driving CacheHitRate > 0, got %f", driving.CacheHitRate)
		}
		if driving.TopEffort == "" {
			t.Errorf("expected user driving TopEffort to be set")
		}
	})

	t.Run("Computes WeeklyAllowancePool, What-If scenarios, Multi-Tier seats, and FairShare vs WhenUsed PAYG", func(t *testing.T) {
		gov, err := engine.GetLicenseGovernance(ctx, 30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(gov.AllSeats) == 0 {
			t.Errorf("expected non-empty AllSeats in governance summary")
		}
		if gov.StandardSeats <= 0 || gov.PlusSeats <= 0 {
			t.Errorf("expected both StandardSeats (%d) and PlusSeats (%d) > 0", gov.StandardSeats, gov.PlusSeats)
		}
		if gov.TrialSeats <= 0 {
			t.Errorf("expected TrialSeats > 0, got %d", gov.TrialSeats)
		}
		if len(gov.Locations) < 3 {
			t.Errorf("expected at least 3 distinct locations (eu, global, us), got %v", gov.Locations)
		}
		if gov.WeeklyPool == nil {
			t.Fatalf("expected non-nil WeeklyPool in LicenseGovernanceSummary")
		}
		if gov.WeeklyPool.PoolLimitUSD <= 0 || gov.WeeklyPool.PoolLimitWTU <= 0 {
			t.Errorf("expected positive pool limits, got USD=%.2f WTU=%d", gov.WeeklyPool.PoolLimitUSD, gov.WeeklyPool.PoolLimitWTU)
		}
		if gov.WeeklyPool.ResetCountdown == "" || gov.WeeklyPool.ResetInHours <= 0 {
			t.Errorf("expected positive ResetInHours and non-empty ResetCountdown, got %d / %q",
				gov.WeeklyPool.ResetInHours, gov.WeeklyPool.ResetCountdown)
		}
		if len(gov.WeeklyPool.Windows) == 0 {
			t.Errorf("expected non-empty 7-day rolling Windows in WeeklyPool")
		}
		for _, key := range []string{"pooled_assigned", "pooled_purchased", "per_user_assigned", "per_user_purchased"} {
			sc, ok := gov.WeeklyPool.Scenarios[key]
			if !ok {
				t.Errorf("expected What-If scenario %q in WeeklyPool.Scenarios", key)
			}
			if sc.PoolLimitUSD <= 0 {
				t.Errorf("expected scenario %q PoolLimitUSD > 0, got %.2f", key, sc.PoolLimitUSD)
			}
		}
		if gov.WeeklyPool.Scenarios["pooled_purchased"].PoolLimitUSD <= gov.WeeklyPool.Scenarios["pooled_assigned"].PoolLimitUSD {
			t.Errorf("expected pooled_purchased limit (%.2f) > pooled_assigned limit (%.2f) when quota > assigned",
				gov.WeeklyPool.Scenarios["pooled_purchased"].PoolLimitUSD, gov.WeeklyPool.Scenarios["pooled_assigned"].PoolLimitUSD)
		}

		overview, err := engine.GetOverviewMetrics(ctx, 30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if overview.WeeklyPool == nil {
			t.Fatalf("expected non-nil WeeklyPool in OverviewMetrics")
		}

		entries, err := engine.GetLeaderboard(ctx, 30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, e := range entries {
			if e.FairShareIncludedCost <= 0 && e.FairSharePAYGCost <= 0 {
				t.Errorf("expected FairShare costs to be populated for %s", e.UserID)
			}
			if e.WhenUsedIncludedCost <= 0 && e.WhenUsedPAYGCost <= 0 {
				t.Errorf("expected WhenUsed costs to be populated for %s", e.UserID)
			}
		}
	})
}

func TestMonitoringQuotaMeterAndForecast(t *testing.T) {
	t.Run("BuildMonitoringQuotaMeter classifies synced, minor_drift, and diverged", func(t *testing.T) {
		synced := attribution.BuildMonitoringQuotaMeter(1012000, 1000000)
		if synced.Status != "synced" || synced.DriftPct != 1.2 || synced.MetricType != attribution.MonitoringQuotaMeterMetric {
			t.Errorf("expected synced at 1.2%% drift, got %+v", synced)
		}

		minor := attribution.BuildMonitoringQuotaMeter(1075000, 1000000)
		if minor.Status != "minor_drift" || minor.DriftPct != 7.5 {
			t.Errorf("expected minor_drift at 7.5%% drift, got %+v", minor)
		}

		diverged := attribution.BuildMonitoringQuotaMeter(1250000, 1000000)
		if diverged.Status != "diverged" || diverged.DriftPct != 25.0 {
			t.Errorf("expected diverged at 25.0%% drift, got %+v", diverged)
		}
	})

	t.Run("ComputeUsageForecast simulates 30-day window-aware spillover and resets", func(t *testing.T) {
		// Create logs with high daily pool-eligible usage across 3 days so weekly total exceeds 1,000,000 WTU pool limit
		baseTime := time.Date(2026, 10, 12, 12, 0, 0, 0, time.UTC) // Monday
		logs := []domain.TelemetryLog{
			{Timestamp: baseTime, UserID: "alice@example.com", Model: "gemini-3.1-pro", PromptTokens: 100000, CompletionTokens: 30000, QuotaType: "SHARED_POOL"},
			{Timestamp: baseTime.AddDate(0, 0, 1), UserID: "alice@example.com", Model: "gemini-3.1-pro", PromptTokens: 100000, CompletionTokens: 30000, QuotaType: "SHARED_POOL"},
			{Timestamp: baseTime.AddDate(0, 0, 2), UserID: "alice@example.com", Model: "gemini-3.1-pro", PromptTokens: 100000, CompletionTokens: 30000, QuotaType: "SHARED_POOL"},
			{Timestamp: baseTime.AddDate(0, 0, 2), UserID: "alice@example.com", Model: "claude-4.6-sonnet", PromptTokens: 20000, CompletionTokens: 10000, QuotaType: "PAYG"},
		}
		pool := &domain.WeeklyAllowancePool{
			PoolLimitWTU:    1000000,
			PoolLimitUSD:    2.00,
			IncludedUsedWTU: 800000,
			IncludedUsedUSD: 1.60,
		}
		licensed := map[string]bool{"alice@example.com": true}

		forecast := attribution.ComputeUsageForecast(logs, pool, licensed)
		if forecast == nil {
			t.Fatalf("expected non-nil UsageForecast")
		}
		if forecast.HorizonDays != 30 || len(forecast.ForecastDays) != 30 {
			t.Fatalf("expected 30 forecast days, got %d (%d)", forecast.HorizonDays, len(forecast.ForecastDays))
		}
		if forecast.DailyPoolEligibleWTU <= 0 || forecast.DailyAlwaysPAYGWTU <= 0 {
			t.Errorf("expected positive daily pool-eligible (%d) and always-PAYG (%d) WTU",
				forecast.DailyPoolEligibleWTU, forecast.DailyAlwaysPAYGWTU)
		}
		if !forecast.WillSpilloverInCurrentWindow || forecast.SpilloverDate == "" || forecast.DaysUntilSpillover < 0 {
			t.Errorf("expected spillover in current window, got willSpillover=%v date=%q days=%d",
				forecast.WillSpilloverInCurrentWindow, forecast.SpilloverDate, forecast.DaysUntilSpillover)
		}
		if forecast.Projected30dTotalWTU <= 0 || forecast.Projected30dTotalUSD <= 0 {
			t.Errorf("expected positive 30d totals, got WTU=%d USD=%.2f",
				forecast.Projected30dTotalWTU, forecast.Projected30dTotalUSD)
		}

		// Verify RecomputeForecastForPool with a much larger pool limit eliminates spillover
		largePool := &domain.WeeklyAllowancePool{
			PoolLimitWTU:    50000000,
			PoolLimitUSD:    100.00,
			IncludedUsedWTU: 600000,
			IncludedUsedUSD: 1.20,
		}
		recomputed := attribution.RecomputeForecastForPool(forecast, largePool)
		if recomputed.WillSpilloverInCurrentWindow || recomputed.SpilloverDate != "" || recomputed.DaysUntilSpillover != -1 {
			t.Errorf("expected no spillover with 50M WTU pool limit, got willSpillover=%v date=%q days=%d",
				recomputed.WillSpilloverInCurrentWindow, recomputed.SpilloverDate, recomputed.DaysUntilSpillover)
		}
	})
}

func TestEngineFilterByProject(t *testing.T) {
	ctx := t.Context()
	provider := bigquery.NewDemoDataProvider()
	engine := attribution.NewEngine(provider, 1*time.Minute)
	engine.SetMonitoredProjects([]string{"gcp-ai-platform-prod", "gcp-finops-analytics", "gcp-agent-sandbox"})

	allOverview, err := engine.GetOverviewMetrics(ctx, 30)
	if err != nil {
		t.Fatalf("unexpected error for all projects: %v", err)
	}
	if allOverview.SelectedProject != "all" {
		t.Errorf("expected SelectedProject='all', got %q", allOverview.SelectedProject)
	}
	if len(allOverview.MonitoredProjects) < 3 {
		t.Errorf("expected at least 3 MonitoredProjects, got %v", allOverview.MonitoredProjects)
	}
	if allOverview.Forecast == nil || len(allOverview.Forecast.ForecastDays) != 30 {
		t.Errorf("expected 30-day Forecast on OverviewMetrics, got %+v", allOverview.Forecast)
	}
	if allOverview.WeeklyPool == nil || allOverview.WeeklyPool.MonitoringMeter == nil {
		t.Errorf("expected MonitoringMeter on WeeklyPool, got %+v", allOverview.WeeklyPool)
	}

	scopedEngine := engine.FilterByProject("gcp-ai-platform-prod")
	prodOverview, err := scopedEngine.GetOverviewMetrics(ctx, 30)
	if err != nil {
		t.Fatalf("unexpected error for gcp-ai-platform-prod: %v", err)
	}
	if prodOverview.SelectedProject != "gcp-ai-platform-prod" {
		t.Errorf("expected SelectedProject='gcp-ai-platform-prod', got %q", prodOverview.SelectedProject)
	}
	if len(prodOverview.MonitoredProjects) != len(allOverview.MonitoredProjects) {
		t.Errorf("expected scoped engine to preserve full MonitoredProjects list (%v), got %v",
			allOverview.MonitoredProjects, prodOverview.MonitoredProjects)
	}
	if prodOverview.TotalTokens <= 0 || prodOverview.TotalTokens >= allOverview.TotalTokens {
		t.Errorf("expected scoped TotalTokens (%d) to be > 0 and < allOverview.TotalTokens (%d)",
			prodOverview.TotalTokens, allOverview.TotalTokens)
	}
	if prodOverview.AssignedSeats <= 0 || prodOverview.AssignedSeats >= allOverview.AssignedSeats {
		t.Errorf("expected scoped AssignedSeats (%d) to be > 0 and < allOverview.AssignedSeats (%d)",
			prodOverview.AssignedSeats, allOverview.AssignedSeats)
	}

	prodCosts, err := scopedEngine.GetAttributedCosts(ctx, 30, "")
	if err != nil {
		t.Fatalf("unexpected error for scoped costs: %v", err)
	}
	for _, c := range prodCosts {
		if c.ProjectID != "gcp-ai-platform-prod" {
			t.Errorf("expected all scoped costs to have ProjectID='gcp-ai-platform-prod', got %q", c.ProjectID)
		}
	}

	// Verify filtering by a project with zero telemetry logs returns non-nil empty slices and zero-spend dashboard
	emptyScopedEngine := engine.FilterByProject("julienbreux-ai")
	emptyCosts, err := emptyScopedEngine.GetAttributedCosts(ctx, 30, "")
	if err != nil {
		t.Fatalf("unexpected error for empty project costs: %v", err)
	}
	if emptyCosts == nil {
		t.Errorf("expected non-nil empty slice for GetAttributedCosts on project with zero logs, got nil")
	}
	if len(emptyCosts) != 0 {
		t.Errorf("expected 0 costs for empty project, got %d", len(emptyCosts))
	}

	driving, err := emptyScopedEngine.GetUserConsumptionDriving(ctx, "alice@example.com", 30, 100.0)
	if err != nil {
		t.Fatalf("expected GetUserConsumptionDriving to succeed on empty project, got error: %v", err)
	}
	if driving.TotalSpendInWindow != 0 {
		t.Errorf("expected TotalSpendInWindow=0 on empty project, got %f", driving.TotalSpendInWindow)
	}
	if driving.DailyTrends == nil {
		t.Errorf("expected non-nil DailyTrends slice on empty project")
	}
	if driving.Currency != "EUR" {
		t.Errorf("expected empty project to preserve native billing currency 'EUR' from baseProvider, got %q", driving.Currency)
	}
}

