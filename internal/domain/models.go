package domain

import (
	"fmt"
	"math"
	"time"
)

// SeatStatus represents the activity status of an assigned license seat.
type SeatStatus string

const (
	SeatStatusActive  SeatStatus = "active"
	SeatStatusAtRisk  SeatStatus = "at_risk"
	SeatStatusDormant SeatStatus = "dormant"
)

// TelemetryLog represents an individual Antigravity inference event from Cloud Logging.
type TelemetryLog struct {
	Timestamp        time.Time `json:"timestamp"`
	ProjectID        string    `json:"project_id"`
	UserID           string    `json:"user_id"`
	Model            string    `json:"model"`
	TotalTokens      int64     `json:"total_tokens"`
	PromptTokens     int64     `json:"prompt_tokens"`
	CompletionTokens int64     `json:"completion_tokens"`
}

// BilledCost represents net billed costs per SKU from Cloud Billing export.
type BilledCost struct {
	UsageDate      string  `json:"usage_date"`
	SKUDescription string  `json:"sku_description"`
	NetCost        float64 `json:"net_cost"`
	Currency       string  `json:"currency"`
}

// AllocatedUserCost represents the calculated proportional cost for a single user, model, and date.
type AllocatedUserCost struct {
	UserID           string  `json:"user_id"`
	Model            string  `json:"model"`
	UsageDate        string  `json:"usage_date"`
	UserTokens       int64   `json:"user_tokens"`
	TotalModelTokens int64   `json:"total_model_tokens"`
	TokenShare       float64 `json:"token_share"`
	AllocatedCost    float64 `json:"allocated_cost"`
	Currency         string  `json:"currency"`
}

// ToCSVRow converts the allocated cost item to a slice of strings for CSV output.
func (a AllocatedUserCost) ToCSVRow() []string {
	return []string{
		a.UserID,
		a.Model,
		a.UsageDate,
		fmt.Sprintf("%d", a.UserTokens),
		fmt.Sprintf("%.4f", a.TokenShare),
		fmt.Sprintf("%.4f", a.AllocatedCost),
		a.Currency,
	}
}

// UserCostSummary represents aggregated metrics for an individual developer.
type UserCostSummary struct {
	UserID         string                     `json:"user_id"`
	TotalTokens    int64                      `json:"total_tokens"`
	TotalCost      float64                    `json:"total_cost"`
	Currency       string                     `json:"currency"`
	ModelBreakdown map[string]ModelCostDetail `json:"model_breakdown"`
	LastActive     time.Time                  `json:"last_active"`
	SeatStatus     SeatStatus                 `json:"seat_status"`
}

// ModelCostDetail holds metrics for a specific model under a user.
type ModelCostDetail struct {
	Tokens int64   `json:"tokens"`
	Cost   float64 `json:"cost"`
	Share  float64 `json:"share"`
}

// LicenseSeat tracks the assignment and activity for a developer's seat.
type LicenseSeat struct {
	UserID              string     `json:"user_id"`
	Assigned            bool       `json:"assigned"`
	Status              SeatStatus `json:"status"`
	LastActivity        time.Time  `json:"last_activity"`
	TotalTokensInWindow int64      `json:"total_tokens_in_window"`
}

// LicenseGovernanceSummary provides high-level seat quota and utilization metrics.
type LicenseGovernanceSummary struct {
	SeatQuota               int           `json:"seat_quota"`
	AssignedSeats           int           `json:"assigned_seats"`
	ActiveSeats             int           `json:"active_seats"`
	DormantSeats            int           `json:"dormant_seats"`
	UtilizationPct          float64       `json:"utilization_pct"`
	EstimatedMonthlySavings float64       `json:"estimated_monthly_savings"`
	DormantUsers            []LicenseSeat `json:"dormant_users"`
}

// OverviewMetrics holds top-level KPIs for the web dashboard and CLI summary.
type OverviewMetrics struct {
	TotalBilledCost         float64                    `json:"total_billed_cost"`
	TotalTokens             int64                      `json:"total_tokens"`
	ActiveUsersCount        int                        `json:"active_users_count"`
	SeatQuota               int                        `json:"seat_quota"`
	AssignedSeats           int                        `json:"assigned_seats"`
	UtilizationPct          float64                    `json:"utilization_pct"`
	EstimatedMonthlySavings float64                    `json:"estimated_monthly_savings"`
	DailyTrends             []DailySpendTrend          `json:"daily_trends"`
	ModelBreakdown          map[string]ModelCostDetail `json:"model_breakdown"`
	Currency                string                     `json:"currency"`
}

// DailySpendTrend tracks daily spend and token volume for trend charts.
type DailySpendTrend struct {
	Date        string  `json:"date"`
	TotalCost   float64 `json:"total_cost"`
	TotalTokens int64   `json:"total_tokens"`
	ActiveUsers int     `json:"active_users"`
}

// CalculateProportionalCost calculates token share and allocated cost safely.
// Guaranteed not to divide by zero.
func CalculateProportionalCost(userTokens, totalTokens int64, netCost float64) (tokenShare, allocatedCost float64) {
	if totalTokens <= 0 || userTokens <= 0 || netCost <= 0 {
		return 0.0, 0.0
	}

	tokenShare = float64(userTokens) / float64(totalTokens)
	allocatedCost = math.Round((tokenShare*netCost)*10000) / 10000

	return tokenShare, allocatedCost
}

// ClassifySeatStatus classifies user seat activity based on token threshold and inactivity.
func ClassifySeatStatus(tokensInWindow int64, lastActivity time.Time, windowDays int) SeatStatus {
	cutoff := time.Now().AddDate(0, 0, -windowDays)

	if tokensInWindow == 0 || lastActivity.Before(cutoff) {
		return SeatStatusDormant
	}

	if tokensInWindow < 1000 {
		return SeatStatusAtRisk
	}

	return SeatStatusActive
}
