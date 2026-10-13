# Implementation Plan: Native Currency Support Across Engine, CLI, TUI, Statusline, and Web

## Phase 1: Go Domain & Attribution Engine Currency Propagation [checkpoint: 3f09423]
- [x] 8142e1e Task: Domain Currency Helpers & Unit Tests (TDD Red -> Green)
  - Acceptance: `CurrencySymbol` maps ISO-4217 codes (`USD` -> `$`, `EUR` -> `€`, `GBP` -> `£`, `JPY` -> `¥`, `CAD` -> `CA$`, `AUD` -> `A$`, `CHF` -> `CHF `, `INR` -> `₹`) with safe fallback to `CODE `; `FormatCost` formats cost with 2 decimals and currency symbol; `GenerateOptimizationTips` reflects dynamic currency formatting.
  - Verify: `go test -v -race ./internal/domain/...`
  - Files: `internal/domain/currency.go`, `internal/domain/currency_test.go`, `internal/domain/models.go`, `internal/domain/models_test.go`
- [x] 3f09423 Task: Engine Currency Resolution & Tests (TDD Red -> Green)
  - Acceptance: `Engine` inspects `billedCosts` from `FetchBilledCosts` and adopts the billing currency; sets `.Currency` across `AllocatedUserCost`, `UserCostSummary`, `OverviewMetrics`, and `UserConsumptionDriving`; falls back to `"USD"` when billed costs are empty.
  - Verify: `go test -v -race ./internal/attribution/...`
  - Files: `internal/attribution/engine.go`, `internal/attribution/engine_test.go`
- [x] 3f09423 Task: Conductor - User Manual Verification 'Phase 1: Go Domain & Attribution Engine Currency Propagation'

## Phase 2: Go CLI, TUI & Statusline Script Integration
- [x] e812cec Task: CLI & TUI Dynamic Currency Formatting & Tests (TDD Red -> Green)
  - Acceptance: CLI `cost`, `user`, and `license` commands format monetary columns using `FormatCost` / `CurrencySymbol`; TUI `model.go` displays dynamic currency in KPI cards (`TOTAL BILLED`, `POTENTIAL SAVINGS`), cost list, and developer breakdown.
  - Verify: `go test -v -race ./internal/cli/... ./internal/tui/...`
  - Files: `internal/cli/cost.go`, `internal/cli/user.go`, `internal/cli/license.go`, `internal/tui/model.go`, `internal/cli/formatters_test.go`, `internal/tui/tui_test.go`
- [ ] Task: Statusline Script Dynamic Currency & Tests (TDD Red -> Green)
  - Acceptance: `statusline.sh.tmpl` parses `currency` from `/api/v1/users/{id}` API response; maps or passes currency symbol into `COST_FMT` (e.g. ` · cost €18`); tests confirm dynamic currency rendering.
  - Verify: `go test -v -race ./internal/server/...`
  - Files: `internal/server/statusline.sh.tmpl`, `internal/server/statusline_test.go`, `internal/server/statusline_exec_test.go`
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Go CLI, TUI & Statusline Script Integration'

## Phase 3: Web Dashboard Centralized Currency & Component Modernization
- [ ] Task: Web Currency Formatter & Unit Tests (TDD Red -> Green)
  - Acceptance: `web/src/utils/currency.ts` provides `formatCurrency` and `getCurrencySymbol` utilizing `Intl.NumberFormat`; unit tests verify formatting for USD, EUR, GBP, JPY, CAD.
  - Verify: `npm --prefix web test -- src/utils/currency.test.ts --run`
  - Files: `web/src/utils/currency.ts`, `web/src/utils/currency.test.ts`
- [ ] Task: Modernize Overview, Costs, and License Components & Tests (TDD Red -> Green)
  - Acceptance: `KPICards.tsx`, `CostTable.tsx`, `LicenseTable.tsx`, and `CostChart.tsx` use `formatCurrency` / `getCurrencySymbol`; metric toggle in `CostChart` dynamically displays currency symbol instead of static `$`; component tests verify correct currency rendering.
  - Verify: `npm --prefix web test -- src/components/CostChart.test.tsx src/components/KPICards.test.tsx src/App.test.tsx --run`
  - Files: `web/src/components/KPICards.tsx`, `web/src/components/CostTable.tsx`, `web/src/components/LicenseTable.tsx`, `web/src/components/CostChart.tsx`, `web/src/App.tsx`
- [ ] Task: Modernize Personal Consumption, Driving & Modal Components & Tests (TDD Red -> Green)
  - Acceptance: `PersonalCostChart.tsx`, `PersonalKPICards.tsx`, `BudgetProgressBar.tsx`, `OptimizationCard.tsx`, `RecentActivityTable.tsx`, `StatuslineHelpBox.tsx`, and `UserModal.tsx` use centralized currency formatting; metric toggle in `PersonalCostChart` displays dynamic currency symbol; all tests pass.
  - Verify: `npm --prefix web test -- --run`
  - Files: `web/src/components/PersonalCostChart.tsx`, `web/src/components/PersonalKPICards.tsx`, `web/src/components/BudgetProgressBar.tsx`, `web/src/components/OptimizationCard.tsx`, `web/src/components/RecentActivityTable.tsx`, `web/src/components/StatuslineHelpBox.tsx`, `web/src/components/UserModal.tsx`
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Web Dashboard Centralized Currency & Component Modernization'

## Phase 4: Full Verification, Lint & End-to-End Validation
- [ ] Task: Full Suite Verification & Demo Mode E2E Test
  - Acceptance: All Go tests and Web tests pass; `golangci-lint` passes without errors; demo mode running with EUR correctly renders `€` in both CLI (`go run ./cmd/agy-cost-board cost --demo`) and Web dashboard.
  - Verify: `make test && make lint && npm --prefix web run build`
  - Files: `tests/e2e_test.go`, `README.md`
- [ ] Task: Conductor - User Manual Verification 'Phase 4: Full Verification, Lint & End-to-End Validation'
