package bigquery

import (
	"cmp"
	"context"
	"math/rand/v2"
	"time"

	"github.com/julienbreux/agy-cost-board/internal/domain"
)

// DemoDataProvider generates synthetic BigQuery telemetry and billing data.
type DemoDataProvider struct {
	seed  int64
	quota int
}

// NewDemoDataProvider returns a new DemoDataProvider with default quota 12.
func NewDemoDataProvider() *DemoDataProvider {
	return NewDemoDataProviderWithQuota(12)
}

// NewDemoDataProviderWithQuota returns a new DemoDataProvider with the specified seat quota.
func NewDemoDataProviderWithQuota(quota int) *DemoDataProvider {
	return &DemoDataProvider{
		seed:  42,
		quota: cmp.Or(quota, 12),
	}
}

// FetchTelemetryLogs generates synthetic inference logs over the requested day window.
func (d *DemoDataProvider) FetchTelemetryLogs(ctx context.Context, days int) ([]domain.TelemetryLog, error) {
	days = cmp.Or(days, 30)

	// #nosec G404 -- pseudo-random numbers sufficient and deterministic for demo data generator
	rng := rand.New(rand.NewPCG(uint64(d.seed), 0))
	users := []string{
		"alex.turner@google.com",
		"sophia.chen@google.com",
		"marcus.vance@google.com",
		"elena.rostova@google.com",
		"liam.oconnor@google.com",
		"dev.intern@google.com",
	}

	models := []string{
		"gemini-3.8-flash",
		"gemini-4.0-flash",
		"gemini-4.0-pro",
		"claude-5.5-sonnet-medium",
		"claude-5.5-opus-max",
		"claude-5.5-opus-high",
	}

	now := time.Now().UTC()
	var logs []domain.TelemetryLog

	for i := days; i >= 0; i-- {
		dayTime := now.AddDate(0, 0, -i)

		// Each active user generates multiple inference calls per day
		for _, u := range users {
			// dev.intern only active on rare days
			if u == "dev.intern@google.com" && rng.Float64() > 0.25 {
				continue
			}

			numCalls := rng.IntN(6) + 1
			for range numCalls {
				model := models[rng.IntN(len(models))]
				promptTokens := int64(rng.IntN(15000) + 1000)
				completionTokens := int64(rng.IntN(4000) + 200)

				logs = append(logs, domain.TelemetryLog{
					Timestamp:        dayTime.Add(time.Duration(rng.IntN(24*3600)) * time.Second),
					ProjectID:        "agy-prod",
					UserID:           u,
					Model:            model,
					TotalTokens:      promptTokens + completionTokens,
					PromptTokens:     promptTokens,
					CompletionTokens: completionTokens,
				})
			}
		}
	}

	return logs, nil
}

// FetchBilledCosts generates synthetic billing export items corresponding to the window.
func (d *DemoDataProvider) FetchBilledCosts(ctx context.Context, days int) ([]domain.BilledCost, error) {
	days = cmp.Or(days, 30)

	// #nosec G404 -- pseudo-random numbers sufficient and deterministic for demo data generator
	rng := rand.New(rand.NewPCG(uint64(d.seed+100), 0))
	skus := []struct {
		desc     string
		baseCost float64
	}{
		{"Gemini 4.0 Pro Inference - Net Invoiced", 28.50},
		{"Gemini 3.8 Flash Inference - Net Invoiced", 12.20},
		{"Antigravity Workspace Cloud Seats & Inference", 35.00},
	}

	now := time.Now().UTC()
	var costs []domain.BilledCost

	for i := days; i >= 0; i-- {
		dateStr := now.AddDate(0, 0, -i).Format("2006-01-02")
		for _, s := range skus {
			variation := (rng.Float64() * 0.4) + 0.8 // 80% to 120% of base
			netCost := float64(int(s.baseCost*variation*100)) / 100.0

			costs = append(costs, domain.BilledCost{
				UsageDate:      dateStr,
				SKUDescription: s.desc,
				NetCost:        netCost,
				Currency:       "EUR",
			})
		}
	}

	return costs, nil
}

// FetchLicenseSeats returns synthetic assigned license seats and the total quota.
func (d *DemoDataProvider) FetchLicenseSeats(ctx context.Context, windowDays int) ([]domain.LicenseSeat, int, error) {
	quota := cmp.Or(d.quota, 12)
	now := time.Now().UTC()

	seats := []domain.LicenseSeat{
		{
			UserID:              "alex.turner@google.com",
			Assigned:            true,
			Status:              domain.SeatStatusActive,
			LastActivity:        now.Add(-2 * time.Hour),
			TotalTokensInWindow: 450000,
		},
		{
			UserID:              "sophia.chen@google.com",
			Assigned:            true,
			Status:              domain.SeatStatusActive,
			LastActivity:        now.Add(-4 * time.Hour),
			TotalTokensInWindow: 380000,
		},
		{
			UserID:              "marcus.vance@google.com",
			Assigned:            true,
			Status:              domain.SeatStatusActive,
			LastActivity:        now.AddDate(0, 0, -1),
			TotalTokensInWindow: 210000,
		},
		{
			UserID:              "elena.rostova@google.com",
			Assigned:            true,
			Status:              domain.SeatStatusActive,
			LastActivity:        now.AddDate(0, 0, -2),
			TotalTokensInWindow: 190000,
		},
		{
			UserID:              "liam.oconnor@google.com",
			Assigned:            true,
			Status:              domain.SeatStatusActive,
			LastActivity:        now.AddDate(0, 0, -3),
			TotalTokensInWindow: 120000,
		},
		{
			UserID:              "dev.intern@google.com",
			Assigned:            true,
			Status:              domain.SeatStatusAtRisk,
			LastActivity:        now.AddDate(0, 0, -12),
			TotalTokensInWindow: 850,
		},
		{
			UserID:              "inactive.dev1@google.com",
			Assigned:            true,
			Status:              domain.SeatStatusDormant,
			LastActivity:        now.AddDate(0, 0, -45),
			TotalTokensInWindow: 0,
		},
		{
			UserID:              "inactive.dev2@google.com",
			Assigned:            true,
			Status:              domain.SeatStatusDormant,
			LastActivity:        now.AddDate(0, 0, -60),
			TotalTokensInWindow: 0,
		},
	}

	return seats, quota, nil
}
