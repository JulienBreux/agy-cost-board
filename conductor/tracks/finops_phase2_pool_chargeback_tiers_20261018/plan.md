# Implementation Plan: Phase 2 (`v0.11.0`) — Weekly Allowance Pool Engine, Fair-Share vs. When-Used Chargeback & Multi-Tier License Discovery

## Phase 1: Domain Models, Weekly Allowance Pool Math & Multi-Tier Demo/BigQuery Data
- [ ] Task: Extend `internal/domain/models.go` with Multi-Tier `LicenseSeat` fields, `WeeklyAllowancePool`, `WeeklyPoolWindow`, `PoolWhatIfScenario`, and `FairShare` / `WhenUsed` PAYG fields
    - [ ] Add unit tests in `internal/domain/models_test.go` covering `Standard` (`$2.50/wk` = `1.25M WTU`) vs `Plus` (`$3.75/wk` = `1.875M WTU`) seat allowance helpers and CSV export rows.
    - [ ] Implement structs and helper functions in `internal/domain/models.go`.
- [ ] Task: Enhance `internal/bigquery/client.go` and `internal/bigquery/demo.go` with Multi-Tier (`standard` / `plus`), Multi-Location (`global`, `eu`, `us`), and Trial seat metadata
    - [ ] Add unit tests in `internal/bigquery/demo_test.go` verifying tier, location, trial status, and weekly allowance fields on demo seats.
    - [ ] Update `internal/bigquery/demo.go` and `internal/bigquery/client.go` to populate `Tier`, `WeeklyAllowanceUSD`, `WeeklyAllowanceWTU`, `Location`, `IsTrial`, and `TrialEndsAt`.
- [ ] Task: Implement Weekly Allowance Pool & What-If simulation (`pooled` vs `per_user`, `purchased` vs `assigned`) and 7-day window Fair Share vs. When Used PAYG attribution in `internal/attribution/engine.go`
    - [ ] Add unit tests in `internal/attribution/engine_test.go` verifying 7-day pool window calculation, What-If scenario matrix (`pooled_assigned`, `pooled_purchased`, `per_user_assigned`, `per_user_purchased`), and Fair Share vs. When Used PAYG overage distribution.
    - [ ] Update `internal/attribution/engine.go` to compute `WeeklyAllowancePool` and populate both `FairShare` and `WhenUsed` fields on `AllocatedUserCost`, `LeaderboardEntry`, `UserCostSummary`, `OverviewMetrics`, and `LicenseGovernanceSummary`.
- [ ] Task: Expose Phase 2 query parameters and metrics in REST API (`internal/server`), CLI (`cmd/agy-cost-board`), TUI (`internal/tui`), and MCP (`internal/mcp`)
    - [ ] Add unit tests in `internal/server/server_test.go`, `cmd/agy-cost-board/cli_test.go`, `internal/tui/model_test.go`, and `internal/mcp/server_test.go`.
    - [ ] Update HTTP handlers (`/api/costs`, `/api/leaderboard`, `/api/licenses`, `/api/overview`), CLI formatters, TUI views, and MCP tools.

## Phase 2: Web UI Components (Weekly Allowance Pool Card, Fair Share Toggle, CSV Export & Multi-Tier Seat Governance)
- [ ] Task: Update `web/src/api.ts` and create `web/src/components/WeeklyAllowancePoolCard.tsx` with What-If Simulator toggles
    - [ ] Add unit tests in `web/src/components/WeeklyAllowancePoolCard.test.tsx` verifying 7-day drawdown bar, reset countdown, weekly window history, and interactive `Pooled` / `Per-User` and `Purchased` / `Assigned` toggles.
    - [ ] Implement `WeeklyAllowancePoolCard.tsx` and update `web/src/api.ts`.
- [ ] Task: Integrate `WeeklyAllowancePoolCard` into `Overview` (`App.tsx`) and `Seat Governance` (`LicenseTable.tsx`), and add Multi-Tier (`Standard` / `Plus`), Location (`global`, `eu`, `us`), and `Free Trial` badges to `LicenseTable.tsx`
    - [ ] Add unit tests in `web/src/components/LicenseTable.test.tsx` and `web/src/App.test.tsx`.
    - [ ] Update `web/src/components/LicenseTable.tsx` and `web/src/App.tsx`.
- [ ] Task: Add Fair Share / When Used segmented toggle and Export CSV button to `LeaderboardView.tsx` and `CostTable.tsx`
    - [ ] Add unit tests in `web/src/components/LeaderboardView.test.tsx` and `web/src/components/CostTable.test.tsx` verifying mode switching between `Fair Share` and `When Used` and CSV export triggers.
    - [ ] Update `web/src/components/LeaderboardView.tsx` and `web/src/components/CostTable.tsx`.
- [ ] Task: Rebuild embedded frontend (`web/dist`), run full test suites, and verify in `--demo` mode on port `8080`
    - [ ] Run `go test ./...` and `npm --prefix web test -- --run && npm --prefix web run build`.
    - [ ] Restart `--demo` server on port `8080` and verify all Phase 2 features end-to-end.
