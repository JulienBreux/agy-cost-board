# Specification: Phase 1 (`v0.10.0`) — Token Cache, Thinking Effort, Client Telemetry & Included vs. PAYG Coexistence

## 1. Overview
This track enhances `agy-cost-board` so that it surfaces **both** its existing **Cloud Billing Export Reconciled Cost** and **Gemini Enterprise's Metered Model** (**Weighted Token Units — WTU**, **Included Seat Allowance Usage**, and **PAYG Overage**), alongside deep inference telemetry for **Cached Input Tokens**, **Thinking Effort Levels** (`minimal`..`max`), and **Client Applications** (`client_name` / `client_version`).

## 2. Functional Requirements

### 2.1 Telemetry Extraction & Domain Model (`internal/domain`, `internal/bigquery`)
1. **Extended Inference Fields**:
   - `CachedTokens` (`int64`): Extracted from `jsonPayload.metadata.cachedContentTokenCount`.
   - `QuotaType` (`string`): Extracted from `jsonPayload.metadata.quotaType` (`REGULAR_QUOTA`, `OVERAGE_QUOTA`, `NOOP_QUOTA`).
   - `EffortLevel` (`string`): Extracted from `jsonPayload.experience` suffix (`minimal`, `low`, `medium`, `high`, `xhigh`, `max`, or `default`).
   - `ClientName` & `ClientVersion` (`string`): Extracted from `labels.client_name` and `labels.client_version` (e.g., `vscode`, `jetbrains`, `agy-cli`, `agy-sdk`).
2. **WTU & Included vs. PAYG Calculation (`internal/attribution`)**:
   - Compute **Weighted Token Units (WTU)** per call/group (`1M WTU = $2.00`, minimum `50 WTU` per call, `0.5x` meter share for Gemini Flash 3.6–3.8 & Gemini 4 Argon, `1.0x` for other models, cached token discount rate).
   - Classify metered usage into:
     - **Included Usage (`IncludedCost` / `IncludedWTU`)**: Usage covered by the Gemini Enterprise seat allowance (`REGULAR_QUOTA`, or licensed Gemini calls within weekly pool limits).
     - **PAYG Overage (`PAYGCost` / `PAYGWTU`)**: Usage billed on top of subscriptions (`OVERAGE_QUOTA`, unlicensed users, or Anthropic/Claude models which are always 100% PAYG).
   - Coexist seamlessly with **Reconciled Cloud Billing Cost (`TotalBilledCost` / `AttributedCost`)**.
3. **Aggregated Breakdowns in `OverviewMetrics`, `LeaderboardMetrics`, and `UserDashboardMetrics`**:
   - **Token & Cost Share Breakdown (`TokenCostShares`)**: Share of tokens vs. share of cost across 4 token categories: `Uncached Input`, `Cached Input`, `Output`, and `Thinking`.
   - **Model & Effort Breakdown (`ModelEffortStats`)**: Calls, total tokens, thinking tokens, cache hit rate %, WTU, Included cost, PAYG cost, and Reconciled Billed cost grouped by `(Model, EffortLevel)`.
   - **Client Breakdown (`ClientStats`)**: Calls, tokens, share %, and cost grouped by `(ClientName, ClientVersion)`.
   - **Per-User Cache Hit Rate % (`CacheHitRate`)**, **Included Cost**, and **PAYG Cost** on Leaderboard entries, Attributed Cost rows, and Personal Consumption dashboards.

### 2.2 Web UI Enhancements (`web/src/`)
1. **Overview Tab**:
   - **KPI Cards (`KPICards.tsx`)**: Display Total Reconciled Billed Cost alongside Metered Usage (`Included Pool Usage` + `PAYG Overage`), Total Tokens (with **Cache Hit Rate %** badge), Active Licenses, and Potential Savings.
   - **Daily Cost Chart (`CostChart.tsx`)**: Add a view mode toggle between **By Model (Billed)** and **Included vs. PAYG Breakdown**.
   - **New "Tokens, Effort & Clients" Section (`TokensEffortClientSection.tsx`)**:
     - **"Where Tokens and Money Go"**: Visual comparison of Token Share % vs. Cost Share % for Uncached Input, Cached Input, Output, and Thinking tokens.
     - **"Models & Thinking Effort" Table**: Shows how model selection and thinking effort (`minimal`..`max`) drive thinking tokens and spend.
     - **"Antigravity Clients" Table**: Shows call volume, token share, and cost across IDE/CLI clients (`vscode`, `jetbrains`, `agy-cli`, etc.).
2. **Leaderboard (`LeaderboardView.tsx`) & Attributed Costs (`CostTable.tsx`)**:
   - Show **Included vs. PAYG** split and **Cache Hit Rate %** per developer alongside Reconciled Billed Cost.
3. **My Consumption (`UserDashboardView.tsx`, `PersonalKPICards.tsx`, `OptimizationCard.tsx`)**:
   - Show personal **Included vs. PAYG** usage, **Cache Hit Rate %**, and **Effort Level** recommendations (e.g., alerting when high/max effort or low cache reuse drives disproportionate cost).

### 2.3 CLI, TUI & MCP Parity (`internal/cli`, `internal/tui`, `internal/mcp`)
- Expose `IncludedCost`, `PAYGCost`, `WTU`, `CacheHitRate`, `EffortLevel`, and `ClientBreakdown` in JSON/table outputs, TUI summary views, and MCP server tools (`get_overview_metrics`, `get_users_costs`, `get_user_dashboard`).

## 3. Non-Functional Requirements & Acceptance Criteria
- **Zero-Config Demo Mode (`--demo`)**: Realistic synthetic telemetry in `internal/bigquery/demo.go` with varied `cached_tokens`, `quota_type` (`REGULAR_QUOTA` / `OVERAGE_QUOTA`), `effort_level` (`low`, `medium`, `high`, `max`), and `client_name` (`vscode`, `jetbrains`, `agy-cli`).
- **Backwards Compatibility**: Existing BigQuery tables that lack newer columns gracefully default (`cached_tokens = 0`, `effort_level = 'default'`, `client_name = 'antigravity'`).
- **Test Coverage**: >80% unit test coverage across Go (`go test ./...`) and React/TypeScript (`CI=true npm test`).
