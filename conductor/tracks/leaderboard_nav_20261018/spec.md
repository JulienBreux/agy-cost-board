# Specification: Developer Leaderboard & Two-Line Navigation Bar

## 1. Overview
Introduce a comprehensive **Developer Leaderboard** across the Web Dashboard, CLI (`agy-cost-board leaderboard`), and interactive TUI, while refining the Web Dashboard header into a clean **two-line navigation layout**.

## 2. Functional Requirements

### 2.1 Two-Line Web Navigation (`Navbar.tsx` & `App.tsx`)
- **First Line (Top Header Bar)**:
  - Left: Brand identity (`Usage and Spend` / `Gemini platform & Antigravity`).
  - Right: Global utility controls:
    - Time window selector (`7d`, `14d`, `30d`)
    - `MCP & Quotas` status/modal button
    - `What's new` changelog modal button
    - Theme toggle (Dark / Light mode)
- **Second Line (Secondary Navigation Bar)**:
  - Dedicated horizontal navigation bar (`aria-label="Main Navigation"`) containing the view tabs in order:
    1. **Overview** (`overview`)
    2. **Leaderboard** (`leaderboard`) — with `Trophy` icon
    3. **My Consumption** (`my-consumption`)
    4. **Attributed Costs** (`costs`)
    5. **Seat Governance** (`licenses`)
    6. **Setup & Health** (`setup`)
  - URL query synchronization (`?tab=leaderboard`) so the Leaderboard view is deep-linkable and bookmarkable.

### 2.2 Leaderboard Domain & Aggregation (`internal/domain`, `internal/attribution`)
- Define `LeaderboardEntry` in `internal/domain/models.go`:
  - `Rank` (`int` `json:"rank"`)
  - `UserID` (`string` `json:"user_id"`)
  - `TotalCost` (`float64` `json:"total_cost"`) & `ShareOfSpend` (`float64` `json:"share_of_spend"`)
  - `TotalTokens` (`int64` `json:"total_tokens"`) & `ShareOfTokens` (`float64` `json:"share_of_tokens"`)
  - `ActiveDays` (`int` `json:"active_days"`)
  - `CostPer1KTokens` (`float64` `json:"cost_per_1k_tokens"`)
  - `TopModel` (`string` `json:"top_model"`) & `ModelBreakdown` (`map[string]float64` `json:"model_breakdown"`)
  - `Currency` (`string` `json:"currency"`)
- Expose `GetLeaderboard(ctx, days)` on `attribution.Engine` and `GET /api/v1/leaderboard?days=30` on the REST API.

### 2.3 Web Leaderboard View (`web/src/components/LeaderboardView.tsx`)
- **Top 3 Podium Cards**:
  - Highlight #1 (Gold), #2 (Silver), and #3 (Bronze) developers with medal badges, total attributed spend (formatted in native currency), total tokens, active days, and top model.
  - Clicking a podium card opens the developer drill-down `UserModal`.
- **Ranked Leaderboard Table**:
  - Search filter by developer `user_id`.
  - Sortable columns: **Rank / Total Spend** (default descending), **Total Tokens**, **Active Days**, and **Efficiency (Cost / 1k Tokens)**.
  - Visual spend share bar and stacked model mix progress bar per developer.
  - Clicking any developer row opens `UserModal` (with shortcut to `My Consumption`).

### 2.4 CLI (`agy-cost-board leaderboard`) & Interactive TUI Tab
- **CLI Subcommand (`internal/cli/leaderboard.go`)**:
  - `agy-cost-board leaderboard [--days=30] [--format=table|json|csv]`
  - Displays rank (`#1`, `#2`, ...), developer email, total cost (with native currency symbol), spend share %, total tokens, active days, cost/1k tokens, and top model.
- **TUI Tab (`internal/tui`)**:
  - Add a **Leaderboard** tab to the Bubbletea terminal UI (`[1] Overview`, `[2] Leaderboard`, `[3] Costs`, `[4] Licenses`, `[5] User Breakdown`) with keyboard sorting and `Enter` drill-down.

## 3. Non-Functional Requirements
- **Zero Regression & High Test Coverage**: >80% unit test coverage across Go (`internal/attribution`, `internal/cli`, `internal/server`, `internal/tui`) and React (`Navbar.test.tsx`, `LeaderboardView.test.tsx`, `App.test.tsx`).
- **Responsive & Theme-Aware**: Seamless dark/light mode support and horizontal scroll resilience on mobile viewports.

## 4. Acceptance Criteria
1. Web header renders Brand + Global Controls on Line 1, and `Overview`, `Leaderboard`, `My Consumption`, `Attributed Costs`, `Seat Governance`, `Setup & Health` on Line 2.
2. Web `Leaderboard` tab displays the Top 3 Podium and sortable/searchable ranked table with model mix bars and user drill-down.
3. `agy-cost-board leaderboard --demo` outputs the ranked developer leaderboard in `table`, `json`, and `csv` formats.
4. `agy-cost-board tui --demo` includes the interactive Leaderboard tab.
5. All unit, E2E, and frontend tests pass with zero lint warnings.
