package bigquery

import (
	"context"
	"math/rand"
	"time"

	"github.com/julienbreux/agy-ge-board/internal/domain"
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
	if quota <= 0 {
		quota = 12
	}
	return &DemoDataProvider{
		seed:  42,
		quota: quota,
	}
}

// FetchTelemetryLogs generates synthetic inference logs over the requested day window.
func (d *DemoDataProvider) FetchTelemetryLogs(ctx context.Context, days int) ([]domain.TelemetryLog, error) {
	if days <= 0 {
		days = 30
	}

	rng := rand.New(rand.NewSource(d.seed))
	users := []string{
		"alex.turner@example.com",
		"sophia.chen@example.com",
		"marcus.vance@example.com",
		"elena.rostova@example.com",
		"liam.oconnor@example.com",
		"dev.intern@example.com",
	}

	models := []string{
		"gemini-1.5-pro",
		"gemini-1.5-flash",
		"claude-3-5-sonnet-v2",
	}

	now := time.Now().UTC()
	var logs []domain.TelemetryLog

	for i := days; i >= 0; i-- {
		dayTime := now.AddDate(0, 0, -i)

		// Each active user generates multiple inference calls per day
		for _, u := range users {
			// dev.intern only active on rare days
			if u == "dev.intern@example.com" && rng.Float64() > 0.25 {
				continue
			}

			numCalls := rng.Intn(6) + 1
			for c := 0; c < numCalls; c++ {
				model := models[rng.Intn(len(models))]
				promptTokens := int64(rng.Intn(15000) + 1000)
				completionTokens := int64(rng.Intn(4000) + 200)

				logs = append(logs, domain.TelemetryLog{
					Timestamp:        dayTime.Add(time.Duration(rng.Intn(24*3600)) * time.Second),
					ProjectID:        "prj-antigravity-prod",
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
	if days <= 0 {
		days = 30
	}

	rng := rand.New(rand.NewSource(d.seed + 100))
	skus := []struct {
		desc    string
		baseCost float64
	}{
		{"Gemini 1.5 Pro Inference - Net Invoiced", 28.50},
		{"Gemini 1.5 Flash Inference - Net Invoiced", 12.20},
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
				Currency:       "USD",
			})
		}
	}

	return costs, nil
}

// FetchLicenseSeats returns synthetic assigned license seats and the total quota.
func (d *DemoDataProvider) FetchLicenseSeats(ctx context.Context, windowDays int) ([]domain.LicenseSeat, int, error) {
	quota := d.quota
	if quota <= 0 {
		quota = 12
	}
	now := time.Now().UTC()

	seats := []domain.LicenseSeat{
		{
			UserID:              "alex.turner@example.com",
			Assigned:            true,
			Status:              domain.SeatStatusActive,
			LastActivity:        now.Add(-2 * time.Hour),
			TotalTokensInWindow: 450000,
		},
		{
			UserID:              "sophia.chen@example.com",
			Assigned:            true,
			Status:              domain.SeatStatusActive,
			LastActivity:        now.Add(-4 * time.Hour),
			TotalTokensInWindow: 380000,
		},
		{
			UserID:              "marcus.vance@example.com",
			Assigned:            true,
			Status:              domain.SeatStatusActive,
			LastActivity:        now.AddDate(0, 0, -1),
			TotalTokensInWindow: 210000,
		},
		{
			UserID:              "elena.rostova@example.com",
			Assigned:            true,
			Status:              domain.SeatStatusActive,
			LastActivity:        now.AddDate(0, 0, -2),
			TotalTokensInWindow: 190000,
		},
		{
			UserID:              "liam.oconnor@example.com",
			Assigned:            true,
			Status:              domain.SeatStatusActive,
			LastActivity:        now.AddDate(0, 0, -3),
			TotalTokensInWindow: 120000,
		},
		{
			UserID:              "dev.intern@example.com",
			Assigned:            true,
			Status:              domain.SeatStatusAtRisk,
			LastActivity:        now.AddDate(0, 0, -12),
			TotalTokensInWindow: 850,
		},
		{
			UserID:              "inactive.dev1@example.com",
			Assigned:            true,
			Status:              domain.SeatStatusDormant,
			LastActivity:        now.AddDate(0, 0, -45),
			TotalTokensInWindow: 0,
		},
		{
			UserID:              "inactive.dev2@example.com",
			Assigned:            true,
			Status:              domain.SeatStatusDormant,
			LastActivity:        now.AddDate(0, 0, -60),
			TotalTokensInWindow: 0,
		},
	}

	return seats, quota, nil
}
