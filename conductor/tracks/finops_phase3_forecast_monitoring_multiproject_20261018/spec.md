# Specification: Phase 3 (`v0.12.0`) — 30-Day Window-Aware Forecast, Cloud Monitoring Quota Meter Sync & Multi-Project Support

## 1. Overview
This track delivers **Phase 3 (`v0.12.0`)** of the `agy-finops` FinOps feature roadmap for `agy-cost-board`:
1. **30-Day Window-Aware Usage & PAYG Spillover Forecast**: Replays the trailing 7-day daily WTU and cost velocity across upcoming 7-day Gemini Enterprise allowance pool windows over the next 30 days to predict the exact **PAYG Spillover Date**, projected 30-day **Included** vs. **PAYG Overage** spend, and daily forecast series.
2. **Cloud Monitoring Weekly Quota Meter Comparison (`agentic_wtu_tier_*_pp_weekly_regional`)**: Compares the official Cloud Monitoring weekly WTU quota meter against BigQuery log-derived WTU estimates (with realistic simulated meter telemetry in `--demo` mode) and surfaces a **Cloud Monitoring Quota Meter vs. Log Estimate** reconciliation badge inside the `WeeklyAllowancePoolCard`.
3. **Multi-Project Support (`--monitored-projects` / `AGY_COST_BOARD_MONITORED_PROJECTS`)**: Enables monitoring multiple GCP projects from a single board with a top-navbar **Project Selector** (`All Projects` + individual projects) and `?project=` filtering across REST API, Web UI, CLI, TUI, and MCP.

## 2. Functional Requirements

### 2.1 30-Day Window-Aware Usage & PAYG Spillover Forecast
- **Forecast Engine (`internal/attribution/engine.go`, `internal/domain/models.go`)**:
  - Compute trailing 7-day daily velocity (`daily_pool_wtu`, `daily_direct_payg_usd`, `daily_total_tokens`).
  - Simulate the next 30 calendar days across rolling 7-day allowance pool windows (resetting remaining pool allowance at each 7-day boundary).
  - Produce a `UsageForecast` struct attached to `OverviewMetrics` (`forecast`):
    - `horizon_days` (`30`)
    - `daily_avg_wtu`, `daily_avg_cost_usd`
    - `projected_30d_total_usd`, `projected_30d_included_usd`, `projected_30d_payg_usd`, `projected_30d_wtu`
    - `spillover_date` (first forecasted date where weekly pool exhausts and spills into PAYG overage, or `""` if within pool)
    - `days_until_spillover` (`int`, `-1` if no spillover within 30 days)
    - `forecast_days` (`[]ForecastDay` with `date`, `included_cost`, `payg_cost`, `total_cost`, `wtu`, `pool_remaining_wtu`, `is_spillover`)
- **Web UI (`ForecastCard.tsx` in Overview tab)**:
  - Render a **30-Day Window-Aware PAYG Spillover Forecast** panel in the **Overview** tab displaying:
    - KPI summary badges: **Projected 30d Spend** (`Included` + `PAYG`), **Next Pool Exhaustion / Spillover Date** (`in X days` or `Pool Sufficient`), and **Trailing 7d Velocity** (`WTU/day`).
    - Interactive 30-day stacked bar chart showing projected daily **Included** (emerald) vs. **PAYG Spillover** (amber/rose) spend with 7-day window reset markers.

### 2.2 Cloud Monitoring Weekly Quota Meter Sync
- **Domain & Data (`internal/domain/models.go`, `internal/bigquery/demo.go`, `internal/bigquery/client.go`)**:
  - Extend `WeeklyAllowancePool` with `monitoring_meter`:
    - `metric_name` (`"discoveryengine.googleapis.com/quota/agentic_wtu_tier_pp_weekly_regional"`)
    - `official_used_wtu` (`float64`)
    - `log_estimated_wtu` (`float64`)
    - `drift_wtu` (`float64`)
    - `drift_pct` (`float64`)
    - `synced` (`bool`)
  - Surface a **Cloud Monitoring Meter vs. Log Estimate** reconciliation badge in `WeeklyAllowancePoolCard.tsx`, CLI `license`, TUI, and MCP.

### 2.3 Multi-Project Filtering & Top-Navbar Project Selector
- **Configuration & Data (`internal/config/config.go`, `internal/domain/models.go`, `internal/bigquery/demo.go`, `internal/bigquery/client.go`)**:
  - Add `MonitoredProjects []string` (`--monitored-projects` CLI flag / `AGY_COST_BOARD_MONITORED_PROJECTS` env var).
  - Add `ProjectID string` (`project_id`) on `UserUsageLog`, `DailyModelBilling`, and `LicenseSeat`, plus `monitored_projects` (`[]string`) and `selected_project` (`string`) on `OverviewMetrics` and `LicenseGovernanceSummary`.
  - In `--demo` mode, distribute realistic telemetry, billing, and seats across 3 demo projects: `gcp-ai-platform-prod`, `gcp-finops-analytics`, and `gcp-agent-sandbox`.
  - Support `?project=<project_id>` query parameter across `/api/v1/metrics/overview`, `/api/v1/costs/users`, `/api/v1/leaderboard`, and `/api/v1/licenses/status` (where `all` or empty aggregates all monitored projects).
- **Web UI (`Navbar.tsx` & `App.tsx`)**:
  - Add a **Project Selector** dropdown (`All Projects (3)` + individual project IDs) in `Navbar.tsx` that filters all dashboard tabs in real time.

## 3. Acceptance Criteria
- `OverviewMetrics` includes `forecast` (`UsageForecast`), `monitored_projects`, and `selected_project`, and `WeeklyAllowancePool` includes `monitoring_meter`.
- Switching the project selector in `Navbar.tsx` re-queries `/api/v1/*` with `project=<id>` and updates all tabs.
- All Go tests (`go test ./...`) and Web Vitest tests (`npm --prefix web test -- --run && npm --prefix web run build`) pass with zero errors.
