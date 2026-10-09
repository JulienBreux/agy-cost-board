# Implementation Plan: User Personal Consumption Dashboard ('My Consumption')

## Phase 1: Backend APIs & Telemetry Endpoints (TDD)
- [ ] Task: Backend User Activity & Identity Endpoints
    - [ ] Write unit tests for `/api/v1/me` identity detection and `/api/v1/users/{id}/activity`
    - [ ] Implement identity detection middleware/handler for `X-Goog-Authenticated-User-Email`
    - [ ] Implement `/api/v1/users/{id}/activity` returning recent telemetry logs, token counts, and cost estimates
- [ ] Task: Budget Projection & Optimization Insights Logic
    - [ ] Write unit tests for burn rate velocity, month-end projection, and recommendation algorithms
    - [ ] Implement budget calculation and rule-based optimization insights (Pro vs Flash, caching)
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Backend APIs & Telemetry Endpoints' (Protocol in workflow.md)

## Phase 2: Frontend Data Client & Identity Selection (TDD)
- [ ] Task: API Client & Identity Persistence
    - [ ] Write unit tests for personal activity fetching, identity resolution, and localStorage sync
    - [ ] Implement `fetchUserActivity`, `fetchCurrentUser`, and `fetchUserBudgets` in `web/src/api.ts`
- [ ] Task: Searchable User Selector Component
    - [ ] Write component tests for `UserSelector` with debounced search and keyboard selection
    - [ ] Implement `UserSelector` component with persistence and clear empty states
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Frontend Data Client & Identity Selection' (Protocol in workflow.md)

## Phase 3: Personal Consumption & Driving UI Components (TDD)
- [ ] Task: Personal KPI Cards & Budget Progress Gauge
    - [ ] Write component tests for `PersonalKPICards` and `BudgetProgressBar`
    - [ ] Implement `PersonalKPICards` (MTD spend, daily/weekly burn rate, org share) and `BudgetProgressBar` with projection alerts
- [ ] Task: Spend Trajectory & Model Distribution Charts
    - [ ] Write component tests for `PersonalCostChart` and `TokenDistributionChart`
    - [ ] Implement Recharts visualizations for user daily spend by model and token type breakdowns
- [ ] Task: Actionable Optimization Insights Card
    - [ ] Write component tests for `OptimizationCard` with model switching and caching tips
    - [ ] Implement `OptimizationCard` surfacing actionable recommendations and estimated savings
- [ ] Task: Recent Activity Log & Live Sync Indicator
    - [ ] Write component tests for `RecentActivityTable` and `LiveSyncBadge`
    - [ ] Implement live telemetry activity table with model/date filtering and real-time pulse badge
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Personal Consumption & Driving UI Components' (Protocol in workflow.md)

## Phase 4: Full View Assembly, Navigation & Verification
- [ ] Task: Navigation Bar & App Assembly
    - [ ] Add "My Consumption" tab to `Navbar.tsx` and integrate `UserDashboardView.tsx` into `App.tsx`
    - [ ] Update `UserModal` with a quick link to open full user dashboard in "My Consumption" tab
    - [ ] Support URL query parameters (`?tab=my-consumption&user=...`)
- [ ] Task: End-to-End Validation & Verification
    - [ ] Write integration tests verifying tab switching, user identity resolution, and live data refresh
    - [ ] Run full test suite (`go test ./...`, frontend tests/build) and verify zero linter warnings
- [ ] Task: Conductor - User Manual Verification 'Phase 4: Full View Assembly, Navigation & Verification' (Protocol in workflow.md)
