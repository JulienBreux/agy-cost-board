# Architecture & FinOps Model

`agy-cost-board` combines **reconciled Google Cloud Billing attribution** with **Gemini Enterprise Weighted Token Unit (WTU) metering**, **7-day floating Weekly Allowance Pools**, **Fair-Share vs. When-Used PAYG chargeback**, and **30-day window-aware spillover forecasting**.

---

## 1. Mathematical Proportional Attribution (Cloud Billing Reconciliation)

When enterprises adopt AI developer tools like Antigravity and Gemini Enterprise, standard Google Cloud invoices pool all model API consumption into aggregate SKU line items.

Based on the Google Cloud architecture pattern [*"Per-user cost attribution for Antigravity with BigQuery"*](https://medium.com/google-cloud/per-user-cost-attribution-for-antigravity-with-bigquery-3e98fd997c58), `agy-cost-board` joins Cloud Logging inference records (`InferenceResponseLog`) with Cloud Billing exports (`gcp_billing_export_v1_*`) by date and model. Each developer $u$ using model $m$ on day $d$ is allocated:

$$\text{Allocated Cost}(u, m, d) = \text{Billed Cost}(m, d) \times \frac{\text{Tokens}(u, m, d)}{\sum_{k} \text{Tokens}(k, m, d)}$$

- **Zero Unallocated Residue**: 100% of the daily model SKU invoice is distributed across active callers on that day.
- **Heuristic Fallback**: If Cloud Billing exports are delayed or absent for a given day/model, `agy-cost-board` falls back cleanly to token/WTU rate estimation until the billing export lands.

---

## 2. Dual Cost Model Coexistence: Reconciled Billing + WTU & Included vs. PAYG (`v0.10.0+`)

In Gemini Enterprise, Antigravity usage is governed by **Weighted Token Units (WTU)** (`$2.00 per 1,000,000 WTU`, with a minimum of `50 WTU` per inference call) that draw down from a floating 7-day **Weekly Allowance Pool** before triggering **Pay-As-You-Go (PAYG) Overage**:

| Dimension | Metric | Description |
| :--- | :--- | :--- |
| **Total Reconciled Cost** | `total_billed_cost` | Exact invoice spend reconciled from Cloud Billing exports (`gcp_billing_export_v1_*`). |
| **Weighted Token Units** | `total_wtu` | Metered WTU accounting for model meter shares (`0.5x` promotional rate on `gemini-3.6/3.7/3.8-flash` & `gemini-4-argon`, `1.0x` standard), cached input discounts, `>200k` long-context multipliers, and the `50 WTU` call floor. |
| **Included Quota Value** | `included_cost` | Dollar value (`$2 / 1M WTU`) of usage absorbed by the Gemini Enterprise seat allowance (`REGULAR_QUOTA`). |
| **PAYG Overage Spend** | `payg_cost` | Incremental usage billed on top of seat subscriptions (`OVERAGE_QUOTA`, unlicensed callers, or third-party models like Anthropic Claude that bypass the Gemini pool). |

---

## 3. Floating 7-Day Weekly Allowance Pool & What-If Simulator (`v0.11.0+`)

Each Gemini Enterprise seat contributes a weekly WTU allowance to a 7-day rolling window:

| Subscription Tier | Discovery Engine Tier Enum | Monthly Seat Reference | Weekly Included Pool Allowance | Weekly WTU Contribution |
| :--- | :--- | :--- | :--- | :--- |
| **Standard** | `SUBSCRIPTION_TIER_STANDARD` / `SUBSCRIPTION_TIER_ENTERPRISE` | `$10.00 / mo` | **`$2.50 / week`** | `1,250,000 WTU / week` |
| **Plus** | `SUBSCRIPTION_TIER_PLUS` / `SUBSCRIPTION_TIER_SEARCH_AND_ASSISTANT` | `$15.00 / mo` | **`$3.75 / week`** | `1,875,000 WTU / week` |

### Interactive What-If Pool Controls
Both the backend engine (`internal/attribution`) and the Web UI (`WeeklyAllowancePoolCard`) let you simulate pool behavior across two axes:
1. **Pool Mode (`--pool-mode`)**:
   - `pooled` *(default)*: All seats combine into a shared organization/project weekly WTU pool.
   - `per_user`: Each developer's usage is capped against their own individual weekly seat allowance (`1.25M` or `1.875M WTU`).
2. **Seat Count Basis (`--pool-basis`)**:
   - `assigned` *(default)*: Weekly pool capacity is calculated only from currently assigned seats (`Standard` + `Plus`).
   - `purchased`: Weekly pool capacity includes unassigned seats in your total `seat_quota`.

---

## 4. Fair-Share vs. When-Used PAYG Chargeback (`v0.11.0+`)

When an organization's shared weekly pool runs out (e.g., on Thursday afternoon), chronological (`OVERAGE_QUOTA`) logs mark all subsequent Friday calls as PAYG overage—unfairly penalizing developers whose work happened later in the week.

`agy-cost-board` computes both chargeback methodologies simultaneously and lets you toggle between them in the **Attributed Costs** and **Leaderboard** views (or via `--payg-mode`):

- **Fair Share (`fair_share` — Default for FinOps Chargeback)**:
  - Spreads a window's pool-exhaustion overage across all licensed Gemini callers in proportion to their total WTU share during that 7-day window.
  - Note: Calls by **unlicensed users** or to **non-pool models (Anthropic Claude)** always remain 100% attributed to the caller who made them.
- **When Used (`when_used` — Chronological)**:
  - Attributes PAYG overage strictly to the specific inference calls that occurred while the weekly pool was exhausted (`OVERAGE_QUOTA`).

---

## 5. Cloud Monitoring Quota Meter Reconciliation (`v0.12.0+`)

Log-derived WTU estimates from BigQuery (`InferenceResponseLog`) can occasionally diverge slightly from Google Cloud's internal quota enforcement meter due to sink latency or regional quota bucketing.

`agy-cost-board` surfaces a **Cloud Monitoring Quota Meter vs. Log Estimate** reconciliation banner (`monitoring_meter`) inside the Weekly Allowance Pool card:
- **`metric_name`**: `discoveryengine.googleapis.com/agentic_wtu_tier_plus_pp_weekly_regional`
- **`metered_wtu` vs. `log_estimated_wtu`**: Compares the Cloud Monitoring meter reading against BigQuery log-derived WTU.
- **`drift_wtu` & `drift_pct`**: Highlights exact variance and flags `synced` ($\le 5\%$ drift) vs. `drift_detected` ($> 5\%$ drift).

---

## 6. 30-Day Window-Aware Usage & Spend Forecast (`v0.12.0+`)

Linear burn-rate projections fail when usage is buffered by resetting 7-day allowance pools. `agy-cost-board` implements a **30-Day Window-Aware Forecast** (`UsageForecast`):
1. Computes trailing 7-day daily WTU velocity (`avg_daily_wtu`), daily PAYG velocity (`avg_daily_payg_cost`), and daily total cost velocity (`avg_daily_total_cost`).
2. Compares daily WTU velocity against the **daily pool capacity threshold** (`weekly_pool_limit_wtu / 7`).
3. Projects 30 forward days (`daily_projections`), splitting each future day into **Included Pool WTU** vs. **Spillover PAYG WTU** and estimating `projected_30d_spillover_cost` and `projected_30d_total_cost`.

---

## 7. Token Cache Efficiency, Thinking Effort & Client Telemetry (`v0.10.0+`)

### Where Tokens and Money Go (Token Share vs. Cost Share)
By parsing `cachedContentTokenCount` alongside `promptTokenCount`, `candidatesTokenCount`, and `thoughtsTokenCount`, `agy-cost-board` contrasts **Token Share %** against **Cost Share %** across four categories:
1. **Uncached Input**
2. **Cached Input** (often 40%–75%+ of total tokens at ~10% of input cost)
3. **Output (Candidates)**
4. **Thinking (Reasoning)**

### Models & Thinking Effort Levels (`minimal` .. `max`)
`agy-cost-board` extracts the reasoning effort suffix from `jsonPayload.experience` (`minimal`, `low`, `medium`, `high`, `xhigh`, `max`, or `default`). The **Models & Thinking Effort** table and personal **Optimization Card** highlight when high-effort (`high`/`xhigh`/`max`) reasoning calls drive disproportionate thinking token spend.

### Antigravity Client Breakdown
Tracks `labels.client_name` and `labels.client_version` (`vscode`, `jetbrains`, `agy-cli`) so platform teams can see which IDE or CLI versions drive adoption and token volume.

---

## 8. Gemini Enterprise Shelfware Reclamation & Multi-Location Discovery

`agy-cost-board` discovers seats across configured locations (`global`, `eu`, `us`) and classifies every assigned license over your selected lookback window (`7d`, `14d`, `30d`, `90d`):
- **Active**: Regular developer engagement ($\ge 1,000$ tokens).
- **At-Risk**: Minimal engagement ($1\text{–}999$ tokens); candidate for onboarding coaching.
- **Dormant**: Zero token activity ($0$ tokens); candidate for immediate seat reclamation (`estimated_monthly_savings`).
- **Trial & Lifecycle Tracking**: Tracks `tier` (`standard` vs. `plus`), `location`, `free_trial`, `start_date`, and `end_date`.

---

## 9. Native ISO-4217 Multi-Currency (Zero FX Loss)

Google Cloud invoices organizations in their contracted local billing currency (`USD`, `EUR`, `GBP`, `JPY`, `CAD`, `CHF`, `AUD`, etc.). `agy-cost-board` ingests the ISO-4217 `currency` code directly from the BigQuery Cloud Billing export and formats all CLI tables, TUI views, Statusline outputs, MCP responses, and React UI components natively—**never applying lossy foreign exchange (FX) conversions**.
