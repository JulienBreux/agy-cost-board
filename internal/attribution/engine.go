package attribution

import (
	"cmp"
	"context"
	"fmt"
	"math"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/julienbreux/agy-cost-board/internal/bigquery"
	"github.com/julienbreux/agy-cost-board/internal/domain"
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

// resolveCurrency extracts the ISO-4217 currency from billed costs, falling back to USD.
func resolveCurrency(billedCosts []domain.BilledCost) string {
	for _, b := range billedCosts {
		if b.Currency != "" {
			return b.Currency
		}
	}
	return "USD"
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

// isUserLicensedMap builds a lookup map of licensed user IDs from FetchLicenseSeats.
func (e *Engine) isUserLicensedMap(ctx context.Context, days int) map[string]bool {
	licensed := make(map[string]bool)
	seats, _, err := e.provider.FetchLicenseSeats(ctx, days)
	if err != nil {
		return licensed
	}
	for _, s := range seats {
		if s.Assigned {
			licensed[s.UserID] = true
		}
	}
	return licensed
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

	licensedUsers := e.isUserLicensedMap(ctx, days)
	currency := resolveCurrency(billedCosts)

	// Group billing by date
	dailyBilled := make(map[string]float64)
	for _, b := range billedCosts {
		dailyBilled[b.UsageDate] += b.NetCost
	}

	type userDayModelStats struct {
		tokens       int64
		promptTokens int64
		cachedTokens int64
		wtu          int64
		includedCost float64
		paygCost     float64
	}

	// Group tokens by Date -> Model -> User -> Stats
	type ModelUserTokens struct {
		users       map[string]*userDayModelStats
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
				users: make(map[string]*userDayModelStats),
			}
		}
		uStats, ok := dayModelMap[dateStr][l.Model].users[l.UserID]
		if !ok {
			uStats = &userDayModelStats{}
			dayModelMap[dateStr][l.Model].users[l.UserID] = uStats
		}

		isLicensed := true
		if len(licensedUsers) > 0 {
			isLicensed = licensedUsers[l.UserID]
		}
		wtu, inc, payg := domain.CalculateCallWTU(
			l.Model,
			l.PromptTokens,
			l.CachedTokens,
			l.CompletionTokens,
			l.ThoughtsTokens,
			l.QuotaType,
			isLicensed,
		)

		uStats.tokens += l.TotalTokens
		uStats.promptTokens += l.PromptTokens
		uStats.cachedTokens += l.CachedTokens
		uStats.wtu += wtu
		uStats.includedCost += inc
		uStats.paygCost += payg

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

			for userID, uStats := range mut.users {
				share, allocated := domain.CalculateProportionalCost(uStats.tokens, mut.totalTokens, modelCost)
				var cacheHitRate float64
				if uStats.promptTokens > 0 {
					cacheHitRate = math.Round((float64(uStats.cachedTokens)/float64(uStats.promptTokens))*1000) / 10.0
				}

				results = append(results, domain.AllocatedUserCost{
					UserID:           userID,
					Model:            model,
					UsageDate:        dateStr,
					UserTokens:       uStats.tokens,
					CachedTokens:     uStats.cachedTokens,
					CacheHitRate:     cacheHitRate,
					TotalModelTokens: mut.totalTokens,
					TokenShare:       share,
					AllocatedCost:    allocated,
					WTU:              uStats.wtu,
					IncludedCost:     math.Round(uStats.includedCost*10000) / 10000,
					PAYGCost:         math.Round(uStats.paygCost*10000) / 10000,
					Currency:         currency,
				})
			}
		}
	}

	// Sort results by date descending, then allocated cost descending
	slices.SortFunc(results, func(a, b domain.AllocatedUserCost) int {
		if a.UsageDate != b.UsageDate {
			return cmp.Compare(b.UsageDate, a.UsageDate)
		}
		return cmp.Compare(b.AllocatedCost, a.AllocatedCost)
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

	summaryCurrency := "USD"
	if len(allCosts) > 0 && allCosts[0].Currency != "" {
		summaryCurrency = allCosts[0].Currency
	} else if billedCosts, err := e.provider.FetchBilledCosts(ctx, days); err == nil {
		summaryCurrency = resolveCurrency(billedCosts)
	}

	summary := &domain.UserCostSummary{
		UserID:         userID,
		Currency:       summaryCurrency,
		ModelBreakdown: make(map[string]domain.ModelCostDetail),
	}

	var latestActivity time.Time
	var userPromptTokens int64
	for _, l := range logs {
		if l.UserID == userID {
			userPromptTokens += l.PromptTokens
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
			summary.WTU += c.WTU
			summary.IncludedCost += c.IncludedCost
			summary.PAYGCost += c.PAYGCost
			summary.CachedTokens += c.CachedTokens

			detail := summary.ModelBreakdown[c.Model]
			detail.Tokens += c.UserTokens
			detail.Cost += c.AllocatedCost
			summary.ModelBreakdown[c.Model] = detail
		}
	}
	summary.IncludedCost = math.Round(summary.IncludedCost*100) / 100
	summary.PAYGCost = math.Round(summary.PAYGCost*100) / 100
	if userPromptTokens > 0 {
		summary.CacheHitRate = math.Round((float64(summary.CachedTokens)/float64(userPromptTokens))*1000) / 10.0
	}

	// Compute model share percentages
	for m, d := range summary.ModelBreakdown {
		if summary.TotalTokens > 0 {
			d.Share = float64(d.Tokens) / float64(summary.TotalTokens)
		}
		summary.ModelBreakdown[m] = d
	}

	// If user has no inference activity in the window, check if they exist in license seats
	if latestActivity.IsZero() && summary.TotalTokens == 0 {
		seats, _, err := e.provider.FetchLicenseSeats(ctx, days)
		if err == nil {
			for _, s := range seats {
				if s.UserID == userID {
					summary.SeatStatus = s.Status
					summary.LastActive = s.LastActivity
					return summary, nil
				}
			}
		}
		return nil, domain.ErrUserNotFound
	}

	summary.SeatStatus = domain.ClassifySeatStatus(summary.TotalTokens, summary.LastActive, days)
	return summary, nil
}

// GetLeaderboard computes ranked developer entries ordered by total attributed spend descending.
func (e *Engine) GetLeaderboard(ctx context.Context, days int) ([]domain.LeaderboardEntry, error) {
	if days <= 0 {
		days = 30
	}
	cacheKey := fmt.Sprintf("leaderboard_%d", days)
	if val, ok := e.getFromCache(cacheKey); ok {
		return val.([]domain.LeaderboardEntry), nil
	}

	allCosts, err := e.GetAttributedCosts(ctx, days, "")
	if err != nil {
		return nil, err
	}

	logs, _ := e.provider.FetchTelemetryLogs(ctx, days)

	currency := "USD"
	if len(allCosts) > 0 && allCosts[0].Currency != "" {
		currency = allCosts[0].Currency
	} else if billedCosts, err := e.provider.FetchBilledCosts(ctx, days); err == nil {
		currency = resolveCurrency(billedCosts)
	}

	type userLeaderboardAgg struct {
		userID       string
		totalCost    float64
		wtu          int64
		includedCost float64
		paygCost     float64
		totalTokens  int64
		promptTokens int64
		cachedTokens int64
		activeDates  map[string]bool
		modelTokens  map[string]int64
		effortTokens map[string]int64
	}

	aggMap := make(map[string]*userLeaderboardAgg)
	var orgTotalCost float64
	var orgTotalTokens int64

	for _, c := range allCosts {
		u, ok := aggMap[c.UserID]
		if !ok {
			u = &userLeaderboardAgg{
				userID:       c.UserID,
				activeDates:  make(map[string]bool),
				modelTokens:  make(map[string]int64),
				effortTokens: make(map[string]int64),
			}
			aggMap[c.UserID] = u
		}
		u.totalCost += c.AllocatedCost
		u.wtu += c.WTU
		u.includedCost += c.IncludedCost
		u.paygCost += c.PAYGCost
		u.totalTokens += c.UserTokens
		u.cachedTokens += c.CachedTokens
		if c.UsageDate != "" {
			u.activeDates[c.UsageDate] = true
		}
		u.modelTokens[c.Model] += c.UserTokens

		orgTotalCost += c.AllocatedCost
		orgTotalTokens += c.UserTokens
	}

	for _, l := range logs {
		if u, ok := aggMap[l.UserID]; ok {
			u.promptTokens += l.PromptTokens
			eff := cmp.Or(l.EffortLevel, domain.ExtractEffortLevel(l.Model))
			u.effortTokens[eff] += l.TotalTokens
		}
	}

	entries := make([]domain.LeaderboardEntry, 0, len(aggMap))
	for _, u := range aggMap {
		var shareOfSpend float64
		if orgTotalCost > 0 {
			shareOfSpend = math.Round((u.totalCost/orgTotalCost)*10000) / 100
		}
		var shareOfTokens float64
		if orgTotalTokens > 0 {
			shareOfTokens = math.Round((float64(u.totalTokens)/float64(orgTotalTokens))*10000) / 100
		}
		var costPer1K float64
		if u.totalTokens > 0 {
			costPer1K = math.Round(((u.totalCost/float64(u.totalTokens))*1000)*10000) / 10000
		}
		var cacheHitRate float64
		if u.promptTokens > 0 {
			cacheHitRate = math.Round((float64(u.cachedTokens)/float64(u.promptTokens))*1000) / 10.0
		}

		topModel := ""
		var maxModelTokens int64 = -1
		modelBreakdown := make(map[string]float64, len(u.modelTokens))
		for m, mTokens := range u.modelTokens {
			if mTokens > maxModelTokens || (mTokens == maxModelTokens && (topModel == "" || m < topModel)) {
				maxModelTokens = mTokens
				topModel = m
			}
			if u.totalTokens > 0 {
				modelBreakdown[m] = math.Round((float64(mTokens)/float64(u.totalTokens))*1000) / 10
			}
		}

		topEffort := "default"
		var maxEffortTokens int64 = -1
		for eff, eTokens := range u.effortTokens {
			if eTokens > maxEffortTokens || (eTokens == maxEffortTokens && eff < topEffort) {
				maxEffortTokens = eTokens
				topEffort = eff
			}
		}

		entries = append(entries, domain.LeaderboardEntry{
			UserID:          u.userID,
			TotalCost:       math.Round(u.totalCost*100) / 100,
			ShareOfSpend:    shareOfSpend,
			WTU:             u.wtu,
			IncludedCost:    math.Round(u.includedCost*100) / 100,
			PAYGCost:        math.Round(u.paygCost*100) / 100,
			TotalTokens:     u.totalTokens,
			CachedTokens:    u.cachedTokens,
			CacheHitRate:    cacheHitRate,
			ShareOfTokens:   shareOfTokens,
			ActiveDays:      len(u.activeDates),
			CostPer1KTokens: costPer1K,
			TopModel:        topModel,
			TopEffort:       topEffort,
			ModelBreakdown:  modelBreakdown,
			Currency:        currency,
		})
	}

	slices.SortFunc(entries, func(a, b domain.LeaderboardEntry) int {
		if a.TotalCost != b.TotalCost {
			return cmp.Compare(b.TotalCost, a.TotalCost)
		}
		if a.TotalTokens != b.TotalTokens {
			return cmp.Compare(b.TotalTokens, a.TotalTokens)
		}
		return cmp.Compare(a.UserID, b.UserID)
	})

	for i := range entries {
		entries[i].Rank = i + 1
	}

	e.setToCache(cacheKey, entries)
	return entries, nil
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

	govCurrency := "USD"
	if billedCosts, err := e.provider.FetchBilledCosts(ctx, windowDays); err == nil {
		govCurrency = resolveCurrency(billedCosts)
	}

	summary := &domain.LicenseGovernanceSummary{
		SeatQuota:     quota,
		AssignedSeats: len(seats),
		Currency:      govCurrency,
		DormantUsers:  make([]domain.LicenseSeat, 0),
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

	logs, err := e.provider.FetchTelemetryLogs(ctx, days)
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

	licensedUsers := e.isUserLicensedMap(ctx, days)

	overview := &domain.OverviewMetrics{
		SeatQuota:               gov.SeatQuota,
		AssignedSeats:           gov.AssignedSeats,
		UtilizationPct:          gov.UtilizationPct,
		EstimatedMonthlySavings: gov.EstimatedMonthlySavings,
		Currency:                resolveCurrency(billedCosts),
		ModelBreakdown:          make(map[string]domain.ModelCostDetail),
		DailyTrends:             make([]domain.DailySpendTrend, 0),
		TokenCostShares:         make([]domain.TokenCostShare, 0, 4),
		ModelEffortStats:        make([]domain.ModelEffortStats, 0),
		ClientBreakdown:         make([]domain.ClientStats, 0),
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
		overview.TotalCachedTokens += c.CachedTokens
		overview.TotalWTU += c.WTU
		overview.IncludedCost += c.IncludedCost
		overview.PAYGCost += c.PAYGCost
		activeUsers[c.UserID] = true

		if _, ok := dailyAgg[c.UsageDate]; !ok {
			dailyAgg[c.UsageDate] = &domain.DailySpendTrend{
				Date:    c.UsageDate,
				ByModel: make(map[string]float64),
			}
			dailyUsers[c.UsageDate] = make(map[string]bool)
		}
		dailyAgg[c.UsageDate].TotalCost += c.AllocatedCost
		dailyAgg[c.UsageDate].IncludedCost += c.IncludedCost
		dailyAgg[c.UsageDate].PAYGCost += c.PAYGCost
		dailyAgg[c.UsageDate].WTU += c.WTU
		dailyAgg[c.UsageDate].TotalTokens += c.UserTokens
		dailyAgg[c.UsageDate].ByModel[c.Model] += c.AllocatedCost
		dailyUsers[c.UsageDate][c.UserID] = true

		d := overview.ModelBreakdown[c.Model]
		d.Tokens += c.UserTokens
		d.Cost += c.AllocatedCost
		overview.ModelBreakdown[c.Model] = d
	}

	overview.IncludedCost = math.Round(overview.IncludedCost*100) / 100
	overview.PAYGCost = math.Round(overview.PAYGCost*100) / 100
	overview.ActiveUsersCount = len(activeUsers)

	// Format daily trends
	trendList := make([]domain.DailySpendTrend, 0, len(dailyAgg))
	for d, tr := range dailyAgg {
		tr.ActiveUsers = len(dailyUsers[d])
		tr.TotalCost = math.Round(tr.TotalCost*100) / 100
		tr.Cost = tr.TotalCost
		tr.IncludedCost = math.Round(tr.IncludedCost*100) / 100
		tr.PAYGCost = math.Round(tr.PAYGCost*100) / 100
		tr.Tokens = tr.TotalTokens
		for m, cost := range tr.ByModel {
			tr.ByModel[m] = math.Round(cost*100) / 100
		}
		trendList = append(trendList, *tr)
	}

	slices.SortFunc(trendList, func(a, b domain.DailySpendTrend) int {
		return cmp.Compare(a.Date, b.Date)
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

	// Aggregate TokenCostShares, ModelEffortStats, and ClientBreakdown from telemetry logs
	var totalPromptTokens, uncachedInputTokens, cachedInputTokens, outputTokens, thinkingTokens int64
	var uncachedInputUSD, cachedInputUSD, outputUSD, thinkingUSD float64

	type effortKey struct {
		model  string
		effort string
	}
	effortMap := make(map[effortKey]*domain.ModelEffortStats)

	type clientKey struct {
		name    string
		version string
	}
	clientMap := make(map[clientKey]*domain.ClientStats)

	for _, l := range logs {
		pTok := max(l.PromptTokens, 0)
		cTok := min(max(l.CachedTokens, 0), pTok)
		uTok := pTok - cTok
		oTok := max(l.CompletionTokens, 0)
		tTok := max(l.ThoughtsTokens, 0)

		totalPromptTokens += pTok
		uncachedInputTokens += uTok
		cachedInputTokens += cTok
		outputTokens += oTok
		thinkingTokens += tTok

		uCost, cCost, oCost, tCost := domain.CalculateCallTokenCategoryCosts(l.Model, pTok, cTok, oTok, tTok)
		uncachedInputUSD += uCost
		cachedInputUSD += cCost
		outputUSD += oCost
		thinkingUSD += tCost

		isLicensed := true
		if len(licensedUsers) > 0 {
			isLicensed = licensedUsers[l.UserID]
		}
		callWTU, callInc, callPayg := domain.CalculateCallWTU(l.Model, pTok, cTok, oTok, tTok, l.QuotaType, isLicensed)

		eff := cmp.Or(l.EffortLevel, domain.ExtractEffortLevel(l.Model))
		ek := effortKey{model: l.Model, effort: eff}
		es, ok := effortMap[ek]
		if !ok {
			es = &domain.ModelEffortStats{
				Model:       l.Model,
				EffortLevel: eff,
			}
			effortMap[ek] = es
		}
		es.Calls++
		es.TotalTokens += l.TotalTokens
		es.PromptTokens += pTok
		es.CachedTokens += cTok
		es.ThinkingTokens += tTok
		es.WTU += callWTU
		es.IncludedCost += callInc
		es.PAYGCost += callPayg

		cName := cmp.Or(l.ClientName, "antigravity")
		ck := clientKey{name: cName, version: l.ClientVersion}
		cs, ok := clientMap[ck]
		if !ok {
			cs = &domain.ClientStats{
				ClientName:    cName,
				ClientVersion: l.ClientVersion,
			}
			clientMap[ck] = cs
		}
		cs.Calls++
		cs.TotalTokens += l.TotalTokens
		cs.WTU += callWTU
		cs.IncludedCost += callInc
		cs.PAYGCost += callPayg
	}

	if totalPromptTokens > 0 {
		overview.CacheHitRate = math.Round((float64(cachedInputTokens)/float64(totalPromptTokens))*1000) / 10.0
	}

	totalCatTokens := uncachedInputTokens + cachedInputTokens + outputTokens + thinkingTokens
	totalCatUSD := uncachedInputUSD + cachedInputUSD + outputUSD + thinkingUSD

	makeShare := func(cat, label string, tok int64, usd float64) domain.TokenCostShare {
		var tShare, cShare float64
		if totalCatTokens > 0 {
			tShare = math.Round((float64(tok)/float64(totalCatTokens))*1000) / 10.0
		}
		if totalCatUSD > 0 {
			cShare = math.Round((usd/totalCatUSD)*1000) / 10.0
		}
		return domain.TokenCostShare{
			Category:   cat,
			Label:      label,
			Tokens:     tok,
			TokenShare: tShare,
			CostUSD:    math.Round(usd*100) / 100,
			CostShare:  cShare,
		}
	}

	overview.TokenCostShares = []domain.TokenCostShare{
		makeShare("uncached_input", "Uncached Input", uncachedInputTokens, uncachedInputUSD),
		makeShare("cached_input", "Cached Input", cachedInputTokens, cachedInputUSD),
		makeShare("output", "Output", outputTokens, outputUSD),
		makeShare("thinking", "Thinking", thinkingTokens, thinkingUSD),
	}

	for _, es := range effortMap {
		if es.PromptTokens > 0 {
			es.CacheHitRate = math.Round((float64(es.CachedTokens)/float64(es.PromptTokens))*1000) / 10.0
		}
		es.IncludedCost = math.Round(es.IncludedCost*100) / 100
		es.PAYGCost = math.Round(es.PAYGCost*100) / 100
		if mDetail, ok := overview.ModelBreakdown[es.Model]; ok && mDetail.Tokens > 0 {
			es.AllocatedCost = math.Round((float64(es.TotalTokens)/float64(mDetail.Tokens))*mDetail.Cost*100) / 100
		}
		overview.ModelEffortStats = append(overview.ModelEffortStats, *es)
	}
	slices.SortFunc(overview.ModelEffortStats, func(a, b domain.ModelEffortStats) int {
		if a.WTU != b.WTU {
			return cmp.Compare(b.WTU, a.WTU)
		}
		return cmp.Compare(b.TotalTokens, a.TotalTokens)
	})

	for _, cs := range clientMap {
		if overview.TotalTokens > 0 {
			cs.TokenShare = math.Round((float64(cs.TotalTokens)/float64(overview.TotalTokens))*1000) / 10.0
			cs.AllocatedCost = math.Round((float64(cs.TotalTokens)/float64(overview.TotalTokens))*overview.TotalBilledCost*100) / 100
		}
		cs.IncludedCost = math.Round(cs.IncludedCost*100) / 100
		cs.PAYGCost = math.Round(cs.PAYGCost*100) / 100
		overview.ClientBreakdown = append(overview.ClientBreakdown, *cs)
	}
	slices.SortFunc(overview.ClientBreakdown, func(a, b domain.ClientStats) int {
		return cmp.Compare(b.TotalTokens, a.TotalTokens)
	})

	e.setToCache(cacheKey, overview)
	return overview, nil
}

// GetUserActivity returns recent inference telemetry logs and estimated cost for a given user.
func (e *Engine) GetUserActivity(ctx context.Context, userID string, days int, limit int) ([]domain.UserActivityLog, error) {
	if limit <= 0 {
		limit = 50
	}
	logs, err := e.provider.FetchTelemetryLogs(ctx, days)
	if err != nil {
		return nil, fmt.Errorf("fetch telemetry logs: %w", err)
	}

	allCosts, _ := e.GetAttributedCosts(ctx, days, "")
	licensedUsers := e.isUserLicensedMap(ctx, days)

	// Calculate model average cost per token
	type modelCostAgg struct {
		totalCost   float64
		totalTokens int64
	}
	modelUnitCost := make(map[string]float64)
	agg := make(map[string]*modelCostAgg)
	for _, c := range allCosts {
		if _, ok := agg[c.Model]; !ok {
			agg[c.Model] = &modelCostAgg{}
		}
		agg[c.Model].totalCost += c.AllocatedCost
		agg[c.Model].totalTokens += c.UserTokens
	}
	for m, a := range agg {
		if a.totalTokens > 0 {
			modelUnitCost[m] = a.totalCost / float64(a.totalTokens)
		}
	}

	userLogs := make([]domain.TelemetryLog, 0)
	for _, l := range logs {
		if l.UserID == userID {
			userLogs = append(userLogs, l)
		}
	}

	// Sort descending by timestamp (newest first)
	slices.SortFunc(userLogs, func(a, b domain.TelemetryLog) int {
		if a.Timestamp.Equal(b.Timestamp) {
			return 0
		}
		if a.Timestamp.After(b.Timestamp) {
			return -1
		}
		return 1
	})

	if len(userLogs) > limit {
		userLogs = userLogs[:limit]
	}

	isLicensed := true
	if len(licensedUsers) > 0 {
		isLicensed = licensedUsers[userID]
	}

	result := make([]domain.UserActivityLog, 0, len(userLogs))
	for _, l := range userLogs {
		rate := modelUnitCost[l.Model]
		cost := float64(l.TotalTokens) * rate
		cost = math.Round(cost*10000) / 10000
		wtu, _, _ := domain.CalculateCallWTU(l.Model, l.PromptTokens, l.CachedTokens, l.CompletionTokens, l.ThoughtsTokens, l.QuotaType, isLicensed)

		result = append(result, domain.UserActivityLog{
			Timestamp:        l.Timestamp,
			UserID:           l.UserID,
			Model:            l.Model,
			TotalTokens:      l.TotalTokens,
			PromptTokens:     l.PromptTokens,
			CachedTokens:     l.CachedTokens,
			CompletionTokens: l.CompletionTokens,
			ThoughtsTokens:   l.ThoughtsTokens,
			EffortLevel:      cmp.Or(l.EffortLevel, domain.ExtractEffortLevel(l.Model)),
			ClientName:       cmp.Or(l.ClientName, "antigravity"),
			WTU:              wtu,
			EstimatedCost:    cost,
		})
	}

	return result, nil
}

// GetUserConsumptionDriving computes consumption driving analytics, velocity, budget metrics, and recommendations.
func (e *Engine) GetUserConsumptionDriving(ctx context.Context, userID string, days int, monthlyBudget float64) (*domain.UserConsumptionDriving, error) {
	if days <= 0 {
		days = 30
	}
	if monthlyBudget <= 0 {
		monthlyBudget = 150.0
	}

	summary, err := e.GetUserSummary(ctx, userID, days)
	if err != nil {
		return nil, err
	}

	overview, err := e.GetOverviewMetrics(ctx, days)
	if err != nil {
		return nil, err
	}

	allCosts, err := e.GetAttributedCosts(ctx, days, "")
	if err != nil {
		return nil, err
	}

	logs, _ := e.provider.FetchTelemetryLogs(ctx, days)
	effortTokens := make(map[string]int64)
	for _, l := range logs {
		if l.UserID == userID {
			eff := cmp.Or(l.EffortLevel, domain.ExtractEffortLevel(l.Model))
			effortTokens[eff] += l.TotalTokens
		}
	}
	topEffort := "default"
	var maxEffortTok int64 = -1
	for eff, tok := range effortTokens {
		if tok > maxEffortTok || (tok == maxEffortTok && eff < topEffort) {
			maxEffortTok = tok
			topEffort = eff
		}
	}

	userDaily := make(map[string]*domain.PersonalDailyTrend)
	for _, c := range allCosts {
		if c.UserID == userID {
			if _, ok := userDaily[c.UsageDate]; !ok {
				userDaily[c.UsageDate] = &domain.PersonalDailyTrend{
					Date:    c.UsageDate,
					ByModel: make(map[string]float64),
				}
			}
			tr := userDaily[c.UsageDate]
			tr.TotalCost += c.AllocatedCost
			tr.IncludedCost += c.IncludedCost
			tr.PAYGCost += c.PAYGCost
			tr.WTU += c.WTU
			tr.TotalTokens += c.UserTokens
			tr.ByModel[c.Model] += c.AllocatedCost
		}
	}

	trends := make([]domain.PersonalDailyTrend, 0, len(userDaily))
	for _, tr := range userDaily {
		tr.TotalCost = math.Round(tr.TotalCost*100) / 100
		tr.IncludedCost = math.Round(tr.IncludedCost*100) / 100
		tr.PAYGCost = math.Round(tr.PAYGCost*100) / 100
		trends = append(trends, *tr)
	}
	slices.SortFunc(trends, func(a, b domain.PersonalDailyTrend) int {
		return cmp.Compare(a.Date, b.Date)
	})

	dailyBurnRate := math.Round((summary.TotalCost/float64(days))*100) / 100
	weeklyBurnRate := math.Round((dailyBurnRate*7.0)*100) / 100

	var orgShare float64
	if overview.TotalBilledCost > 0 {
		orgShare = math.Round((summary.TotalCost/overview.TotalBilledCost)*1000) / 10.0
	}

	budgetStatus, consumedPct, projectedMonthEnd := domain.CalculateBudgetMetrics(summary.TotalCost, monthlyBudget, days)
	recommendations := domain.GenerateOptimizationTips(userID, summary.ModelBreakdown, summary.TotalTokens, summary.SeatStatus, summary.Currency)

	if topEffort == "max" || topEffort == "xhigh" || topEffort == "high" {
		recommendations = append(recommendations, domain.OptimizationTip{
			ID:                  "thinking-effort-tuning",
			Title:               "Right-Size Thinking Effort Level",
			Description:         fmt.Sprintf("Your dominant thinking effort is '%s'. Lowering routine coding turns to 'medium' or 'low' reduces thinking token output and PAYG overage.", topEffort),
			Severity:            "info",
			EstimatedSavingsUSD: math.Round(summary.PAYGCost*0.25*100) / 100,
		})
	}

	return &domain.UserConsumptionDriving{
		UserID:                 userID,
		Currency:               summary.Currency,
		TotalSpendInWindow:     math.Round(summary.TotalCost*100) / 100,
		IncludedCost:           summary.IncludedCost,
		PAYGCost:               summary.PAYGCost,
		TotalWTU:               summary.WTU,
		TotalTokensInWindow:    summary.TotalTokens,
		CachedTokensInWindow:   summary.CachedTokens,
		CacheHitRate:           summary.CacheHitRate,
		TopEffort:              topEffort,
		DailyBurnRate:          dailyBurnRate,
		WeeklyBurnRate:         weeklyBurnRate,
		OrgSpendSharePct:       orgShare,
		MonthlyBudget:          monthlyBudget,
		ProjectedMonthEndSpend: projectedMonthEnd,
		BudgetConsumedPct:      consumedPct,
		BudgetStatus:           budgetStatus,
		Recommendations:        recommendations,
		DailyTrends:            trends,
	}, nil
}


