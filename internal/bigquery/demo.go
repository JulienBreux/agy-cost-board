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

// DemoProjects lists the synthetic GCP projects used in --demo mode.
var DemoProjects = []string{
	"gcp-ai-platform-prod",
	"gcp-finops-analytics",
	"gcp-agent-sandbox",
}

// FetchTelemetryLogs generates synthetic inference logs over the requested day window.
func (d *DemoDataProvider) FetchTelemetryLogs(ctx context.Context, days int) ([]domain.TelemetryLog, error) {
	days = cmp.Or(days, 30)

	// #nosec G404 -- pseudo-random numbers sufficient and deterministic for demo data generator
	rng := rand.New(rand.NewPCG(uint64(d.seed), 0))
	users := []struct {
		id             string
		primaryProject string
	}{
		{"alex.turner@google.com", "gcp-ai-platform-prod"},
		{"sophia.chen@google.com", "gcp-ai-platform-prod"},
		{"marcus.vance@google.com", "gcp-finops-analytics"},
		{"elena.rostova@google.com", "gcp-finops-analytics"},
		{"liam.oconnor@google.com", "gcp-agent-sandbox"},
		{"dev.intern@google.com", "gcp-agent-sandbox"},
	}

	models := []string{
		"gemini-3.8-flash",
		"gemini-4.0-flash",
		"gemini-4.0-pro",
		"claude-5.5-sonnet-medium",
		"claude-5.5-opus-max",
		"claude-5.5-opus-high",
	}

	efforts := []string{"low", "medium", "high", "max"}
	clients := []struct {
		name    string
		version string
	}{
		{"vscode", "2.18.0"},
		{"jetbrains", "2.14.2"},
		{"agy-cli", "0.9.3"},
		{"agy-sdk", "1.4.0"},
	}

	now := time.Now().UTC()
	var logs []domain.TelemetryLog

	for i := days; i >= 0; i-- {
		dayTime := now.AddDate(0, 0, -i)

		// Each active user generates multiple inference calls per day
		for _, u := range users {
			// dev.intern only active on rare days
			if u.id == "dev.intern@google.com" && rng.Float64() > 0.25 {
				continue
			}

			numCalls := rng.IntN(6) + 1
			for range numCalls {
				model := models[rng.IntN(len(models))]
				promptTokens := int64(rng.IntN(15000) + 1000)
				cachedTokens := int64(float64(promptTokens) * (0.20 + rng.Float64()*0.50)) // 20% to 70% cache hit
				completionTokens := int64(rng.IntN(4000) + 200)
				thoughtsTokens := int64(rng.IntN(2500))

				effort := domain.ExtractEffortLevel(model)
				if effort == "default" {
					effort = efforts[rng.IntN(len(efforts))]
				}

				quotaType := "REGULAR_QUOTA"
				if rng.Float64() < 0.20 {
					quotaType = "OVERAGE_QUOTA"
				}

				client := clients[rng.IntN(len(clients))]
				projectID := u.primaryProject
				if rng.Float64() < 0.20 {
					projectID = DemoProjects[rng.IntN(len(DemoProjects))]
				}

				logs = append(logs, domain.TelemetryLog{
					Timestamp:        dayTime.Add(time.Duration(rng.IntN(24*3600)) * time.Second),
					ProjectID:        projectID,
					UserID:           u.id,
					Model:            model,
					TotalTokens:      promptTokens + completionTokens + thoughtsTokens,
					PromptTokens:     promptTokens,
					CachedTokens:     cachedTokens,
					CompletionTokens: completionTokens,
					ThoughtsTokens:   thoughtsTokens,
					QuotaType:        quotaType,
					EffortLevel:      effort,
					ClientName:       client.name,
					ClientVersion:    client.version,
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
	projectWeights := []struct {
		id     string
		weight float64
	}{
		{"gcp-ai-platform-prod", 0.50},
		{"gcp-finops-analytics", 0.30},
		{"gcp-agent-sandbox", 0.20},
	}

	now := time.Now().UTC()
	var costs []domain.BilledCost

	for i := days; i >= 0; i-- {
		dateStr := now.AddDate(0, 0, -i).Format("2006-01-02")
		for _, s := range skus {
			variation := (rng.Float64() * 0.4) + 0.8 // 80% to 120% of base
			dayNetCost := s.baseCost * variation
			for _, pw := range projectWeights {
				netCost := float64(int(dayNetCost*pw.weight*100)) / 100.0
				costs = append(costs, domain.BilledCost{
					UsageDate:      dateStr,
					ProjectID:      pw.id,
					SKUDescription: s.desc,
					NetCost:        netCost,
					Currency:       "EUR",
				})
			}
		}
	}

	return costs, nil
}

// FetchLicenseSeats returns synthetic assigned license seats and the total quota.
func (d *DemoDataProvider) FetchLicenseSeats(ctx context.Context, windowDays int) ([]domain.LicenseSeat, int, error) {
	quota := cmp.Or(d.quota, 12)
	now := time.Now().UTC()
	trialEnd := now.AddDate(0, 0, 14).Format("2006-01-02")

	rawSeats := []domain.LicenseSeat{
		{
			ProjectID:           "gcp-ai-platform-prod",
			UserID:              "alex.turner@google.com",
			Assigned:            true,
			Status:              domain.SeatStatusActive,
			LastActivity:        now.Add(-2 * time.Hour),
			TotalTokensInWindow: 450000,
			Tier:                "plus",
			Location:            "global",
		},
		{
			ProjectID:           "gcp-ai-platform-prod",
			UserID:              "sophia.chen@google.com",
			Assigned:            true,
			Status:              domain.SeatStatusActive,
			LastActivity:        now.Add(-4 * time.Hour),
			TotalTokensInWindow: 380000,
			Tier:                "plus",
			Location:            "us",
		},
		{
			ProjectID:           "gcp-finops-analytics",
			UserID:              "marcus.vance@google.com",
			Assigned:            true,
			Status:              domain.SeatStatusActive,
			LastActivity:        now.AddDate(0, 0, -1),
			TotalTokensInWindow: 210000,
			Tier:                "standard",
			Location:            "eu",
		},
		{
			ProjectID:           "gcp-finops-analytics",
			UserID:              "elena.rostova@google.com",
			Assigned:            true,
			Status:              domain.SeatStatusActive,
			LastActivity:        now.AddDate(0, 0, -2),
			TotalTokensInWindow: 190000,
			Tier:                "plus",
			Location:            "eu",
		},
		{
			ProjectID:           "gcp-agent-sandbox",
			UserID:              "liam.oconnor@google.com",
			Assigned:            true,
			Status:              domain.SeatStatusActive,
			LastActivity:        now.AddDate(0, 0, -3),
			TotalTokensInWindow: 120000,
			Tier:                "standard",
			Location:            "us",
		},
		{
			ProjectID:           "gcp-agent-sandbox",
			UserID:              "dev.intern@google.com",
			Assigned:            true,
			Status:              domain.SeatStatusAtRisk,
			LastActivity:        now.AddDate(0, 0, -12),
			TotalTokensInWindow: 850,
			Tier:                "standard",
			Location:            "global",
			IsTrial:             true,
			TrialEndsAt:         trialEnd,
		},
		{
			ProjectID:           "gcp-ai-platform-prod",
			UserID:              "inactive.dev1@google.com",
			Assigned:            true,
			Status:              domain.SeatStatusDormant,
			LastActivity:        now.AddDate(0, 0, -45),
			TotalTokensInWindow: 0,
			Tier:                "plus",
			Location:            "global",
		},
		{
			ProjectID:           "gcp-finops-analytics",
			UserID:              "inactive.dev2@google.com",
			Assigned:            true,
			Status:              domain.SeatStatusDormant,
			LastActivity:        now.AddDate(0, 0, -60),
			TotalTokensInWindow: 0,
			Tier:                "standard",
			Location:            "eu",
		},
	}

	seats := make([]domain.LicenseSeat, 0, len(rawSeats))
	for _, s := range rawSeats {
		tier, usd, wtu := domain.ResolveSeatAllowance(s.Tier)
		s.Tier = tier
		s.WeeklyAllowanceUSD = usd
		s.WeeklyAllowanceWTU = wtu
		seats = append(seats, s)
	}

	return seats, quota, nil
}

