package domain_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/julienbreux/agy-cost-board/internal/domain"
)

func TestCalculateProportionalCost(t *testing.T) {
	tests := []struct {
		name              string
		userTokens        int64
		totalTokens       int64
		netCost           float64
		expectedShare     float64
		expectedAllocated float64
	}{
		{
			name:              "Standard proportional distribution",
			userTokens:        250000,
			totalTokens:       1000000,
			netCost:           100.0,
			expectedShare:     0.25,
			expectedAllocated: 25.0,
		},
		{
			name:              "Zero user tokens",
			userTokens:        0,
			totalTokens:       1000000,
			netCost:           100.0,
			expectedShare:     0.0,
			expectedAllocated: 0.0,
		},
		{
			name:              "Zero total tokens (safe division guard)",
			userTokens:        0,
			totalTokens:       0,
			netCost:           100.0,
			expectedShare:     0.0,
			expectedAllocated: 0.0,
		},
		{
			name:              "Full single user allocation",
			userTokens:        500000,
			totalTokens:       500000,
			netCost:           84.50,
			expectedShare:     1.0,
			expectedAllocated: 84.50,
		},
		{
			name:              "Rounding precision check",
			userTokens:        1,
			totalTokens:       3,
			netCost:           10.0,
			expectedShare:     0.3333333333333333,
			expectedAllocated: 3.3333,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			share, cost := domain.CalculateProportionalCost(tc.userTokens, tc.totalTokens, tc.netCost)

			// Check share within precision tolerance
			if diff := share - tc.expectedShare; diff > 0.0001 || diff < -0.0001 {
				t.Errorf("expected share ~%f, got %f", tc.expectedShare, share)
			}

			// Check cost within precision tolerance
			if diff := cost - tc.expectedAllocated; diff > 0.001 || diff < -0.001 {
				t.Errorf("expected allocated cost ~%f, got %f", tc.expectedAllocated, cost)
			}
		})
	}
}

func TestClassifySeatStatus(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name           string
		tokensInWindow int64
		lastActivity   time.Time
		windowDays     int
		expectedStatus domain.SeatStatus
	}{
		{
			name:           "Active seat with recent token consumption",
			tokensInWindow: 150000,
			lastActivity:   now.AddDate(0, 0, -2),
			windowDays:     30,
			expectedStatus: domain.SeatStatusActive,
		},
		{
			name:           "At risk seat with low consumption (< 1000 tokens)",
			tokensInWindow: 200,
			lastActivity:   now.AddDate(0, 0, -5),
			windowDays:     30,
			expectedStatus: domain.SeatStatusAtRisk,
		},
		{
			name:           "Dormant seat with zero consumption in window",
			tokensInWindow: 0,
			lastActivity:   now.AddDate(0, 0, -45),
			windowDays:     30,
			expectedStatus: domain.SeatStatusDormant,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			status := domain.ClassifySeatStatus(tc.tokensInWindow, tc.lastActivity, tc.windowDays)
			if status != tc.expectedStatus {
				t.Errorf("expected status %s, got %s", tc.expectedStatus, status)
			}
		})
	}
}

func TestAllocatedUserCostSerialization(t *testing.T) {
	item := domain.AllocatedUserCost{
		UserID:           "dev1@google.com",
		Model:            "gemini-1.5-pro",
		UsageDate:        "2026-09-20",
		UserTokens:       100000,
		TotalModelTokens: 400000,
		TokenShare:       0.25,
		AllocatedCost:    12.50,
		Currency:         "USD",
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}

	var parsed domain.AllocatedUserCost
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if parsed.UserID != item.UserID {
		t.Errorf("expected UserID %s, got %s", item.UserID, parsed.UserID)
	}
	if parsed.AllocatedCost != item.AllocatedCost {
		t.Errorf("expected AllocatedCost %f, got %f", item.AllocatedCost, parsed.AllocatedCost)
	}

	csvRow := item.ToCSVRow()
	if len(csvRow) != 7 {
		t.Errorf("expected 7 CSV columns, got %d", len(csvRow))
	}
	if csvRow[0] != "dev1@google.com" {
		t.Errorf("expected first column to be user ID, got %s", csvRow[0])
	}
}

func TestCalculateBudgetMetrics(t *testing.T) {
	t.Run("On track status when spend is well below budget", func(t *testing.T) {
		status, pct, projected := domain.CalculateBudgetMetrics(30.0, 150.0, 10)
		if status != "on_track" {
			t.Errorf("expected status on_track, got %s", status)
		}
		if pct != 20.0 {
			t.Errorf("expected pct 20.0, got %f", pct)
		}
		if projected <= 0 {
			t.Errorf("expected projected > 0, got %f", projected)
		}
	})

	t.Run("Warning status when spend is high or projected to exceed", func(t *testing.T) {
		status, pct, _ := domain.CalculateBudgetMetrics(120.0, 150.0, 20)
		if status != "warning" {
			t.Errorf("expected status warning, got %s", status)
		}
		if pct != 80.0 {
			t.Errorf("expected pct 80.0, got %f", pct)
		}
	})

	t.Run("Exceeded status when spend exceeds monthly budget", func(t *testing.T) {
		status, pct, _ := domain.CalculateBudgetMetrics(160.0, 150.0, 25)
		if status != "exceeded" {
			t.Errorf("expected status exceeded, got %s", status)
		}
		if pct != 106.7 {
			t.Errorf("expected pct 106.7, got %f", pct)
		}
	})
}

func TestGenerateOptimizationTips(t *testing.T) {
	t.Run("Generates model switching tip when Pro usage is dominant", func(t *testing.T) {
		modelBreakdown := map[string]domain.ModelCostDetail{
			"gemini-1.5-pro":   {Tokens: 800000, Cost: 80.0, Share: 0.8},
			"gemini-1.5-flash": {Tokens: 200000, Cost: 2.0, Share: 0.2},
		}
		tips := domain.GenerateOptimizationTips("user@google.com", modelBreakdown, 1000000, domain.SeatStatusActive)
		if len(tips) == 0 {
			t.Fatalf("expected optimization tips")
		}
		var foundProTip bool
		for _, tip := range tips {
			if tip.ID == "model-tier-switch" {
				foundProTip = true
				if tip.EstimatedSavingsUSD <= 0 {
					t.Errorf("expected positive savings estimate, got %f", tip.EstimatedSavingsUSD)
				}
			}
		}
		if !foundProTip {
			t.Errorf("expected model-tier-switch recommendation")
		}
	})

	t.Run("Generates license tip when user seat is dormant", func(t *testing.T) {
		tips := domain.GenerateOptimizationTips("user@google.com", nil, 0, domain.SeatStatusDormant)
		var foundLicenseTip bool
		for _, tip := range tips {
			if tip.ID == "dormant-license" {
				foundLicenseTip = true
			}
		}
		if !foundLicenseTip {
			t.Errorf("expected dormant-license recommendation")
		}
	})
}
