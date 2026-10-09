# Implementation Plan: User Personal Consumption Dashboard ('My Consumption')

## Phase 1: Backend APIs & Telemetry Endpoints (TDD)
- [x] Task: Backend User Activity & Identity Endpoints
    - [x] Write unit tests for `/api/v1/me` identity detection and `/api/v1/users/{id}/activity`
    - [x] Implement identity detection middleware/handler for `X-Goog-Authenticated-User-Email`
    - [x] Implement `/api/v1/users/{id}/activity` returning recent telemetry logs, token counts, and cost estimates
- [x] Task: Budget Projection & Optimization Insights Logic
    - [x] Implement budget calculation and rule-based optimization insights (Pro vs Flash, caching)
- [x] Task: Conductor - User Manual Verification 'Phase 1: Backend APIs & Telemetry Endpoints' (Protocol in workflow.md) [1dac6a7]

## Phase 2: Frontend Data Client & Identity Selection (TDD)
- [x] Task: API Client & Identity Persistence [32b7bee]
    - [x] Write unit tests for personal activity fetching, identity resolution, and localStorage sync
    - [x] Implement `fetchUserActivity`, `fetchCurrentUser`, and `fetchUserBudgets` in `web/src/api.ts`
- [x] Task: Searchable User Selector Component [af02613]
    - [x] Write component tests for `UserSelector` with debounced search and keyboard selection
    - [x] Implement `UserSelector` component with persistence and clear empty states
- [x] Task: Conductor - User Manual Verification 'Phase 2: Frontend Data Client & Identity Selection' (Protocol in workflow.md) [af02613]

## Phase 3: Personal Consumption & Driving UI Components (TDD)
- [x] Task: Personal KPI Cards & Budget Progress Gauge [f7724d2]
    - [x] Write component tests for `PersonalKPICards` and `BudgetProgressBar`
    - [x] Implement `PersonalKPICards` (MTD spend, daily/weekly burn rate, org share) and `BudgetProgressBar` with projection alerts
- [x] Task: Spend Trajectory & Model Distribution Charts [ff0f386]
    - [x] Write component tests for `PersonalCostChart`
    - [x] Implement lightweight SVG visualization for user daily spend and model breakdown
- [x] Task: Actionable Optimization Insights Card [cc33f72]
    - [x] Write component tests for `OptimizationCard` with model switching and caching tips
    - [x] Implement `OptimizationCard` surfacing actionable recommendations and estimated savings
- [x] Task: Recent Activity Log & Live Sync Indicator [5f393ab]
    - [x] Write component tests for `RecentActivityTable` and `LiveSyncBadge`
    - [x] Implement live telemetry activity table with model/date filtering and real-time pulse badge
- [x] Task: Conductor - User Manual Verification 'Phase 3: Personal Consumption & Driving UI Components' (Protocol in workflow.md) [5f393ab]

## Phase 4: Full View Assembly, Navigation & Verification
- [ ] Task: Navigation Bar & App Assembly
    - [ ] Add "My Consumption" tab to `Navbar.tsx` and integrate `UserDashboardView.tsx` into `App.tsx`
    - [ ] Update `UserModal` with a quick link to open full user dashboard in "My Consumption" tab
    - [ ] Support URL query parameters (`?tab=my-consumption&user=...`)
- [ ] Task: End-to-End Validation & Verification
    - [ ] Write integration tests verifying tab switching, user identity resolution, and live data refresh
    - [ ] Run full test suite (`go test ./...`, frontend tests/build) and verify zero linter warnings
- [ ] Task: Conductor - User Manual Verification 'Phase 4: Full View Assembly, Navigation & Verification' (Protocol in workflow.md)
