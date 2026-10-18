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
	CachedTokens     int64     `json:"cached_tokens"`
	CompletionTokens int64     `json:"completion_tokens"`
	ThoughtsTokens   int64     `json:"thoughts_tokens"`
	QuotaType        string    `json:"quota_type"`
	EffortLevel      string    `json:"effort_level"`
	ClientName       string    `json:"client_name"`
	ClientVersion    string    `json:"client_version"`
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
	UserID                string  `json:"user_id"`
	Model                 string  `json:"model"`
	UsageDate             string  `json:"usage_date"`
	UserTokens            int64   `json:"user_tokens"`
	CachedTokens          int64   `json:"cached_tokens"`
	CacheHitRate          float64 `json:"cache_hit_rate"`
	TotalModelTokens      int64   `json:"total_model_tokens"`
	TokenShare            float64 `json:"token_share"`
	AllocatedCost         float64 `json:"allocated_cost"`
	WTU                   int64   `json:"wtu"`
	IncludedCost          float64 `json:"included_cost"`
	PAYGCost              float64 `json:"payg_cost"`
	FairShareIncludedCost float64 `json:"fair_share_included_cost"`
	FairSharePAYGCost     float64 `json:"fair_share_payg_cost"`
	WhenUsedIncludedCost  float64 `json:"when_used_included_cost"`
	WhenUsedPAYGCost      float64 `json:"when_used_payg_cost"`
	Currency              string  `json:"currency"`
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
	UserID                string                     `json:"user_id"`
	TotalTokens           int64                      `json:"total_tokens"`
	TotalCost             float64                    `json:"total_cost"`
	WTU                   int64                      `json:"wtu"`
	IncludedCost          float64                    `json:"included_cost"`
	PAYGCost              float64                    `json:"payg_cost"`
	FairShareIncludedCost float64                    `json:"fair_share_included_cost"`
	FairSharePAYGCost     float64                    `json:"fair_share_payg_cost"`
	WhenUsedIncludedCost  float64                    `json:"when_used_included_cost"`
	WhenUsedPAYGCost      float64                    `json:"when_used_payg_cost"`
	CachedTokens          int64                      `json:"cached_tokens"`
	CacheHitRate          float64                    `json:"cache_hit_rate"`
	Currency              string                     `json:"currency"`
	ModelBreakdown        map[string]ModelCostDetail `json:"model_breakdown"`
	LastActive            time.Time                  `json:"last_active"`
	SeatStatus            SeatStatus                 `json:"seat_status"`
}

// ModelCostDetail holds metrics for a specific model under a user.
type ModelCostDetail struct {
	Tokens int64   `json:"tokens"`
	Cost   float64 `json:"cost"`
	Share  float64 `json:"share"`
}

// LeaderboardEntry represents a ranked developer's aggregated spend, token efficiency, and model usage.
type LeaderboardEntry struct {
	Rank                  int                `json:"rank"`
	UserID                string             `json:"user_id"`
	TotalCost             float64            `json:"total_cost"`
	ShareOfSpend          float64            `json:"share_of_spend"`
	WTU                   int64              `json:"wtu"`
	IncludedCost          float64            `json:"included_cost"`
	PAYGCost              float64            `json:"payg_cost"`
	FairShareIncludedCost float64            `json:"fair_share_included_cost"`
	FairSharePAYGCost     float64            `json:"fair_share_payg_cost"`
	WhenUsedIncludedCost  float64            `json:"when_used_included_cost"`
	WhenUsedPAYGCost      float64            `json:"when_used_payg_cost"`
	TotalTokens           int64              `json:"total_tokens"`
	CachedTokens          int64              `json:"cached_tokens"`
	CacheHitRate          float64            `json:"cache_hit_rate"`
	ShareOfTokens         float64            `json:"share_of_tokens"`
	ActiveDays            int                `json:"active_days"`
	CostPer1KTokens       float64            `json:"cost_per_1k_tokens"`
	TopModel              string             `json:"top_model"`
	TopEffort             string             `json:"top_effort"`
	ModelBreakdown        map[string]float64 `json:"model_breakdown"`
	Currency              string             `json:"currency"`
}

// ToCSVRow converts the leaderboard entry to a slice of strings for CSV output.
func (l LeaderboardEntry) ToCSVRow() []string {
	return []string{
		strconv.Itoa(l.Rank),
		l.UserID,
		fmt.Sprintf("%.4f", l.TotalCost),
		fmt.Sprintf("%.2f", l.ShareOfSpend),
		strconv.FormatInt(l.TotalTokens, 10),
		strconv.Itoa(l.ActiveDays),
		fmt.Sprintf("%.4f", l.CostPer1KTokens),
		l.TopModel,
		l.Currency,
	}
}

// TokenCostShare compares token volume share vs. dollar cost share across token categories.
type TokenCostShare struct {
	Category   string  `json:"category"` // "uncached_input", "cached_input", "output", "thinking"
	Label      string  `json:"label"`
	Tokens     int64   `json:"tokens"`
	TokenShare float64 `json:"token_share"` // percentage 0..100
	CostUSD    float64 `json:"cost_usd"`
	CostShare  float64 `json:"cost_share"` // percentage 0..100
}

// ModelEffortStats holds aggregated metrics grouped by model and thinking effort level.
type ModelEffortStats struct {
	Model          string  `json:"model"`
	EffortLevel    string  `json:"effort_level"`
	Calls          int     `json:"calls"`
	TotalTokens    int64   `json:"total_tokens"`
	PromptTokens   int64   `json:"prompt_tokens"`
	CachedTokens   int64   `json:"cached_tokens"`
	ThinkingTokens int64   `json:"thinking_tokens"`
	CacheHitRate   float64 `json:"cache_hit_rate"`
	WTU            int64   `json:"wtu"`
	IncludedCost   float64 `json:"included_cost"`
	PAYGCost       float64 `json:"payg_cost"`
	AllocatedCost  float64 `json:"allocated_cost"`
}

// ClientStats holds aggregated metrics grouped by client name and version.
type ClientStats struct {
	ClientName    string  `json:"client_name"`
	ClientVersion string  `json:"client_version"`
	Calls         int     `json:"calls"`
	TotalTokens   int64   `json:"total_tokens"`
	TokenShare    float64 `json:"token_share"`
	WTU           int64   `json:"wtu"`
	IncludedCost  float64 `json:"included_cost"`
	PAYGCost      float64 `json:"payg_cost"`
	AllocatedCost float64 `json:"allocated_cost"`
}

// LicenseSeat tracks the assignment, subscription tier, region, and activity for a developer's seat.
type LicenseSeat struct {
	UserID              string     `json:"user_id"`
	Assigned            bool       `json:"assigned"`
	Status              SeatStatus `json:"status"`
	Tier                string     `json:"tier,omitempty"` // "standard" or "plus"
	WeeklyAllowanceUSD  float64    `json:"weekly_allowance_usd,omitempty"`
	WeeklyAllowanceWTU  int64      `json:"weekly_allowance_wtu,omitempty"`
	Location            string     `json:"location,omitempty"` // "global", "eu", "us"
	IsTrial             bool       `json:"is_trial,omitempty"`
	TrialEndsAt         string     `json:"trial_ends_at,omitempty"`
	LastActivity        time.Time  `json:"last_activity"`
	TotalTokensInWindow int64      `json:"total_tokens_in_window"`
}

// WeeklyPoolWindow represents a single 7-day allowance window's limit vs. actual consumption.
type WeeklyPoolWindow struct {
	WindowLabel     string  `json:"window_label"`
	WindowStart     string  `json:"window_start"`
	WindowEnd       string  `json:"window_end"`
	PoolLimitUSD    float64 `json:"pool_limit_usd"`
	PoolLimitWTU    int64   `json:"pool_limit_wtu"`
	IncludedUsedUSD float64 `json:"included_used_usd"`
	IncludedUsedWTU int64   `json:"included_used_wtu"`
	PAYGOverageUSD  float64 `json:"payg_overage_usd"`
	PAYGOverageWTU  int64   `json:"payg_overage_wtu"`
	UtilizationPct  float64 `json:"utilization_pct"`
}

// PoolWhatIfScenario represents the current 7-day window metrics under a specific What-If configuration.
type PoolWhatIfScenario struct {
	PoolMode        string             `json:"pool_mode"`  // "pooled" or "per_user"
	SeatBasis       string             `json:"seat_basis"` // "assigned" or "purchased"
	PoolLimitUSD    float64            `json:"pool_limit_usd"`
	PoolLimitWTU    int64              `json:"pool_limit_wtu"`
	IncludedUsedUSD float64            `json:"included_used_usd"`
	IncludedUsedWTU int64              `json:"included_used_wtu"`
	PAYGOverageUSD  float64            `json:"payg_overage_usd"`
	PAYGOverageWTU  int64              `json:"payg_overage_wtu"`
	RemainingUSD    float64            `json:"remaining_usd"`
	RemainingWTU    int64              `json:"remaining_wtu"`
	UtilizationPct  float64            `json:"utilization_pct"`
	Windows         []WeeklyPoolWindow `json:"windows,omitempty"`
}

// WeeklyAllowancePool represents Gemini Enterprise's floating 7-day Weekly Allowance Pool and What-If simulations.
type WeeklyAllowancePool struct {
	WindowStart     string                        `json:"window_start"`
	WindowEnd       string                        `json:"window_end"`
	ResetInHours    int                           `json:"reset_in_hours"`
	ResetCountdown  string                        `json:"reset_countdown"`
	PoolMode        string                        `json:"pool_mode"`  // "pooled" or "per_user"
	SeatBasis       string                        `json:"seat_basis"` // "assigned" or "purchased"
	PoolLimitUSD    float64                       `json:"pool_limit_usd"`
	PoolLimitWTU    int64                         `json:"pool_limit_wtu"`
	IncludedUsedUSD float64                       `json:"included_used_usd"`
	IncludedUsedWTU int64                         `json:"included_used_wtu"`
	PAYGOverageUSD  float64                       `json:"payg_overage_usd"`
	PAYGOverageWTU  int64                         `json:"payg_overage_wtu"`
	RemainingUSD    float64                       `json:"remaining_usd"`
	RemainingWTU    int64                         `json:"remaining_wtu"`
	UtilizationPct  float64                       `json:"utilization_pct"`
	Windows         []WeeklyPoolWindow            `json:"windows,omitempty"`
	Scenarios       map[string]PoolWhatIfScenario `json:"scenarios,omitempty"`
}

// LicenseGovernanceSummary provides high-level seat quota, tier distribution, and weekly allowance pool metrics.
type LicenseGovernanceSummary struct {
	SeatQuota               int                  `json:"seat_quota"`
	AssignedSeats           int                  `json:"assigned_seats"`
	ActiveSeats             int                  `json:"active_seats"`
	DormantSeats            int                  `json:"dormant_seats"`
	StandardSeats           int                  `json:"standard_seats"`
	PlusSeats               int                  `json:"plus_seats"`
	TrialSeats              int                  `json:"trial_seats"`
	Locations               []string             `json:"locations,omitempty"`
	UtilizationPct          float64              `json:"utilization_pct"`
	EstimatedMonthlySavings float64              `json:"estimated_monthly_savings"`
	Currency                string               `json:"currency"`
	DormantUsers            []LicenseSeat        `json:"dormant_users"`
	AllSeats                []LicenseSeat        `json:"all_seats,omitempty"`
	WeeklyPool              *WeeklyAllowancePool `json:"weekly_pool,omitempty"`
}

// OverviewMetrics holds top-level KPIs for the web dashboard and CLI summary.
type OverviewMetrics struct {
	TotalBilledCost         float64                    `json:"total_billed_cost"`
	TotalWTU                int64                      `json:"total_wtu"`
	IncludedCost            float64                    `json:"included_cost"`
	PAYGCost                float64                    `json:"payg_cost"`
	TotalTokens             int64                      `json:"total_tokens"`
	TotalCachedTokens       int64                      `json:"total_cached_tokens"`
	CacheHitRate            float64                    `json:"cache_hit_rate"`
	ActiveUsersCount        int                        `json:"active_users_count"`
	SeatQuota               int                        `json:"seat_quota"`
	AssignedSeats           int                        `json:"assigned_seats"`
	UtilizationPct          float64                    `json:"utilization_pct"`
	EstimatedMonthlySavings float64                    `json:"estimated_monthly_savings"`
	DailyTrends             []DailySpendTrend          `json:"daily_trends"`
	ModelBreakdown          map[string]ModelCostDetail `json:"model_breakdown"`
	TokenCostShares         []TokenCostShare           `json:"token_cost_shares"`
	ModelEffortStats        []ModelEffortStats         `json:"model_effort_stats"`
	ClientBreakdown         []ClientStats              `json:"client_breakdown"`
	WeeklyPool              *WeeklyAllowancePool       `json:"weekly_pool,omitempty"`
	Currency                string                     `json:"currency"`
}

// DailySpendTrend tracks daily spend and token volume for trend charts.
type DailySpendTrend struct {
	Date         string             `json:"date"`
	TotalCost    float64            `json:"total_cost"`
	Cost         float64            `json:"cost"`
	IncludedCost float64            `json:"included_cost"`
	PAYGCost     float64            `json:"payg_cost"`
	WTU          int64              `json:"wtu"`
	TotalTokens  int64              `json:"total_tokens"`
	Tokens       int64              `json:"tokens"`
	ActiveUsers  int                `json:"active_users"`
	ByModel      map[string]float64 `json:"by_model,omitempty"`
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
	CachedTokens     int64     `json:"cached_tokens"`
	CompletionTokens int64     `json:"completion_tokens"`
	ThoughtsTokens   int64     `json:"thoughts_tokens"`
	EffortLevel      string    `json:"effort_level"`
	ClientName       string    `json:"client_name"`
	WTU              int64     `json:"wtu"`
	EstimatedCost    float64   `json:"estimated_cost"`
}

// UserConsumptionDriving represents personal driving metrics, budget tracking, and recommendations.
type UserConsumptionDriving struct {
	UserID                 string               `json:"user_id"`
	Currency               string               `json:"currency"`
	TotalSpendInWindow     float64              `json:"total_spend_in_window"`
	IncludedCost           float64              `json:"included_cost"`
	PAYGCost               float64              `json:"payg_cost"`
	TotalWTU               int64                `json:"total_wtu"`
	TotalTokensInWindow    int64                `json:"total_tokens_in_window"`
	CachedTokensInWindow   int64                `json:"cached_tokens_in_window"`
	CacheHitRate           float64              `json:"cache_hit_rate"`
	TopEffort              string               `json:"top_effort"`
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
	Date         string             `json:"date"`
	TotalCost    float64            `json:"total_cost"`
	IncludedCost float64            `json:"included_cost"`
	PAYGCost     float64            `json:"payg_cost"`
	WTU          int64              `json:"wtu"`
	TotalTokens  int64              `json:"total_tokens"`
	ByModel      map[string]float64 `json:"by_model"`
}

// OptimizationTip provides an actionable insight to reduce cost or improve token efficiency.
type OptimizationTip struct {
	ID                  string  `json:"id"`
	Title               string  `json:"title"`
	Description         string  `json:"description"`
	Severity            string  `json:"severity"` // "info", "warning", "success"
	EstimatedSavingsUSD float64 `json:"estimated_savings_usd"`
}

// ExtractEffortLevel extracts the thinking effort level suffix from an experience or model string.
// Supported levels: minimal, low, medium, high, xhigh, max. Defaults to "default".
func ExtractEffortLevel(s string) string {
	lower := strings.ToLower(strings.TrimSpace(s))
	if lower == "" {
		return "default"
	}
	validLevels := []string{"minimal", "low", "medium", "high", "xhigh", "max"}
	for _, lvl := range validLevels {
		if lower == lvl || strings.HasSuffix(lower, "-"+lvl) {
			return lvl
		}
	}
	return "default"
}

// modelRateSpec holds list pricing per 1M tokens (USD), WTU meter share, and whether the model is always PAYG.
type modelRateSpec struct {
	inputRate   float64
	cachedRate  float64
	outputRate  float64
	meterShare  float64
	alwaysPAYG  bool
}

func resolveModelRates(model string, promptTokens int64) modelRateSpec {
	lower := strings.ToLower(model)
	switch {
	case strings.Contains(lower, "claude") && strings.Contains(lower, "opus"):
		return modelRateSpec{inputRate: 5.00, cachedRate: 0.50, outputRate: 25.00, meterShare: 1.0, alwaysPAYG: true}
	case strings.Contains(lower, "claude"):
		return modelRateSpec{inputRate: 3.00, cachedRate: 0.30, outputRate: 15.00, meterShare: 1.0, alwaysPAYG: true}
	case strings.Contains(lower, "argon"):
		return modelRateSpec{inputRate: 4.00, cachedRate: 0.20, outputRate: 20.00, meterShare: 0.5, alwaysPAYG: false}
	case strings.Contains(lower, "flash"):
		return modelRateSpec{inputRate: 0.50, cachedRate: 0.05, outputRate: 3.00, meterShare: 0.5, alwaysPAYG: false}
	default: // Gemini Pro or fallback
		if promptTokens > 200000 {
			return modelRateSpec{inputRate: 4.00, cachedRate: 0.40, outputRate: 18.00, meterShare: 1.0, alwaysPAYG: false}
		}
		return modelRateSpec{inputRate: 2.00, cachedRate: 0.20, outputRate: 12.00, meterShare: 1.0, alwaysPAYG: false}
	}
}

// CalculateCallTokenCategoryCosts computes metered USD value per token category (uncached input, cached input, output, thinking).
func CalculateCallTokenCategoryCosts(model string, promptTokens, cachedTokens, completionTokens, thoughtsTokens int64) (uncachedInputCost, cachedInputCost, outputCost, thinkingCost float64) {
	if promptTokens < 0 {
		promptTokens = 0
	}
	if cachedTokens < 0 {
		cachedTokens = 0
	}
	if cachedTokens > promptTokens {
		cachedTokens = promptTokens
	}
	if completionTokens < 0 {
		completionTokens = 0
	}
	if thoughtsTokens < 0 {
		thoughtsTokens = 0
	}

	rates := resolveModelRates(model, promptTokens)
	uncachedPrompt := promptTokens - cachedTokens

	uncachedInputCost = (float64(uncachedPrompt) / 1_000_000.0) * rates.inputRate * rates.meterShare
	cachedInputCost = (float64(cachedTokens) / 1_000_000.0) * rates.cachedRate * rates.meterShare
	outputCost = (float64(completionTokens) / 1_000_000.0) * rates.outputRate * rates.meterShare
	thinkingCost = (float64(thoughtsTokens) / 1_000_000.0) * rates.outputRate * rates.meterShare

	return uncachedInputCost, cachedInputCost, outputCost, thinkingCost
}

// CalculateCallWTU computes Weighted Token Units (WTU), Included Seat Allowance Cost (USD), and PAYG Overage Cost (USD)
// for a single inference call following Gemini Enterprise's metered model (1M WTU = $2.00, minimum 50 WTU per call).
func CalculateCallWTU(model string, promptTokens, cachedTokens, completionTokens, thoughtsTokens int64, quotaType string, isLicensed bool) (wtu int64, includedCost, paygCost float64) {
	if strings.EqualFold(quotaType, "NOOP_QUOTA") {
		return 0, 0, 0
	}
	totalActivity := promptTokens + completionTokens + thoughtsTokens
	if totalActivity <= 0 {
		return 0, 0, 0
	}

	rates := resolveModelRates(model, promptTokens)
	uCost, cCost, oCost, tCost := CalculateCallTokenCategoryCosts(model, promptTokens, cachedTokens, completionTokens, thoughtsTokens)
	meteredUSD := uCost + cCost + oCost + tCost

	// 1 WTU = $0.000002 (1M WTU = $2.00), truncated to whole units with 50 WTU minimum
	rawWTU := int64(meteredUSD / 0.000002)
	if rawWTU < 50 {
		rawWTU = 50
	}
	wtu = rawWTU
	effectiveUSD := math.Round(float64(wtu)*0.000002*10000) / 10000

	isPAYG := rates.alwaysPAYG || !isLicensed || strings.EqualFold(quotaType, "OVERAGE_QUOTA")
	if isPAYG {
		return wtu, 0.0, effectiveUSD
	}
	return wtu, effectiveUSD, 0.0
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

	switch {
	case consumedPct >= 100.0:
		status = "exceeded"
	case consumedPct >= 75.0 || projectedMonthEnd > budget:
		status = "warning"
	default:
		status = "on_track"
	}

	return status, consumedPct, projectedMonthEnd
}

// GenerateOptimizationTips evaluates user telemetry to generate contextual cost-saving tips.
func GenerateOptimizationTips(userID string, modelBreakdown map[string]ModelCostDetail, totalTokens int64, seatStatus SeatStatus, currency ...string) []OptimizationTip {
	cur := "USD"
	if len(currency) > 0 && currency[0] != "" {
		cur = currency[0]
	}
	sym := CurrencySymbol(cur)

	tips := make([]OptimizationTip, 0)

	// 1. Check Seat Status
	if seatStatus == SeatStatusDormant {
		tips = append(tips, OptimizationTip{
			ID:                  "dormant-license",
			Title:               "Dormant Gemini Enterprise License",
			Description:         fmt.Sprintf("No recent inference activity detected. Consider releasing or reallocating this license to save %s45/mo.", sym),
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

// ResolveSeatAllowance normalizes a Gemini Enterprise subscription tier and returns its weekly pool allowance in USD and WTU.
// Standard: $2.50/week (1,250,000 WTU/week at $2.00 per 1M WTU)
// Plus:     $3.75/week (1,875,000 WTU/week at $2.00 per 1M WTU)
func ResolveSeatAllowance(rawTier string) (tier string, weeklyUSD float64, weeklyWTU int64) {
	t := strings.ToLower(strings.TrimSpace(rawTier))
	switch {
	case t == "standard" || strings.Contains(t, "standard") || strings.Contains(t, "enterprise"):
		return "standard", 2.50, 1250000
	default:
		return "plus", 3.75, 1875000
	}
}

// IsPoolEligibleModel returns true if the model can consume from the Weekly Allowance Pool (Gemini & Argon models),
// and false if the model is always 100% PAYG (e.g., Anthropic Claude models).
func IsPoolEligibleModel(model string) bool {
	return !resolveModelRates(model, 0).alwaysPAYG
}
