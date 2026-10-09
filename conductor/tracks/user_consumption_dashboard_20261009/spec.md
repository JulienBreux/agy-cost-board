# Specification: User Personal Consumption Dashboard ('My Consumption')

## 1. Overview
Introduce a dedicated **"My Consumption"** tab in the `agy-cost-board` web dashboard designed specifically for individual developers to inspect, monitor, and actively drive their Antigravity and Gemini Enterprise consumption. The dashboard empowers users to track their individual cost burn rates, analyze token shares across models, manage personal budgets and quotas, receive actionable efficiency recommendations, and view recent inference activity with live telemetry updates.

## 2. Functional Requirements

### 2.1 Navigation & Tab Architecture
- Add a new **"My Consumption"** tab in the main navigation bar alongside `Overview`, `Costs`, `Licenses`, and `Setup`.
- Ensure seamless URL deep-linking (e.g. `/my-consumption` or `?tab=my-consumption&user=<userId>`).
- Enhance/expand the existing `UserModal` so that clicking a user row in Costs or Licenses can navigate directly to or pre-fill the "My Consumption" dashboard view.

### 2.2 User Identity & Resolution
- **Identity Auto-Detection:** Automatically detect the logged-in user identity from GCP Cloud Run authentication headers (`X-Goog-Authenticated-User-Email` or `X-Forwarded-Email`) when deployed behind IAP / Cloud Run IAM.
- **Searchable User Selector:** Provide a searchable dropdown allowing users to select or switch identities (with debounced search across all known active user IDs).
- **Persistence:** Persist the chosen or detected user ID in browser `localStorage` as default for subsequent sessions.
- **Demo / Fallback:** In `--demo` or unauthenticated local mode, default to the most active demo user while allowing instant selector switching.

### 2.3 Personal Spend & Token Analytics
- **Summary KPI Cards:**
  - Total Spend (USD) in selected period and month-to-date (MTD).
  - Total Tokens consumed (Input, Output, Cached).
  - Daily & Weekly average burn rate with trend delta (% change vs previous period).
  - Relative share of organization spend (e.g., "Responsible for 4.2% of total team AI spend").
- **Cost & Token Trajectory Charts:**
  - Daily spend stacked by model (e.g., Gemini 1.5 Pro vs. Gemini 1.5 Flash).
  - Token distribution comparison chart highlighting input vs output vs cache hits.

### 2.4 Budget & Quota Driving
- **Personal Budget Targets:** Configurable personal monthly budget threshold (persisted locally or read from config/user limits).
- **Burn Rate Projections:** End-of-month projected spend based on current daily velocity and remaining working days.
- **Visual Gauge / Progress Bar:** Color-coded budget consumption indicator (Green < 75%, Amber 75-90%, Red > 90% or overrun).
- **Proactive Alerts:** Visual warnings when current burn rate is on pace to exceed personal or team-allocated quota.

### 2.5 Efficiency & Optimization Insights
- **Smart Recommendations Engine:**
  - Model tier optimization: Identify prompts or workloads heavily using Pro that could leverage Flash.
  - Context caching opportunities: Surface repetitive large system instructions or file contexts that could leverage Gemini context caching.
  - License status badge: Active Gemini Enterprise license indicator with activity status (active, idle, or unassigned).
- **Estimated Savings:** Potential USD or token savings if recommendations are adopted.

### 2.6 Live Telemetry & Recent Activity Log
- **Live Indicator:** Real-time pulse / status badge showing "Live" status and last telemetry sync timestamp.
- **Manual & Auto-Refresh:** Quick refresh button with optional 60-second polling for live day-of activity.
- **Recent Inference Activity Table:**
  - Timestamp, Model (e.g. Gemini 1.5 Pro), Total Tokens (Prompt + Response), and estimated cost per request.
  - Filtering by model and status.

### 2.7 Time Range Filtering
- Synchronized with global navbar date filter (`7d`, `30d`, `90d`) plus a 1-click **Month-to-Date (MTD)** toggle.

## 3. Non-Functional Requirements
- **Performance:** Instant client-side filtering and efficient backend endpoints (< 250ms query response on cached BigQuery datasets).
- **Theme & Design Consistency:** Full compliance with existing Google Cloud dark/light palette, Tailwind CSS, Lucide icons, and Recharts visuals.
- **Single-Binary Dual Mode:** Fully embedded in the single Go binary via Go `embed.FS` with zero runtime external asset dependencies.
- **Graceful Error States:** Clear empty/loading states when a user has zero recorded inference logs in the selected timeframe.

## 4. Acceptance Criteria
- [x] Navbar includes "My Consumption" tab and persists state across page reloads.
- [x] Cloud Run IAP header auto-populates user identity, with fallback to searchable dropdown and localStorage.
- [x] Personal KPIs, burn rate charts, and model distributions render accurately from API.
- [x] Budget bar and projected month-end spend calculations function dynamically.
- [x] Optimization insights card generates actionable recommendations based on user token patterns.
- [x] Recent inference activity log renders user requests with timestamps, models, and token counts.
- [x] Live sync indicator and refresh button successfully refresh telemetry.
- [x] Unit and component tests pass with >80% coverage.

## 5. Out of Scope
- Administrative modification of GCP Cloud Billing budgets or IAM permissions from the UI.
- Direct billing chargeback invoicing to credit cards.
