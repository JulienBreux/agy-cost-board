# Implementation Plan: Phase 1 (`v0.10.0`) — Token Cache, Thinking Effort, Client Telemetry & Included vs. PAYG Coexistence

## Phase 1: Domain Model, WTU/PAYG Attribution Engine & Demo Telemetry [checkpoint: 8a25fcd]
- [x] Task: Extend Domain Models & BigQuery/Demo Telemetry for Cached Tokens, QuotaType, EffortLevel, and Clients cdb84c9
  - [x] Write failing unit tests in `internal/domain/models_test.go` and `internal/bigquery/demo_test.go` for `CachedTokens`, `QuotaType`, `EffortLevel`, `ClientName`, `ClientVersion`, `WTU`, `IncludedCost`, `PAYGCost`, `TokenCostShare`, `ModelEffortStats`, and `ClientStats`
  - [x] Implement domain struct updates in `internal/domain/models.go` and enrich BigQuery queries (`internal/bigquery/client.go`) and synthetic demo data (`internal/bigquery/demo.go`)
- [x] Task: Implement WTU Metering, Included vs. PAYG Coexistence, and Token/Effort/Client Aggregations in Attribution Engine 76b65e5
  - [x] Write failing unit tests in `internal/attribution/engine_test.go` verifying WTU math (50-WTU minimum, Flash/Argon 0.5x meter share, Claude 100% PAYG, `REGULAR_QUOTA` vs `OVERAGE_QUOTA`), Token Share vs. Cost Share (Uncached Input, Cached Input, Output, Thinking), Model+Effort breakdown, Client breakdown, and per-user Cache Hit Rate %
  - [x] Implement WTU/PAYG calculation and aggregation logic in `internal/attribution/engine.go`
- [x] Task: Expose New FinOps Fields Across CLI, TUI, and MCP Tools 8a25fcd
  - [x] Write failing unit tests in `internal/cli/formatters_test.go`, `internal/tui/tui_test.go`, and `internal/mcp/server_test.go` for Included/PAYG cost, Cache Hit Rate %, Effort, and Client metrics
  - [x] Update CLI formatters (`internal/cli/formatters.go`), TUI views (`internal/tui/model.go`), and MCP server tool payloads (`internal/mcp/server.go`)
- [x] Task: Conductor - User Manual Verification 'Phase 1: Domain Model, WTU/PAYG Attribution Engine & Demo Telemetry' (Protocol in workflow.md) 8a25fcd

## Phase 2: Web UI Coexistence (Overview, Tokens/Effort/Clients Section, Leaderboard & Personal Dashboard) [checkpoint: 00fdb08]
- [x] Task: Update Frontend API Types, KPI Cards, and Daily Cost Chart (By Model vs. Included/PAYG Toggle) 8a4d9cf
  - [x] Write failing tests in `web/src/api.test.ts`, `web/src/components/KPICards.test.tsx`, and `web/src/components/CostChart.test.tsx` for Included vs. PAYG coexistence and Cache Hit Rate %
  - [x] Update `web/src/api.ts`, `web/src/components/KPICards.tsx`, and `web/src/components/CostChart.tsx`
- [x] Task: Create `TokensEffortClientSection` Component ("Where Tokens & Money Go", "Models & Thinking Effort", and "Clients") 6792565
  - [x] Write failing unit tests in `web/src/components/TokensEffortClientSection.test.tsx` and `web/src/App.test.tsx`
  - [x] Implement `web/src/components/TokensEffortClientSection.tsx` and integrate it into the Overview tab in `web/src/App.tsx`
- [x] Task: Enhance Leaderboard, Attributed Costs Table, and My Consumption Dashboard with Included/PAYG, Cache Hit %, and Effort Optimization Insights 00fdb08
  - [x] Write failing unit tests in `web/src/components/LeaderboardView.test.tsx`, `web/src/components/PersonalKPICards.test.tsx`, and `web/src/components/OptimizationCard.test.tsx`
  - [x] Update `web/src/components/LeaderboardView.tsx`, `web/src/components/CostTable.tsx`, `web/src/components/PersonalKPICards.tsx`, and `web/src/components/OptimizationCard.tsx`
- [x] Task: Conductor - User Manual Verification 'Phase 2: Web UI Coexistence (Overview, Tokens/Effort/Clients Section, Leaderboard & Personal Dashboard)' (Protocol in workflow.md) 00fdb08

