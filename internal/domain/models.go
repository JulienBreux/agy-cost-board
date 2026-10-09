package domain

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// ErrUserNotFound indicates that the requested user has no activity or records.
var ErrUserNotFound = errors.New("user not found")

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
		strconv.FormatInt(a.UserTokens, 10),
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
	Cost        float64 `json:"cost"`
	TotalTokens int64   `json:"total_tokens"`
	Tokens      int64   `json:"tokens"`
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

// CurrentUserIdentity represents the resolved identity of the authenticated or current session user.
type CurrentUserIdentity struct {
	UserID        string `json:"user_id"`
	Email         string `json:"email"`
	Authenticated bool   `json:"authenticated"`
	Source        string `json:"source"` // "iap", "header", "demo", "none"
}

// UserActivityLog represents an individual inference log event for a user with estimated cost.
type UserActivityLog struct {
	Timestamp        time.Time `json:"timestamp"`
	UserID           string    `json:"user_id"`
	Model            string    `json:"model"`
	TotalTokens      int64     `json:"total_tokens"`
	PromptTokens     int64     `json:"prompt_tokens"`
	CompletionTokens int64     `json:"completion_tokens"`
	EstimatedCost    float64   `json:"estimated_cost"`
}

// UserConsumptionDriving represents personal driving metrics, budget tracking, and recommendations.
type UserConsumptionDriving struct {
	UserID                 string               `json:"user_id"`
	Currency               string               `json:"currency"`
	TotalSpendInWindow     float64              `json:"total_spend_in_window"`
	TotalTokensInWindow    int64                `json:"total_tokens_in_window"`
	DailyBurnRate          float64              `json:"daily_burn_rate"`
	WeeklyBurnRate         float64              `json:"weekly_burn_rate"`
	OrgSpendSharePct       float64              `json:"org_spend_share_pct"`
	MonthlyBudget          float64              `json:"monthly_budget"`
	ProjectedMonthEndSpend float64              `json:"projected_month_end_spend"`
	BudgetConsumedPct      float64              `json:"budget_consumed_pct"`
	BudgetStatus           string               `json:"budget_status"` // "on_track", "warning", "exceeded"
	Recommendations        []OptimizationTip    `json:"recommendations"`
	DailyTrends            []PersonalDailyTrend `json:"daily_trends"`
}

// PersonalDailyTrend records day-by-day cost and token consumption for a single user.
type PersonalDailyTrend struct {
	Date        string             `json:"date"`
	TotalCost   float64            `json:"total_cost"`
	TotalTokens int64              `json:"total_tokens"`
	ByModel     map[string]float64 `json:"by_model"`
}

// OptimizationTip provides an actionable insight to reduce cost or improve token efficiency.
type OptimizationTip struct {
	ID                  string  `json:"id"`
	Title               string  `json:"title"`
	Description         string  `json:"description"`
	Severity            string  `json:"severity"` // "info", "warning", "success"
	EstimatedSavingsUSD float64 `json:"estimated_savings_usd"`
}

// CalculateBudgetMetrics computes budget consumption percentage, burn rate projection, and health status.
func CalculateBudgetMetrics(spend float64, budget float64, days int) (status string, consumedPct float64, projectedMonthEnd float64) {
	if budget <= 0 {
		budget = 150.0
	}
	if days <= 0 {
		days = 30
	}

	consumedPct = math.Round((spend/budget)*1000) / 10.0
	dailyBurn := spend / float64(days)
	projectedMonthEnd = math.Round((dailyBurn*30.0)*100) / 100

	if consumedPct >= 100.0 {
		status = "exceeded"
	} else if consumedPct >= 75.0 || projectedMonthEnd > budget {
		status = "warning"
	} else {
		status = "on_track"
	}

	return status, consumedPct, projectedMonthEnd
}

// GenerateOptimizationTips evaluates user telemetry to generate contextual cost-saving tips.
func GenerateOptimizationTips(userID string, modelBreakdown map[string]ModelCostDetail, totalTokens int64, seatStatus SeatStatus) []OptimizationTip {
	tips := make([]OptimizationTip, 0)

	// 1. Check Seat Status
	if seatStatus == SeatStatusDormant {
		tips = append(tips, OptimizationTip{
			ID:                  "dormant-license",
			Title:               "Dormant Gemini Enterprise License",
			Description:         "No recent inference activity detected. Consider releasing or reallocating this license to save $45/mo.",
			Severity:            "warning",
			EstimatedSavingsUSD: 45.0,
		})
	}

	// 2. Check Model Tier Switching (Gemini Pro vs Flash)
	var proTokens int64
	var proCost float64
	for m, d := range modelBreakdown {
		if strings.Contains(strings.ToLower(m), "pro") {
			proTokens += d.Tokens
			proCost += d.Cost
		}
	}
	if totalTokens > 0 {
		proShare := float64(proTokens) / float64(totalTokens)
		if proShare >= 0.5 && proCost > 10.0 {
			savings := math.Round((proCost*0.65)*100) / 100
			tips = append(tips, OptimizationTip{
				ID:                  "model-tier-switch",
				Title:               "Route Routine Tasks to Gemini Flash",
				Description:         fmt.Sprintf("%.0f%% of your queries use Gemini Pro. Shifting simple code edits and test generation to Gemini 1.5 Flash could reduce your monthly spend significantly.", proShare*100),
				Severity:            "info",
				EstimatedSavingsUSD: savings,
			})
		}
	}

	// 3. Context Caching Recommendation
	if totalTokens > 500000 {
		tips = append(tips, OptimizationTip{
			ID:                  "context-caching",
			Title:               "Leverage Prompt & Context Caching",
			Description:         "High token throughput observed. Using Gemini Context Caching for large repeated repositories or docs can save up to 75% on prompt tokens.",
			Severity:            "info",
			EstimatedSavingsUSD: math.Round((float64(totalTokens)*0.000002)*100) / 100,
		})
	}

	if len(tips) == 0 {
		tips = append(tips, OptimizationTip{
			ID:                  "optimal-usage",
			Title:               "Healthy Consumption Profile",
			Description:         "Your usage is balanced and within expected budget bounds with good token efficiency.",
			Severity:            "success",
			EstimatedSavingsUSD: 0.0,
		})
	}

	return tips
}

