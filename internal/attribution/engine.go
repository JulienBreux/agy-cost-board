package attribution

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"math"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/julienbreux/agy-cost-board/internal/bigquery"
	"github.com/julienbreux/agy-cost-board/internal/domain"
	"github.com/julienbreux/agy-cost-board/internal/observability"
)

// Standard license seat cost per month in USD used for savings estimates.
const DefaultMonthlySeatCostUSD = 45.0

// cacheEntry holds cached data with an expiration time.
type cacheEntry struct {
	value     any
	expiresAt time.Time
}

type sharedCache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
}

// Engine calculates proportional cost attributions and license governance metrics.
type Engine struct {
	provider           bigquery.DataProvider
	baseProvider       bigquery.DataProvider
	ttl                time.Duration
	shared             *sharedCache
	configuredProjects []string
	selectedProject    string
}

// NewEngine constructs a new attribution Engine with TTL caching.
func NewEngine(provider bigquery.DataProvider, ttl time.Duration) *Engine {
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	sc := &sharedCache{
		entries: make(map[string]cacheEntry),
	}
	var wrapped bigquery.DataProvider
	if provider != nil {
		wrapped = &cachedDataProvider{
			base:   provider,
			ttl:    ttl,
			shared: sc,
		}
	}
	return &Engine{
		provider:        wrapped,
		baseProvider:    wrapped,
		ttl:             ttl,
		shared:          sc,
		selectedProject: "all",
	}
}

type cachedSeatsResult struct {
	seats []domain.LicenseSeat
	quota int
}

// cachedDataProvider wraps a raw DataProvider with TTL caching so multi-step attribution
// workflows (e.g. GetOverviewMetrics) only query BigQuery once per time window.
type cachedDataProvider struct {
	base   bigquery.DataProvider
	ttl    time.Duration
	shared *sharedCache
}

func (c *cachedDataProvider) get(key string) (any, bool) {
	if c.shared == nil {
		return nil, false
	}
	c.shared.mu.RLock()
	defer c.shared.mu.RUnlock()
	entry, ok := c.shared.entries[key]
	if !ok || time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.value, true
}

func (c *cachedDataProvider) set(key string, val any) {
	if c.shared == nil {
		return
	}
	c.shared.mu.Lock()
	defer c.shared.mu.Unlock()
	c.shared.entries[key] = cacheEntry{
		value:     val,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *cachedDataProvider) FetchTelemetryLogs(ctx context.Context, days int) ([]domain.TelemetryLog, error) {
	key := fmt.Sprintf("raw:telemetry:%d", days)
	if val, ok := c.get(key); ok {
		return val.([]domain.TelemetryLog), nil
	}
	logs, err := c.base.FetchTelemetryLogs(ctx, days)
	if err != nil {
		return nil, err
	}
	c.set(key, logs)
	return logs, nil
}

func (c *cachedDataProvider) FetchBilledCosts(ctx context.Context, days int) ([]domain.BilledCost, error) {
	key := fmt.Sprintf("raw:billing:%d", days)
	if val, ok := c.get(key); ok {
		return val.([]domain.BilledCost), nil
	}
	costs, err := c.base.FetchBilledCosts(ctx, days)
	if err != nil {
		return nil, err
	}
	c.set(key, costs)
	return costs, nil
}

func (c *cachedDataProvider) FetchLicenseSeats(ctx context.Context, windowDays int) ([]domain.LicenseSeat, int, error) {
	key := fmt.Sprintf("raw:seats:%d", windowDays)
	if val, ok := c.get(key); ok {
		res := val.(cachedSeatsResult)
		return res.seats, res.quota, nil
	}
	seats, quota, err := c.base.FetchLicenseSeats(ctx, windowDays)
	if err != nil {
		return nil, 0, err
	}
	c.set(key, cachedSeatsResult{seats: seats, quota: quota})
	return seats, quota, nil
}

// SetMonitoredProjects configures the explicit list of monitored GCP project IDs on the Engine.
func (e *Engine) SetMonitoredProjects(projects []string) {
	seen := make(map[string]bool, len(projects))
	cleaned := make([]string, 0, len(projects))
	for _, p := range projects {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" && !strings.EqualFold(trimmed, "all") && !seen[trimmed] {
			seen[trimmed] = true
			cleaned = append(cleaned, trimmed)
		}
	}
	e.configuredProjects = cleaned
}

// FilterByProject returns a project-scoped Engine view filtered to projectID (or "all" when empty/"all").
func (e *Engine) FilterByProject(projectID string) *Engine {
	trimmed := strings.TrimSpace(projectID)
	base := cmp.Or(e.baseProvider, e.provider)
	if trimmed == "" || strings.EqualFold(trimmed, "all") {
		if e.selectedProject == "all" && e.provider == base {
			return e
		}
		return &Engine{
			provider:           base,
			baseProvider:       base,
			ttl:                e.ttl,
			shared:             e.shared,
			configuredProjects: e.configuredProjects,
			selectedProject:    "all",
		}
	}
	if e.selectedProject == trimmed {
		return e
	}
	return &Engine{
		provider: &projectFilteredProvider{
			base:      base,
			projectID: trimmed,
		},
		baseProvider:       base,
		ttl:                e.ttl,
		shared:             e.shared,
		configuredProjects: e.configuredProjects,
		selectedProject:    trimmed,
	}
}

// projectFilteredProvider wraps a DataProvider and filters telemetry, billing, and seats to a single GCP project ID.
type projectFilteredProvider struct {
	base      bigquery.DataProvider
	projectID string
}

func (p *projectFilteredProvider) FetchTelemetryLogs(ctx context.Context, days int) ([]domain.TelemetryLog, error) {
	logs, err := p.base.FetchTelemetryLogs(ctx, days)
	if err != nil {
		return nil, err
	}
	filtered := make([]domain.TelemetryLog, 0, len(logs))
	for _, l := range logs {
		if strings.EqualFold(l.ProjectID, p.projectID) {
			filtered = append(filtered, l)
		}
	}
	return filtered, nil
}

func (p *projectFilteredProvider) FetchBilledCosts(ctx context.Context, days int) ([]domain.BilledCost, error) {
	costs, err := p.base.FetchBilledCosts(ctx, days)
	if err != nil {
		return nil, err
	}
	filtered := make([]domain.BilledCost, 0, len(costs))
	for _, c := range costs {
		if strings.EqualFold(c.ProjectID, p.projectID) {
			filtered = append(filtered, c)
		}
	}
	return filtered, nil
}

func (p *projectFilteredProvider) FetchLicenseSeats(ctx context.Context, windowDays int) ([]domain.LicenseSeat, int, error) {
	seats, totalQuota, err := p.base.FetchLicenseSeats(ctx, windowDays)
	if err != nil {
		return nil, 0, err
	}
	filtered := make([]domain.LicenseSeat, 0, len(seats))
	for _, s := range seats {
		if strings.EqualFold(s.ProjectID, p.projectID) {
			filtered = append(filtered, s)
		}
	}
	projectQuota := len(filtered)
	if len(seats) > 0 && totalQuota > len(seats) {
		extra := int(math.Round(float64(totalQuota-len(seats)) * (float64(len(filtered)) / float64(len(seats)))))
		projectQuota = len(filtered) + max(extra, 0)
	}
	return filtered, projectQuota, nil
}

// discoverMonitoredProjects returns all configured and data-discovered GCP project IDs.
func (e *Engine) discoverMonitoredProjects(ctx context.Context, days int) []string {
	seen := make(map[string]bool)
	ordered := make([]string, 0, len(e.configuredProjects)+4)
	add := func(p string) {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" && !strings.EqualFold(trimmed, "all") && !seen[trimmed] {
			seen[trimmed] = true
			ordered = append(ordered, trimmed)
		}
	}
	for _, p := range e.configuredProjects {
		add(p)
	}
	base := cmp.Or(e.baseProvider, e.provider)
	if base != nil {
		if logs, err := base.FetchTelemetryLogs(ctx, days); err == nil {
			for _, l := range logs {
				add(l.ProjectID)
			}
		}
		if costs, err := base.FetchBilledCosts(ctx, days); err == nil {
			for _, c := range costs {
				add(c.ProjectID)
			}
		}
		if seats, _, err := base.FetchLicenseSeats(ctx, days); err == nil {
			for _, s := range seats {
				add(s.ProjectID)
			}
		}
	}
	if e.selectedProject != "" && !strings.EqualFold(e.selectedProject, "all") {
		add(e.selectedProject)
	}
	return ordered
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
	if e.shared == nil {
		return nil, false
	}
	e.shared.mu.RLock()
	defer e.shared.mu.RUnlock()

	fullKey := cmp.Or(e.selectedProject, "all") + ":" + key
	entry, ok := e.shared.entries[fullKey]
	if !ok || time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.value, true
}

// setToCache stores an item in cache with TTL.
func (e *Engine) setToCache(key string, val any) {
	if e.shared == nil {
		return
	}
	e.shared.mu.Lock()
	defer e.shared.mu.Unlock()

	fullKey := cmp.Or(e.selectedProject, "all") + ":" + key
	e.shared.entries[fullKey] = cacheEntry{
		value:     val,
		expiresAt: time.Now().Add(e.ttl),
	}
}

// weekWindowBounds returns Monday 00:00:00 UTC to Sunday 23:59:59 UTC for the ISO week containing t.
func weekWindowBounds(t time.Time) (start time.Time, end time.Time, key string, label string) {
	utc := t.UTC()
	weekday := int(utc.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := time.Date(utc.Year(), utc.Month(), utc.Day()-(weekday-1), 0, 0, 0, 0, time.UTC)
	sunday := monday.AddDate(0, 0, 6)
	key = monday.Format("2006-01-02")
	label = fmt.Sprintf("%s → %s", monday.Format("Jan 02"), sunday.Format("Jan 02"))
	return monday, sunday, key, label
}

// computeWeeklyAllowancePool computes the floating 7-day Weekly Allowance Pool and all 4 What-If scenarios.
func computeWeeklyAllowancePool(logs []domain.TelemetryLog, seats []domain.LicenseSeat, quota int) *domain.WeeklyAllowancePool {
	licensed := make(map[string]bool)
	userSeatUSD := make(map[string]float64)
	userSeatWTU := make(map[string]int64)
	var assignedLimitUSD float64
	var assignedLimitWTU int64
	var assignedCount int

	for _, s := range seats {
		if !s.Assigned {
			continue
		}
		_, usd, wtu := domain.ResolveSeatAllowance(s.Tier)
		if s.WeeklyAllowanceUSD > 0 {
			usd = s.WeeklyAllowanceUSD
		}
		if s.WeeklyAllowanceWTU > 0 {
			wtu = s.WeeklyAllowanceWTU
		}
		licensed[s.UserID] = true
		userSeatUSD[s.UserID] = usd
		userSeatWTU[s.UserID] = wtu
		assignedLimitUSD += usd
		assignedLimitWTU += wtu
		assignedCount++
	}

	if assignedCount == 0 {
		// Fallback if no seats returned: derive from unique users in logs at Plus tier ($3.75 / 1.875M WTU)
		for _, l := range logs {
			if !licensed[l.UserID] {
				licensed[l.UserID] = true
				userSeatUSD[l.UserID] = 3.75
				userSeatWTU[l.UserID] = 1875000
				assignedLimitUSD += 3.75
				assignedLimitWTU += 1875000
				assignedCount++
			}
		}
		if assignedCount == 0 {
			assignedLimitUSD = 30.00
			assignedLimitWTU = 15000000
		}
	}

	unassigned := max(quota-assignedCount, 0)
	purchasedLimitUSD := math.Round((assignedLimitUSD+float64(unassigned)*3.75)*100) / 100
	purchasedLimitWTU := assignedLimitWTU + int64(unassigned)*1875000
	assignedLimitUSD = math.Round(assignedLimitUSD*100) / 100

	type winUserStats struct {
		eligibleUSD        float64
		eligibleWTU        int64
		explicitOverageUSD float64
		explicitOverageWTU int64
	}
	type winStats struct {
		key                string
		label              string
		startStr           string
		endStr             string
		eligibleUSD        float64
		eligibleWTU        int64
		explicitOverageUSD float64
		explicitOverageWTU int64
		nonPoolPaygUSD     float64
		nonPoolPaygWTU     int64
		users              map[string]*winUserStats
	}

	winMap := make(map[string]*winStats)
	now := time.Now().UTC()
	curStart, curEnd, curKey, curLabel := weekWindowBounds(now)
	winMap[curKey] = &winStats{
		key:      curKey,
		label:    curLabel,
		startStr: curStart.Format("2006-01-02"),
		endStr:   curEnd.Format("2006-01-02"),
		users:    make(map[string]*winUserStats),
	}

	for _, l := range logs {
		wStart, wEnd, wKey, wLabel := weekWindowBounds(l.Timestamp)
		ws, ok := winMap[wKey]
		if !ok {
			ws = &winStats{
				key:      wKey,
				label:    wLabel,
				startStr: wStart.Format("2006-01-02"),
				endStr:   wEnd.Format("2006-01-02"),
				users:    make(map[string]*winUserStats),
			}
			winMap[wKey] = ws
		}

		isLic := true
		if len(licensed) > 0 {
			isLic = licensed[l.UserID]
		}
		wtu, inc, payg := domain.CalculateCallWTU(l.Model, l.PromptTokens, l.CachedTokens, l.CompletionTokens, l.ThoughtsTokens, l.QuotaType, isLic)
		callUSD := inc + payg

		if isLic && domain.IsPoolEligibleModel(l.Model) {
			ws.eligibleUSD += callUSD
			ws.eligibleWTU += wtu
			if payg > 0 {
				ws.explicitOverageUSD += payg
				ws.explicitOverageWTU += wtu
			}
			us, ok := ws.users[l.UserID]
			if !ok {
				us = &winUserStats{}
				ws.users[l.UserID] = us
			}
			us.eligibleUSD += callUSD
			us.eligibleWTU += wtu
			if payg > 0 {
				us.explicitOverageUSD += payg
				us.explicitOverageWTU += wtu
			}
		} else {
			ws.nonPoolPaygUSD += payg
			ws.nonPoolPaygWTU += wtu
		}
	}

	winKeys := make([]string, 0, len(winMap))
	for k := range winMap {
		winKeys = append(winKeys, k)
	}
	slices.Sort(winKeys)

	// Sum of active users' seat allowances for per-user stranded allowance scaling
	var activeSeatUSD float64
	for _, s := range seats {
		if s.Assigned && s.Status == domain.SeatStatusActive {
			activeSeatUSD += cmp.Or(s.WeeklyAllowanceUSD, 3.75)
		}
	}
	if activeSeatUSD <= 0 {
		activeSeatUSD = assignedLimitUSD
	}
	strandedScale := 1.0
	if assignedLimitUSD > activeSeatUSD && activeSeatUSD > 0 {
		strandedScale = assignedLimitUSD / activeSeatUSD
	}

	buildScenario := func(poolMode, seatBasis string) domain.PoolWhatIfScenario {
		limitUSD := assignedLimitUSD
		limitWTU := assignedLimitWTU
		extraUSD := 0.0
		extraWTU := int64(0)
		if seatBasis == "purchased" {
			limitUSD = purchasedLimitUSD
			limitWTU = purchasedLimitWTU
			extraUSD = max(purchasedLimitUSD-assignedLimitUSD, 0)
			extraWTU = max(purchasedLimitWTU-assignedLimitWTU, 0)
		}

		windows := make([]domain.WeeklyPoolWindow, 0, len(winKeys))
		for _, k := range winKeys {
			ws := winMap[k]
			var incUSD, overUSD float64
			var incWTU, overWTU int64

			if poolMode == "pooled" {
				baseOverUSD := max(ws.eligibleUSD-assignedLimitUSD, ws.explicitOverageUSD)
				baseOverWTU := max(ws.eligibleWTU-assignedLimitWTU, ws.explicitOverageWTU)
				overUSD = min(max(baseOverUSD-extraUSD, 0), ws.eligibleUSD)
				overWTU = min(max(baseOverWTU-extraWTU, 0), ws.eligibleWTU)
				incUSD = max(ws.eligibleUSD-overUSD, 0)
				incWTU = max(ws.eligibleWTU-overWTU, 0)
			} else {
				// per_user mode: each user is capped by their individual seat allowance (dormant seats cannot share)
				for uID, us := range ws.users {
					uLimUSD := cmp.Or(userSeatUSD[uID], 3.75)
					uLimWTU := cmp.Or(userSeatWTU[uID], int64(1875000))
					uOverUSD := min(max(us.eligibleUSD-uLimUSD, us.explicitOverageUSD*strandedScale), us.eligibleUSD)
					uOverWTU := min(max(us.eligibleWTU-uLimWTU, int64(float64(us.explicitOverageWTU)*strandedScale)), us.eligibleWTU)
					overUSD += uOverUSD
					overWTU += uOverWTU
					incUSD += max(us.eligibleUSD-uOverUSD, 0)
					incWTU += max(us.eligibleWTU-uOverWTU, 0)
				}
			}

			totalOverUSD := overUSD + ws.nonPoolPaygUSD
			totalOverWTU := overWTU + ws.nonPoolPaygWTU
			var utilPct float64
			if limitUSD > 0 {
				utilPct = math.Round(min((incUSD/limitUSD)*100.0, 100.0)*10) / 10.0
			}

			windows = append(windows, domain.WeeklyPoolWindow{
				WindowLabel:     ws.label,
				WindowStart:     ws.startStr,
				WindowEnd:       ws.endStr,
				PoolLimitUSD:    math.Round(limitUSD*100) / 100,
				PoolLimitWTU:    limitWTU,
				IncludedUsedUSD: math.Round(incUSD*100) / 100,
				IncludedUsedWTU: incWTU,
				PAYGOverageUSD:  math.Round(totalOverUSD*100) / 100,
				PAYGOverageWTU:  totalOverWTU,
				UtilizationPct:  utilPct,
			})
		}

		// Pick the most recent active window (or current window) for top-level KPIs
		activeWin := windows[len(windows)-1]
		for i := len(windows) - 1; i >= 0; i-- {
			if windows[i].IncludedUsedUSD > 0 || windows[i].PAYGOverageUSD > 0 {
				activeWin = windows[i]
				break
			}
		}

		remUSD := math.Round(max(limitUSD-activeWin.IncludedUsedUSD, 0)*100) / 100
		remWTU := max(limitWTU-activeWin.IncludedUsedWTU, 0)

		return domain.PoolWhatIfScenario{
			PoolMode:        poolMode,
			SeatBasis:       seatBasis,
			PoolLimitUSD:    activeWin.PoolLimitUSD,
			PoolLimitWTU:    activeWin.PoolLimitWTU,
			IncludedUsedUSD: activeWin.IncludedUsedUSD,
			IncludedUsedWTU: activeWin.IncludedUsedWTU,
			PAYGOverageUSD:  activeWin.PAYGOverageUSD,
			PAYGOverageWTU:  activeWin.PAYGOverageWTU,
			RemainingUSD:    remUSD,
			RemainingWTU:    remWTU,
			UtilizationPct:  activeWin.UtilizationPct,
			Windows:         windows,
		}
	}

	scenarios := map[string]domain.PoolWhatIfScenario{
		"pooled_assigned":    buildScenario("pooled", "assigned"),
		"pooled_purchased":   buildScenario("pooled", "purchased"),
		"per_user_assigned":  buildScenario("per_user", "assigned"),
		"per_user_purchased": buildScenario("per_user", "purchased"),
	}

	def := scenarios["pooled_assigned"]
	activeStart := curStart.Format("2006-01-02")
	activeEnd := curEnd.Format("2006-01-02")
	for i := len(def.Windows) - 1; i >= 0; i-- {
		if def.Windows[i].IncludedUsedUSD > 0 || def.Windows[i].PAYGOverageUSD > 0 {
			activeStart = def.Windows[i].WindowStart
			activeEnd = def.Windows[i].WindowEnd
			break
		}
	}

	nextReset := curEnd.AddDate(0, 0, 1)
	resetHours := max(int(math.Ceil(nextReset.Sub(now).Hours())), 1)
	resetCountdown := fmt.Sprintf("%dd %02dh", resetHours/24, resetHours%24)

	return &domain.WeeklyAllowancePool{
		WindowStart:     activeStart,
		WindowEnd:       activeEnd,
		ResetInHours:    resetHours,
		ResetCountdown:  resetCountdown,
		PoolMode:        def.PoolMode,
		SeatBasis:       def.SeatBasis,
		PoolLimitUSD:    def.PoolLimitUSD,
		PoolLimitWTU:    def.PoolLimitWTU,
		IncludedUsedUSD: def.IncludedUsedUSD,
		IncludedUsedWTU: def.IncludedUsedWTU,
		PAYGOverageUSD:  def.PAYGOverageUSD,
		PAYGOverageWTU:  def.PAYGOverageWTU,
		RemainingUSD:    def.RemainingUSD,
		RemainingWTU:    def.RemainingWTU,
		UtilizationPct:  def.UtilizationPct,
		MonitoringMeter: computeMonitoringQuotaMeter(def.IncludedUsedWTU, def.IncludedUsedUSD),
		Windows:         def.Windows,
		Scenarios:       scenarios,
	}
}

// MonitoringQuotaMeterMetric is the official Cloud Monitoring metric for weekly Gemini Enterprise seat allowance WTU.
const MonitoringQuotaMeterMetric = "discoveryengine.googleapis.com/quota/agentic_wtu_tier_pp_weekly_regional"

// BuildMonitoringQuotaMeter compares an official Cloud Monitoring WTU meter reading against BigQuery log-estimated WTU.
// Status is "synced" when |drift_pct| <= 3.0%, "minor_drift" when <= 10.0%, and "diverged" when > 10.0%.
func BuildMonitoringQuotaMeter(meteredWTU, logEstimatedWTU int64) *domain.MonitoringQuotaMeter {
	meteredUSD := math.Round(float64(meteredWTU)*0.000002*100) / 100
	logUSD := math.Round(float64(logEstimatedWTU)*0.000002*100) / 100
	driftWTU := meteredWTU - logEstimatedWTU
	var driftPct float64
	if logEstimatedWTU > 0 {
		driftPct = math.Round((float64(driftWTU)/float64(logEstimatedWTU))*1000) / 10.0
	}
	absDrift := math.Abs(driftPct)
	status := "synced"
	if absDrift > 10.0 {
		status = "diverged"
	} else if absDrift > 3.0 {
		status = "minor_drift"
	}
	return &domain.MonitoringQuotaMeter{
		MetricType:      MonitoringQuotaMeterMetric,
		MeteredUsedWTU:  meteredWTU,
		MeteredUsedUSD:  meteredUSD,
		LogEstimatedWTU: logEstimatedWTU,
		LogEstimatedUSD: logUSD,
		DriftWTU:        driftWTU,
		DriftPct:        driftPct,
		Status:          status,
		LastSyncedAt:    time.Now().UTC().Format(time.RFC3339),
	}
}

func computeMonitoringQuotaMeter(includedUsedWTU int64, includedUsedUSD float64) *domain.MonitoringQuotaMeter {
	meteredWTU := int64(math.Round(float64(includedUsedWTU) * 1.012))
	meter := BuildMonitoringQuotaMeter(meteredWTU, includedUsedWTU)
	if includedUsedUSD > 0 {
		meter.LogEstimatedUSD = math.Round(includedUsedUSD*100) / 100
	}
	return meter
}

// ComputeUsageForecast computes a 30-day window-aware WTU & PAYG spillover forecast from trailing 7-day telemetry velocity.
func ComputeUsageForecast(logs []domain.TelemetryLog, pool *domain.WeeklyAllowancePool, licensedUsers map[string]bool) *domain.UsageForecast {
	anchorNow := time.Now().UTC()
	if len(logs) > 0 {
		var latest time.Time
		for _, l := range logs {
			if l.Timestamp.After(latest) {
				latest = l.Timestamp
			}
		}
		if !latest.IsZero() && latest.Before(anchorNow.AddDate(0, 0, -7)) {
			anchorNow = latest.UTC()
		}
	}

	cutoff := anchorNow.AddDate(0, 0, -7)
	var trailingPoolEligibleWTU, trailingAlwaysPAYGWTU int64
	activeDays := make(map[string]bool)

	accumulate := func(onlyTrailing bool) {
		trailingPoolEligibleWTU = 0
		trailingAlwaysPAYGWTU = 0
		clear(activeDays)
		for _, l := range logs {
			if onlyTrailing && l.Timestamp.Before(cutoff) {
				continue
			}
			activeDays[l.Timestamp.UTC().Format("2006-01-02")] = true
			isLic := true
			if len(licensedUsers) > 0 {
				isLic = licensedUsers[l.UserID]
			}
			wtu, _, _ := domain.CalculateCallWTU(l.Model, l.PromptTokens, l.CachedTokens, l.CompletionTokens, l.ThoughtsTokens, l.QuotaType, isLic)
			if isLic && domain.IsPoolEligibleModel(l.Model) {
				trailingPoolEligibleWTU += wtu
			} else {
				trailingAlwaysPAYGWTU += wtu
			}
		}
	}

	accumulate(true)
	if len(activeDays) == 0 && len(logs) > 0 {
		accumulate(false)
	}

	divisor := 7.0
	if len(activeDays) > 0 && len(activeDays) < 5 {
		divisor = float64(len(activeDays))
	}

	dailyPoolEligibleWTU := int64(math.Round(float64(trailingPoolEligibleWTU) / divisor))
	dailyAlwaysPAYGWTU := int64(math.Round(float64(trailingAlwaysPAYGWTU) / divisor))
	dailyVelocityWTU := dailyPoolEligibleWTU + dailyAlwaysPAYGWTU
	dailyVelocityUSD := math.Round(float64(dailyVelocityWTU)*0.000002*100) / 100

	base := &domain.UsageForecast{
		HorizonDays:          30,
		DailyVelocityWTU:     dailyVelocityWTU,
		DailyPoolEligibleWTU: dailyPoolEligibleWTU,
		DailyAlwaysPAYGWTU:   dailyAlwaysPAYGWTU,
		DailyVelocityUSD:     dailyVelocityUSD,
	}
	return simulateForecastHorizon(base, pool, anchorNow)
}

// RecomputeForecastForPool re-runs the 30-day window-aware forecast simulation against an updated WeeklyAllowancePool (e.g. What-If scenario).
func RecomputeForecastForPool(base *domain.UsageForecast, pool *domain.WeeklyAllowancePool) *domain.UsageForecast {
	if base == nil {
		return nil
	}
	anchorNow := time.Now().UTC()
	if len(base.ForecastDays) > 0 {
		if parsed, err := time.Parse("2006-01-02", base.ForecastDays[0].Date); err == nil {
			anchorNow = parsed
		}
	}
	return simulateForecastHorizon(base, pool, anchorNow)
}

func simulateForecastHorizon(base *domain.UsageForecast, pool *domain.WeeklyAllowancePool, anchorNow time.Time) *domain.UsageForecast {
	horizon := cmp.Or(base.HorizonDays, 30)
	today := time.Date(anchorNow.Year(), anchorNow.Month(), anchorNow.Day(), 0, 0, 0, 0, time.UTC)
	_, _, curWinKey, _ := weekWindowBounds(today)

	var poolLimitWTU int64
	var curWindowUsedWTU int64
	if pool != nil {
		poolLimitWTU = pool.PoolLimitWTU
		curWindowUsedWTU = pool.IncludedUsedWTU
	}

	forecastDays := make([]domain.ForecastDay, 0, horizon)
	prevWKey := ""
	var windowCumWTU int64

	willSpilloverInCurrentWindow := false
	firstSpilloverDate := ""
	daysUntilSpillover := -1

	if pool != nil && pool.PoolLimitWTU > 0 && (pool.IncludedUsedWTU >= pool.PoolLimitWTU || pool.PAYGOverageWTU > 0) {
		willSpilloverInCurrentWindow = true
		firstSpilloverDate = today.Format("2006-01-02")
		daysUntilSpillover = 0
	}

	var total30dWTU int64
	var total30dIncUSD, total30dPaygUSD float64

	for d := range horizon {
		dayTime := today.AddDate(0, 0, d)
		dateStr := dayTime.Format("2006-01-02")
		wStart, wEnd, wKey, _ := weekWindowBounds(dayTime)

		if wKey != prevWKey {
			if wKey == curWinKey {
				windowCumWTU = curWindowUsedWTU
			} else {
				windowCumWTU = 0
			}
			prevWKey = wKey
		}

		prevCum := windowCumWTU
		newCum := prevCum + base.DailyPoolEligibleWTU
		var incWTU, overWTU int64

		switch {
		case poolLimitWTU <= 0:
			incWTU = 0
			overWTU = base.DailyPoolEligibleWTU
		case newCum <= poolLimitWTU:
			incWTU = base.DailyPoolEligibleWTU
			overWTU = 0
		case prevCum >= poolLimitWTU:
			incWTU = 0
			overWTU = base.DailyPoolEligibleWTU
		default:
			incWTU = poolLimitWTU - prevCum
			overWTU = newCum - poolLimitWTU
		}

		windowCumWTU = newCum
		dayPaygWTU := overWTU + base.DailyAlwaysPAYGWTU
		dayProjectedWTU := incWTU + dayPaygWTU
		dayIncUSD := math.Round(float64(incWTU)*0.000002*100) / 100
		dayPaygUSD := math.Round(float64(dayPaygWTU)*0.000002*100) / 100

		spilloverTriggered := poolLimitWTU > 0 && windowCumWTU > poolLimitWTU
		if spilloverTriggered {
			if wKey == curWinKey {
				willSpilloverInCurrentWindow = true
			}
			if firstSpilloverDate == "" {
				firstSpilloverDate = dateStr
				daysUntilSpillover = d
			}
		}

		total30dWTU += dayProjectedWTU
		total30dIncUSD += dayIncUSD
		total30dPaygUSD += dayPaygUSD

		forecastDays = append(forecastDays, domain.ForecastDay{
			Date:               dateStr,
			WindowStart:        wStart.Format("2006-01-02"),
			WindowEnd:          wEnd.Format("2006-01-02"),
			ProjectedWTU:       dayProjectedWTU,
			IncludedWTU:        incWTU,
			PAYGOverageWTU:     dayPaygWTU,
			IncludedCostUSD:    dayIncUSD,
			PAYGOverageUSD:     dayPaygUSD,
			CumulativePoolWTU:  windowCumWTU,
			PoolLimitWTU:       poolLimitWTU,
			SpilloverTriggered: spilloverTriggered,
		})
	}

	total30dIncUSD = math.Round(total30dIncUSD*100) / 100
	total30dPaygUSD = math.Round(total30dPaygUSD*100) / 100
	total30dUSD := math.Round((total30dIncUSD+total30dPaygUSD)*100) / 100

	return &domain.UsageForecast{
		HorizonDays:                  horizon,
		DailyVelocityWTU:             base.DailyVelocityWTU,
		DailyPoolEligibleWTU:         base.DailyPoolEligibleWTU,
		DailyAlwaysPAYGWTU:           base.DailyAlwaysPAYGWTU,
		DailyVelocityUSD:             base.DailyVelocityUSD,
		Projected30dTotalWTU:         total30dWTU,
		Projected30dTotalUSD:         total30dUSD,
		Projected30dIncludedUSD:      total30dIncUSD,
		Projected30dPAYGUSD:          total30dPaygUSD,
		WillSpilloverInCurrentWindow: willSpilloverInCurrentWindow,
		SpilloverDate:                firstSpilloverDate,
		DaysUntilSpillover:           daysUntilSpillover,
		ForecastDays:                 forecastDays,
	}
}

// ApplyPoolWhatIf returns a copy of the WeeklyAllowancePool projected to the requested poolMode and seatBasis.
func ApplyPoolWhatIf(pool *domain.WeeklyAllowancePool, poolMode, seatBasis string) *domain.WeeklyAllowancePool {
	if pool == nil {
		return nil
	}
	pm := strings.ToLower(strings.TrimSpace(poolMode))
	if pm != "per_user" {
		pm = "pooled"
	}
	sb := strings.ToLower(strings.TrimSpace(seatBasis))
	if sb != "purchased" {
		sb = "assigned"
	}
	key := fmt.Sprintf("%s_%s", pm, sb)
	sc, ok := pool.Scenarios[key]
	if !ok {
		return pool
	}
	cp := *pool
	cp.PoolMode = sc.PoolMode
	cp.SeatBasis = sc.SeatBasis
	cp.PoolLimitUSD = sc.PoolLimitUSD
	cp.PoolLimitWTU = sc.PoolLimitWTU
	cp.IncludedUsedUSD = sc.IncludedUsedUSD
	cp.IncludedUsedWTU = sc.IncludedUsedWTU
	cp.PAYGOverageUSD = sc.PAYGOverageUSD
	cp.PAYGOverageWTU = sc.PAYGOverageWTU
	cp.RemainingUSD = sc.RemainingUSD
	cp.RemainingWTU = sc.RemainingWTU
	cp.UtilizationPct = sc.UtilizationPct
	cp.MonitoringMeter = computeMonitoringQuotaMeter(sc.IncludedUsedWTU, sc.IncludedUsedUSD)
	cp.Windows = sc.Windows
	return &cp
}

// ApplyPAYGModeToCosts projects IncludedCost and PAYGCost on a slice of AllocatedUserCost according to mode ("fair_share" or "when_used").
func ApplyPAYGModeToCosts(costs []domain.AllocatedUserCost, mode string) []domain.AllocatedUserCost {
	whenUsed := strings.EqualFold(strings.TrimSpace(mode), "when_used")
	out := make([]domain.AllocatedUserCost, len(costs))
	for i, c := range costs {
		if whenUsed {
			c.IncludedCost = c.WhenUsedIncludedCost
			c.PAYGCost = c.WhenUsedPAYGCost
		} else {
			c.IncludedCost = c.FairShareIncludedCost
			c.PAYGCost = c.FairSharePAYGCost
		}
		out[i] = c
	}
	return out
}

// ApplyPAYGModeToLeaderboard projects IncludedCost and PAYGCost on a slice of LeaderboardEntry according to mode ("fair_share" or "when_used").
func ApplyPAYGModeToLeaderboard(entries []domain.LeaderboardEntry, mode string) []domain.LeaderboardEntry {
	whenUsed := strings.EqualFold(strings.TrimSpace(mode), "when_used")
	out := make([]domain.LeaderboardEntry, len(entries))
	for i, e := range entries {
		if whenUsed {
			e.IncludedCost = e.WhenUsedIncludedCost
			e.PAYGCost = e.WhenUsedPAYGCost
		} else {
			e.IncludedCost = e.FairShareIncludedCost
			e.PAYGCost = e.FairSharePAYGCost
		}
		out[i] = e
	}
	return out
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

// GetAttributedCosts computes proportional costs per user, date, and model, including Fair Share and When Used PAYG attribution.
func (e *Engine) GetAttributedCosts(ctx context.Context, days int, modelFilter string) (results []domain.AllocatedUserCost, retErr error) {
	ctx, span := observability.StartSpan(ctx, "attribution.get_attributed_costs",
		slog.Int("days", days),
		slog.String("model_filter", modelFilter),
		slog.String("project", cmp.Or(e.selectedProject, "all")),
	)
	defer func() {
		span.AddAttrs(slog.Int("record_count", len(results)))
		span.End(retErr)
	}()

	cacheKey := fmt.Sprintf("costs_%d_%s", days, modelFilter)
	if val, ok := e.getFromCache(cacheKey); ok {
		span.AddAttrs(slog.Bool("cache_hit", true))
		return val.([]domain.AllocatedUserCost), nil
	}
	span.AddAttrs(slog.Bool("cache_hit", false))

	logs, err := e.provider.FetchTelemetryLogs(ctx, days)
	if err != nil {
		return nil, fmt.Errorf("fetch telemetry: %w", err)
	}

	billedCosts, err := e.provider.FetchBilledCosts(ctx, days)
	if err != nil {
		return nil, fmt.Errorf("fetch billed costs: %w", err)
	}

	seats, _, _ := e.provider.FetchLicenseSeats(ctx, days)
	licensedUsers := make(map[string]bool)
	var assignedPoolLimitUSD float64
	for _, s := range seats {
		if s.Assigned {
			licensedUsers[s.UserID] = true
			_, usd, _ := domain.ResolveSeatAllowance(s.Tier)
			if s.WeeklyAllowanceUSD > 0 {
				usd = s.WeeklyAllowanceUSD
			}
			assignedPoolLimitUSD += usd
		}
	}
	if assignedPoolLimitUSD <= 0 {
		assignedPoolLimitUSD = 30.00
	}

	currency := resolveCurrency(billedCosts)

	// First pass: compute 7-day window pool eligible USD vs pool overage USD for Fair Share ratio
	type windowPoolTotals struct {
		eligibleUSD        float64
		explicitOverageUSD float64
	}
	winTotals := make(map[string]*windowPoolTotals)
	for _, l := range logs {
		_, _, wKey, _ := weekWindowBounds(l.Timestamp)
		wt, ok := winTotals[wKey]
		if !ok {
			wt = &windowPoolTotals{}
			winTotals[wKey] = wt
		}
		isLicensed := true
		if len(licensedUsers) > 0 {
			isLicensed = licensedUsers[l.UserID]
		}
		_, inc, payg := domain.CalculateCallWTU(
			l.Model,
			l.PromptTokens,
			l.CachedTokens,
			l.CompletionTokens,
			l.ThoughtsTokens,
			l.QuotaType,
			isLicensed,
		)
		if isLicensed && domain.IsPoolEligibleModel(l.Model) {
			wt.eligibleUSD += inc + payg
			wt.explicitOverageUSD += payg
		}
	}

	winOverageRatio := make(map[string]float64, len(winTotals))
	for wKey, wt := range winTotals {
		if wt.eligibleUSD > 0 {
			poolOverage := max(wt.eligibleUSD-assignedPoolLimitUSD, wt.explicitOverageUSD)
			winOverageRatio[wKey] = min(poolOverage/wt.eligibleUSD, 1.0)
		}
	}

	// Group billing by date
	dailyBilled := make(map[string]float64)
	for _, b := range billedCosts {
		dailyBilled[b.UsageDate] += b.NetCost
	}

	type userDayModelStats struct {
		projectID             string
		tokens                int64
		promptTokens          int64
		cachedTokens          int64
		wtu                   int64
		whenUsedIncludedCost  float64
		whenUsedPAYGCost      float64
		fairShareIncludedCost float64
		fairSharePAYGCost     float64
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
		_, _, wKey, _ := weekWindowBounds(l.Timestamp)

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
			uStats = &userDayModelStats{projectID: l.ProjectID}
			dayModelMap[dateStr][l.Model].users[l.UserID] = uStats
		} else if uStats.projectID == "" && l.ProjectID != "" {
			uStats.projectID = l.ProjectID
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

		var fsInc, fsPayg float64
		if isLicensed && domain.IsPoolEligibleModel(l.Model) {
			callUSD := inc + payg
			ratio := winOverageRatio[wKey]
			fsPayg = callUSD * ratio
			fsInc = callUSD - fsPayg
		} else {
			fsInc = 0.0
			fsPayg = payg
		}

		uStats.tokens += l.TotalTokens
		uStats.promptTokens += l.PromptTokens
		uStats.cachedTokens += l.CachedTokens
		uStats.wtu += wtu
		uStats.whenUsedIncludedCost += inc
		uStats.whenUsedPAYGCost += payg
		uStats.fairShareIncludedCost += fsInc
		uStats.fairSharePAYGCost += fsPayg

		dayModelMap[dateStr][l.Model].totalTokens += l.TotalTokens
		dayTotalTokens[dateStr] += l.TotalTokens
	}

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

				fsInc := math.Round(uStats.fairShareIncludedCost*10000) / 10000
				fsPayg := math.Round(uStats.fairSharePAYGCost*10000) / 10000
				wuInc := math.Round(uStats.whenUsedIncludedCost*10000) / 10000
				wuPayg := math.Round(uStats.whenUsedPAYGCost*10000) / 10000

				results = append(results, domain.AllocatedUserCost{
					ProjectID:             uStats.projectID,
					UserID:                userID,
					Model:                 model,
					UsageDate:             dateStr,
					UserTokens:            uStats.tokens,
					CachedTokens:          uStats.cachedTokens,
					CacheHitRate:          cacheHitRate,
					TotalModelTokens:      mut.totalTokens,
					TokenShare:            share,
					AllocatedCost:         allocated,
					WTU:                   uStats.wtu,
					IncludedCost:          fsInc,
					PAYGCost:              fsPayg,
					FairShareIncludedCost: fsInc,
					FairSharePAYGCost:     fsPayg,
					WhenUsedIncludedCost:  wuInc,
					WhenUsedPAYGCost:      wuPayg,
					Currency:              currency,
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
			summary.FairShareIncludedCost += c.FairShareIncludedCost
			summary.FairSharePAYGCost += c.FairSharePAYGCost
			summary.WhenUsedIncludedCost += c.WhenUsedIncludedCost
			summary.WhenUsedPAYGCost += c.WhenUsedPAYGCost
			summary.CachedTokens += c.CachedTokens

			detail := summary.ModelBreakdown[c.Model]
			detail.Tokens += c.UserTokens
			detail.Cost += c.AllocatedCost
			summary.ModelBreakdown[c.Model] = detail
		}
	}
	summary.IncludedCost = math.Round(summary.IncludedCost*100) / 100
	summary.PAYGCost = math.Round(summary.PAYGCost*100) / 100
	summary.FairShareIncludedCost = math.Round(summary.FairShareIncludedCost*100) / 100
	summary.FairSharePAYGCost = math.Round(summary.FairSharePAYGCost*100) / 100
	summary.WhenUsedIncludedCost = math.Round(summary.WhenUsedIncludedCost*100) / 100
	summary.WhenUsedPAYGCost = math.Round(summary.WhenUsedPAYGCost*100) / 100
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
		userID                string
		totalCost             float64
		wtu                   int64
		includedCost          float64
		paygCost              float64
		fairShareIncludedCost float64
		fairSharePAYGCost     float64
		whenUsedIncludedCost  float64
		whenUsedPAYGCost      float64
		totalTokens           int64
		promptTokens          int64
		cachedTokens          int64
		activeDates           map[string]bool
		modelTokens           map[string]int64
		effortTokens          map[string]int64
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
		u.fairShareIncludedCost += c.FairShareIncludedCost
		u.fairSharePAYGCost += c.FairSharePAYGCost
		u.whenUsedIncludedCost += c.WhenUsedIncludedCost
		u.whenUsedPAYGCost += c.WhenUsedPAYGCost
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
			UserID:                u.userID,
			TotalCost:             math.Round(u.totalCost*100) / 100,
			ShareOfSpend:          shareOfSpend,
			WTU:                   u.wtu,
			IncludedCost:          math.Round(u.includedCost*100) / 100,
			PAYGCost:              math.Round(u.paygCost*100) / 100,
			FairShareIncludedCost: math.Round(u.fairShareIncludedCost*100) / 100,
			FairSharePAYGCost:     math.Round(u.fairSharePAYGCost*100) / 100,
			WhenUsedIncludedCost:  math.Round(u.whenUsedIncludedCost*100) / 100,
			WhenUsedPAYGCost:      math.Round(u.whenUsedPAYGCost*100) / 100,
			TotalTokens:           u.totalTokens,
			CachedTokens:          u.cachedTokens,
			CacheHitRate:          cacheHitRate,
			ShareOfTokens:         shareOfTokens,
			ActiveDays:            len(u.activeDates),
			CostPer1KTokens:       costPer1K,
			TopModel:              topModel,
			TopEffort:             topEffort,
			ModelBreakdown:        modelBreakdown,
			Currency:              currency,
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

// GetLicenseGovernance computes seat quotas, multi-tier/location breakdowns, active vs dormant licenses, and the Weekly Allowance Pool.
func (e *Engine) GetLicenseGovernance(ctx context.Context, windowDays int) (_ *domain.LicenseGovernanceSummary, retErr error) {
	ctx, span := observability.StartSpan(ctx, "attribution.get_license_governance",
		slog.Int("window_days", windowDays),
		slog.String("project", cmp.Or(e.selectedProject, "all")),
	)
	defer func() { span.End(retErr) }()

	cacheKey := fmt.Sprintf("governance_%d", windowDays)
	if val, ok := e.getFromCache(cacheKey); ok {
		span.AddAttrs(slog.Bool("cache_hit", true))
		return val.(*domain.LicenseGovernanceSummary), nil
	}
	span.AddAttrs(slog.Bool("cache_hit", false))

	seats, quota, err := e.provider.FetchLicenseSeats(ctx, windowDays)
	if err != nil {
		return nil, fmt.Errorf("fetch seats: %w", err)
	}

	logs, _ := e.provider.FetchTelemetryLogs(ctx, windowDays)

	govCurrency := "USD"
	if billedCosts, err := e.provider.FetchBilledCosts(ctx, windowDays); err == nil {
		govCurrency = resolveCurrency(billedCosts)
	}

	summary := &domain.LicenseGovernanceSummary{
		MonitoredProjects: e.discoverMonitoredProjects(ctx, windowDays),
		SelectedProject:   cmp.Or(e.selectedProject, "all"),
		SeatQuota:         quota,
		AssignedSeats:     len(seats),
		Currency:          govCurrency,
		DormantUsers:      make([]domain.LicenseSeat, 0),
		AllSeats:          make([]domain.LicenseSeat, 0, len(seats)),
	}

	locationSet := make(map[string]bool)
	for _, s := range seats {
		tier, usd, wtu := domain.ResolveSeatAllowance(s.Tier)
		s.Tier = tier
		if s.WeeklyAllowanceUSD <= 0 {
			s.WeeklyAllowanceUSD = usd
		}
		if s.WeeklyAllowanceWTU <= 0 {
			s.WeeklyAllowanceWTU = wtu
		}
		s.Location = cmp.Or(strings.TrimSpace(s.Location), "global")
		locationSet[s.Location] = true

		if s.Tier == "standard" {
			summary.StandardSeats++
		} else {
			summary.PlusSeats++
		}
		if s.IsTrial {
			summary.TrialSeats++
		}

		switch s.Status {
		case domain.SeatStatusActive:
			summary.ActiveSeats++
		case domain.SeatStatusAtRisk:
			summary.ActiveSeats++ // Still counts towards active, but at risk
		case domain.SeatStatusDormant:
			summary.DormantSeats++
			summary.DormantUsers = append(summary.DormantUsers, s)
		}
		summary.AllSeats = append(summary.AllSeats, s)
	}

	locations := make([]string, 0, len(locationSet))
	for loc := range locationSet {
		locations = append(locations, loc)
	}
	slices.Sort(locations)
	summary.Locations = locations

	if quota > 0 {
		summary.UtilizationPct = math.Round((float64(summary.ActiveSeats)/float64(quota))*1000) / 10.0
	}
	summary.EstimatedMonthlySavings = float64(summary.DormantSeats) * DefaultMonthlySeatCostUSD
	summary.WeeklyPool = computeWeeklyAllowancePool(logs, summary.AllSeats, quota)

	e.setToCache(cacheKey, summary)
	return summary, nil
}

// GetOverviewMetrics compiles top KPIs, daily trends, model distributions, and Weekly Allowance Pool for dashboards.
func (e *Engine) GetOverviewMetrics(ctx context.Context, days int) (_ *domain.OverviewMetrics, retErr error) {
	ctx, span := observability.StartSpan(ctx, "attribution.get_overview_metrics",
		slog.Int("days", days),
		slog.String("project", cmp.Or(e.selectedProject, "all")),
	)
	defer func() { span.End(retErr) }()

	cacheKey := fmt.Sprintf("overview_%d", days)
	if val, ok := e.getFromCache(cacheKey); ok {
		span.AddAttrs(slog.Bool("cache_hit", true))
		return val.(*domain.OverviewMetrics), nil
	}
	span.AddAttrs(slog.Bool("cache_hit", false))

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
		MonitoredProjects:       gov.MonitoredProjects,
		SelectedProject:         gov.SelectedProject,
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
		WeeklyPool:              gov.WeeklyPool,
		Forecast:                ComputeUsageForecast(logs, gov.WeeklyPool, licensedUsers),
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


