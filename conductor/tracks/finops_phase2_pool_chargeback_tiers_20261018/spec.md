# Specification: Phase 2 (`v0.11.0`) — Weekly Allowance Pool Engine, Fair-Share vs. When-Used Chargeback & Multi-Tier License Discovery

## 1. Overview
Phase 2 (`v0.11.0`) extends `agy-cost-board` with **Gemini Enterprise's 7-day floating Weekly Allowance Pool**, an interactive **What-If Pool Simulator (`pooled` vs `per_user`, `purchased` vs `assigned`)**, **Fair Share vs. When Used PAYG Chargeback Attribution**, **CSV Chargeback Export**, and **Multi-Tier (`Standard` / `Plus`) & Multi-Location (`global`, `eu`, `us`) License Discovery** across the Go attribution engine, REST API, React Web UI, CLI, TUI, and MCP server.

## 2. Functional Requirements

### 2.1 Multi-Tier (`Standard` / `Plus`) & Multi-Location (`global`, `eu`, `us`) License Discovery
- Extend `LicenseSeat` in `internal/domain/models.go` with:
  - `Tier` (`"standard"` or `"plus"`, default `"plus"`):
    - `Standard`: `$10/mo` subscription tier = `$2.50/week` (`1,250,000 WTU/week` at `$2.00 / 1M WTU`) weekly allowance pool contribution per seat.
    - `Plus`: `$15/mo` subscription tier = `$3.75/week` (`1,875,000 WTU/week` at `$2.00 / 1M WTU`) weekly allowance pool contribution per seat.
  - `WeeklyAllowanceUSD` (`float64`) and `WeeklyAllowanceWTU` (`int64`).
  - `Location` (`"global"`, `"eu"`, `"us"`).
  - `IsTrial` (`bool`) and `TrialEndsAt` (`string`, optional).
- Extend `LicenseGovernanceSummary` with:
  - `AllSeats []LicenseSeat` (`json:"all_seats,omitempty"`).
  - `StandardSeats int` (`json:"standard_seats"`), `PlusSeats int` (`json:"plus_seats"`), `TrialSeats int` (`json:"trial_seats"`).
  - `Locations []string` (`json:"locations,omitempty"`).
  - `WeeklyPool *WeeklyAllowancePool` (`json:"weekly_pool,omitempty"`).

### 2.2 Weekly Allowance Pool & What-If Simulator
- Introduce `WeeklyAllowancePool`, `WeeklyPoolWindow`, and `PoolWhatIfScenario` in `internal/domain/models.go`:
  - `WindowStart` / `WindowEnd` (ISO-8601 date strings for the current 7-day allowance window) and `ResetInHours` (`int`) / `ResetCountdown` (`string`, e.g. `"2d 14h"`).
  - `PoolMode` (`"pooled"` vs `"per_user"`, default `"pooled"`) and `SeatBasis` (`"purchased"` vs `"assigned"`, default `"assigned"`).
  - `PoolLimitUSD` (`float64`), `PoolLimitWTU` (`int64`), `IncludedUsedUSD` (`float64`), `IncludedUsedWTU` (`int64`), `PAYGOverageUSD` (`float64`), `PAYGOverageWTU` (`int64`), `RemainingUSD` (`float64`), `RemainingWTU` (`int64`), and `UtilizationPct` (`float64`).
  - `Windows []WeeklyPoolWindow`: 7-day rolling window history across the lookback period (`window_label`, `window_start`, `window_end`, `pool_limit_usd`, `included_used_usd`, `payg_overage_usd`, `utilization_pct`).
  - `Scenarios map[string]PoolWhatIfScenario`: Pre-computed What-If scenarios keyed by `"pooled_assigned"`, `"pooled_purchased"`, `"per_user_assigned"`, `"per_user_purchased"` so the Web UI can toggle What-If modes instantaneously without network round-trips, in addition to optional API query parameters (`?pool_mode=pooled|per_user&seat_basis=purchased|assigned`).
- Expose `WeeklyPool *WeeklyAllowancePool` on both `OverviewMetrics` and `LicenseGovernanceSummary`.

### 2.3 Fair Share vs. When Used PAYG Chargeback Attribution
- Compute both PAYG attribution modes in `internal/attribution/engine.go`:
  - **When Used (`when_used`)**: Chronological / log-tagged attribution (`WhenUsedIncludedCost`, `WhenUsedPAYGCost`) where calls made after the 7-day pool limit is reached (or tagged `OVERAGE_QUOTA`), unlicensed user calls, and Anthropic Claude calls are charged to the caller at the moment of call.
  - **Fair Share (`fair_share`, default)**: For each 7-day allowance window, pool-eligible usage from licensed users shares the window's `PoolLimitUSD` and any pool overage proportionally according to each licensed user's share of pool-eligible WTU in that window (`FairShareIncludedCost`, `FairSharePAYGCost`), while unlicensed user calls and Anthropic Claude calls remain 100% PAYG on their callers.
- Populate `FairShareIncludedCost`, `FairSharePAYGCost`, `WhenUsedIncludedCost`, and `WhenUsedPAYGCost` on `AllocatedUserCost`, `LeaderboardEntry`, and `UserCostSummary`.
- Default `IncludedCost` and `PAYGCost` to `FairShare` unless `?payg_mode=when_used` is passed on `/api/costs` or `/api/leaderboard`.

### 2.4 Web UI Enhancements
- **Weekly Allowance Pool Card (`web/src/components/WeeklyAllowancePoolCard.tsx`)**:
  - Rendered in both **Overview** and **Seat Governance** tabs.
  - Displays current 7-day window drawdown progress bar (`Included Used / Weekly Pool Limit`), reset countdown badge, and PAYG spillover status.
  - Interactive What-If Simulator toggles: **Pooled vs. Per-User** and **Purchased Quota vs. Assigned Seats**.
  - 7-day window history comparison bars.
- **Fair Share / When Used Toggle & CSV Export (`LeaderboardView.tsx`, `CostTable.tsx`)**:
  - Segmented toggle (`Fair Share` / `When Used`) switching displayed `Included / PAYG` split and KPIs.
  - **Export CSV** button downloading a FinOps chargeback report including `Fair Share` and `When Used` columns.
- **Seat Governance (`LicenseTable.tsx`)**:
  - Displays `Standard` (`$2.50/wk`) vs. `Plus` (`$3.75/wk`) tier breakdown, regional location badges (`global`, `eu`, `us`), `Free Trial` badge, and weekly pool contribution per seat.

### 2.5 CLI, TUI & MCP Parity
- Update `agy-cost-board license`, `agy-cost-board cost`, `agy-cost-board leaderboard`, Interactive TUI (`internal/tui`), and MCP tools (`get_overview_metrics`, `get_license_status`, `get_users_costs`) to include Weekly Allowance Pool metrics, `Standard`/`Plus` tiers, and `Fair Share` vs. `When Used` PAYG breakdowns.

## 3. Acceptance Criteria
- All Go unit tests (`go test ./...`) and Frontend Vitest tests (`npm --prefix web test -- --run`) pass with >80% coverage on new logic.
- Demo mode (`--demo`) simulates realistic `Standard` and `Plus` seats across `global`, `eu`, and `us`, 7-day weekly pool windows with spillover, and distinct `Fair Share` vs. `When Used` PAYG distributions.
