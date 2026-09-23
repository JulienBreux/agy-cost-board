package attribution

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/julienbreux/agy-ge-board/internal/bigquery"
	"github.com/julienbreux/agy-ge-board/internal/domain"
)

// Standard license seat cost per month in USD used for savings estimates.
const DefaultMonthlySeatCostUSD = 45.0

// cacheEntry holds cached data with an expiration time.
type cacheEntry struct {
	value     any
	expiresAt time.Time
}

// Engine calculates proportional cost attributions and license governance metrics.
type Engine struct {
	provider bigquery.DataProvider
	ttl      time.Duration
	cacheMu  sync.RWMutex
	cache    map[string]cacheEntry
}

// NewEngine constructs a new attribution Engine with TTL caching.
func NewEngine(provider bigquery.DataProvider, ttl time.Duration) *Engine {
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	return &Engine{
		provider: provider,
		ttl:      ttl,
		cache:    make(map[string]cacheEntry),
	}
}

// getFromCache retrieves an item from cache if not expired.
func (e *Engine) getFromCache(key string) (any, bool) {
	e.cacheMu.RLock()
	defer e.cacheMu.RUnlock()

	entry, ok := e.cache[key]
	if !ok || time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.value, true
}

// setToCache stores an item in cache with TTL.
func (e *Engine) setToCache(key string, val any) {
	e.cacheMu.Lock()
	defer e.cacheMu.Unlock()

	e.cache[key] = cacheEntry{
		value:     val,
		expiresAt: time.Now().Add(e.ttl),
	}
}

// GetAttributedCosts computes proportional costs per user, date, and model.
func (e *Engine) GetAttributedCosts(ctx context.Context, days int, modelFilter string) ([]domain.AllocatedUserCost, error) {
	cacheKey := fmt.Sprintf("costs_%d_%s", days, modelFilter)
	if val, ok := e.getFromCache(cacheKey); ok {
		return val.([]domain.AllocatedUserCost), nil
	}

	logs, err := e.provider.FetchTelemetryLogs(ctx, days)
	if err != nil {
		return nil, fmt.Errorf("fetch telemetry: %w", err)
	}

	billedCosts, err := e.provider.FetchBilledCosts(ctx, days)
	if err != nil {
		return nil, fmt.Errorf("fetch billed costs: %w", err)
	}

	// Group billing by date
	dailyBilled := make(map[string]float64)
	for _, b := range billedCosts {
		dailyBilled[b.UsageDate] += b.NetCost
	}

	// Group tokens by Date -> Model -> User -> Tokens
	type ModelUserTokens struct {
		userTokens map[string]int64
		totalTokens int64
	}
	dayModelMap := make(map[string]map[string]*ModelUserTokens)
	dayTotalTokens := make(map[string]int64)

	for _, l := range logs {
		dateStr := l.Timestamp.UTC().Format("2006-01-02")
		if _, ok := dayModelMap[dateStr]; !ok {
			dayModelMap[dateStr] = make(map[string]*ModelUserTokens)
		}
		if _, ok := dayModelMap[dateStr][l.Model]; !ok {
			dayModelMap[dateStr][l.Model] = &ModelUserTokens{
				userTokens: make(map[string]int64),
			}
		}

		dayModelMap[dateStr][l.Model].userTokens[l.UserID] += l.TotalTokens
		dayModelMap[dateStr][l.Model].totalTokens += l.TotalTokens
		dayTotalTokens[dateStr] += l.TotalTokens
	}

	var results []domain.AllocatedUserCost

	for dateStr, modelMap := range dayModelMap {
		dayCost := dailyBilled[dateStr]
		totalDayTokens := dayTotalTokens[dateStr]

		for model, mut := range modelMap {
			if modelFilter != "" && !strings.EqualFold(model, modelFilter) {
				continue
			}

			// Proportional cost allocated to this model on this day
			var modelCost float64
			if totalDayTokens > 0 {
				modelCost = (float64(mut.totalTokens) / float64(totalDayTokens)) * dayCost
			}

			for userID, uTokens := range mut.userTokens {
				share, allocated := domain.CalculateProportionalCost(uTokens, mut.totalTokens, modelCost)

				results = append(results, domain.AllocatedUserCost{
					UserID:           userID,
					Model:            model,
					UsageDate:        dateStr,
					UserTokens:       uTokens,
					TotalModelTokens: mut.totalTokens,
					TokenShare:       share,
					AllocatedCost:    allocated,
					Currency:         "USD",
				})
			}
		}
	}

	// Sort results by date descending, then allocated cost descending
	sort.Slice(results, func(i, j int) bool {
		if results[i].UsageDate != results[j].UsageDate {
			return results[i].UsageDate > results[j].UsageDate
		}
		return results[i].AllocatedCost > results[j].AllocatedCost
	})

	e.setToCache(cacheKey, results)
	return results, nil
}

// GetUserSummary aggregates usage, cost, and seat status for a specific developer.
func (e *Engine) GetUserSummary(ctx context.Context, userID string, days int) (*domain.UserCostSummary, error) {
	allCosts, err := e.GetAttributedCosts(ctx, days, "")
	if err != nil {
		return nil, err
	}

	logs, err := e.provider.FetchTelemetryLogs(ctx, days)
	if err != nil {
		return nil, err
	}

	summary := &domain.UserCostSummary{
		UserID:         userID,
		Currency:       "USD",
		ModelBreakdown: make(map[string]domain.ModelCostDetail),
	}

	var latestActivity time.Time
	for _, l := range logs {
		if l.UserID == userID {
			if l.Timestamp.After(latestActivity) {
				latestActivity = l.Timestamp
			}
		}
	}
	summary.LastActive = latestActivity

	for _, c := range allCosts {
		if c.UserID == userID {
			summary.TotalTokens += c.UserTokens
			summary.TotalCost += c.AllocatedCost

			detail := summary.ModelBreakdown[c.Model]
			detail.Tokens += c.UserTokens
			detail.Cost += c.AllocatedCost
			summary.ModelBreakdown[c.Model] = detail
		}
	}

	// Compute model share percentages
	for m, d := range summary.ModelBreakdown {
		if summary.TotalTokens > 0 {
			d.Share = float64(d.Tokens) / float64(summary.TotalTokens)
		}
		summary.ModelBreakdown[m] = d
	}

	summary.SeatStatus = domain.ClassifySeatStatus(summary.TotalTokens, summary.LastActive, days)
	return summary, nil
}

// GetLicenseGovernance computes seat quotas, active vs dormant licenses, and potential savings.
func (e *Engine) GetLicenseGovernance(ctx context.Context, windowDays int) (*domain.LicenseGovernanceSummary, error) {
	cacheKey := fmt.Sprintf("governance_%d", windowDays)
	if val, ok := e.getFromCache(cacheKey); ok {
		return val.(*domain.LicenseGovernanceSummary), nil
	}

	seats, quota, err := e.provider.FetchLicenseSeats(ctx, windowDays)
	if err != nil {
		return nil, fmt.Errorf("fetch seats: %w", err)
	}

	summary := &domain.LicenseGovernanceSummary{
		SeatQuota:     quota,
		AssignedSeats: len(seats),
	}

	for _, s := range seats {
		switch s.Status {
		case domain.SeatStatusActive:
			summary.ActiveSeats++
		case domain.SeatStatusAtRisk:
			summary.ActiveSeats++ // Still counts towards active, but at risk
		case domain.SeatStatusDormant:
			summary.DormantSeats++
			summary.DormantUsers = append(summary.DormantUsers, s)
		}
	}

	if quota > 0 {
		summary.UtilizationPct = math.Round((float64(summary.ActiveSeats)/float64(quota))*1000) / 10.0
	}
	summary.EstimatedMonthlySavings = float64(summary.DormantSeats) * DefaultMonthlySeatCostUSD

	e.setToCache(cacheKey, summary)
	return summary, nil
}

// GetOverviewMetrics compiles top KPIs, daily trends, and model distributions for dashboards.
func (e *Engine) GetOverviewMetrics(ctx context.Context, days int) (*domain.OverviewMetrics, error) {
	cacheKey := fmt.Sprintf("overview_%d", days)
	if val, ok := e.getFromCache(cacheKey); ok {
		return val.(*domain.OverviewMetrics), nil
	}

	allCosts, err := e.GetAttributedCosts(ctx, days, "")
	if err != nil {
		return nil, err
	}

	billedCosts, err := e.provider.FetchBilledCosts(ctx, days)
	if err != nil {
		return nil, err
	}

	gov, err := e.GetLicenseGovernance(ctx, days)
	if err != nil {
		return nil, err
	}

	overview := &domain.OverviewMetrics{
		SeatQuota:               gov.SeatQuota,
		AssignedSeats:           gov.AssignedSeats,
		UtilizationPct:          gov.UtilizationPct,
		EstimatedMonthlySavings: gov.EstimatedMonthlySavings,
		Currency:                "USD",
		ModelBreakdown:          make(map[string]domain.ModelCostDetail),
	}

	for _, b := range billedCosts {
		overview.TotalBilledCost += b.NetCost
	}
	overview.TotalBilledCost = math.Round(overview.TotalBilledCost*100) / 100

	activeUsers := make(map[string]bool)
	dailyAgg := make(map[string]*domain.DailySpendTrend)
	dailyUsers := make(map[string]map[string]bool)

	for _, c := range allCosts {
		overview.TotalTokens += c.UserTokens
		activeUsers[c.UserID] = true

		if _, ok := dailyAgg[c.UsageDate]; !ok {
			dailyAgg[c.UsageDate] = &domain.DailySpendTrend{
				Date: c.UsageDate,
			}
			dailyUsers[c.UsageDate] = make(map[string]bool)
		}
		dailyAgg[c.UsageDate].TotalCost += c.AllocatedCost
		dailyAgg[c.UsageDate].TotalTokens += c.UserTokens
		dailyUsers[c.UsageDate][c.UserID] = true

		d := overview.ModelBreakdown[c.Model]
		d.Tokens += c.UserTokens
		d.Cost += c.AllocatedCost
		overview.ModelBreakdown[c.Model] = d
	}

	overview.ActiveUsersCount = len(activeUsers)

	// Format daily trends
	var trendList []domain.DailySpendTrend
	for d, tr := range dailyAgg {
		tr.ActiveUsers = len(dailyUsers[d])
		tr.TotalCost = math.Round(tr.TotalCost*100) / 100
		trendList = append(trendList, *tr)
	}

	sort.Slice(trendList, func(i, j int) bool {
		return trendList[i].Date < trendList[j].Date
	})
	overview.DailyTrends = trendList

	// Model shares
	for m, d := range overview.ModelBreakdown {
		if overview.TotalTokens > 0 {
			d.Share = float64(d.Tokens) / float64(overview.TotalTokens)
		}
		d.Cost = math.Round(d.Cost*100) / 100
		overview.ModelBreakdown[m] = d
	}

	e.setToCache(cacheKey, overview)
	return overview, nil
}
