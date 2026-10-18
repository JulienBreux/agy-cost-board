# Implementation Plan: Phase 3 (`v0.12.0`) — 30-Day Window-Aware Forecast, Cloud Monitoring Quota Meter Sync & Multi-Project Support

## Phase 1: Backend Domain, Multi-Project Filtering, Cloud Monitoring Meter & 30-Day Forecast Engine
- [ ] Task: Extend `internal/domain/models.go` and `internal/config/config.go` with `UsageForecast`, `ForecastDay`, `MonitoringQuotaMeter`, and `MonitoredProjects` / `ProjectID` fields
    - [ ] Add unit tests in `internal/domain/models_test.go` and `internal/config/config_test.go`.
    - [ ] Implement structs and configuration parsing (`--monitored-projects`, `AGY_COST_BOARD_MONITORED_PROJECTS`).
- [ ] Task: Update `internal/bigquery/demo.go` and `internal/bigquery/client.go` with multi-project telemetry/billing/seat distribution (`gcp-ai-platform-prod`, `gcp-finops-analytics`, `gcp-agent-sandbox`)
    - [ ] Add unit tests in `internal/bigquery/demo_test.go` verifying multi-project assignment on demo usage logs, billing rows, and license seats.
    - [ ] Update `internal/bigquery/demo.go` and `internal/bigquery/client.go`.
- [ ] Task: Implement 30-Day Window-Aware Forecast, Cloud Monitoring Quota Meter comparison, and Project Filtering (`FilterByProject`) in `internal/attribution/engine.go`
    - [ ] Add unit tests in `internal/attribution/engine_test.go` verifying 30-day window-aware forecast spillover calculation, Cloud Monitoring meter drift calculation, and per-project filtering.
    - [ ] Update `internal/attribution/engine.go` to compute `UsageForecast`, `MonitoringQuotaMeter`, and project-scoped views.
- [ ] Task: Expose `?project=` query parameter, 30-day forecast, and Cloud Monitoring meter across REST API (`internal/server`), CLI (`internal/cli`), TUI (`internal/tui`), and MCP (`internal/mcp`)
    - [ ] Add unit tests in `internal/server/server_test.go`, `internal/cli/formatter_test.go`, and `internal/mcp/server_test.go`.
    - [ ] Update HTTP handlers, CLI formatters, TUI views, and MCP tools.

## Phase 2: Web UI Components (30-Day Forecast Card, Cloud Monitoring Meter Badge & Navbar Project Selector)
- [ ] Task: Update `web/src/api.ts`, create `web/src/components/ForecastCard.tsx`, and add Cloud Monitoring Quota Meter badge to `WeeklyAllowancePoolCard.tsx`
    - [ ] Add unit tests in `web/src/components/ForecastCard.test.tsx` and `web/src/components/WeeklyAllowancePoolCard.test.tsx`.
    - [ ] Implement `ForecastCard.tsx` and update `WeeklyAllowancePoolCard.tsx` and `web/src/api.ts`.
- [ ] Task: Add Multi-Project Selector to `Navbar.tsx` and integrate `ForecastCard` and `selectedProject` state into `App.tsx`
    - [ ] Add unit tests in `web/src/App.test.tsx` verifying project selector filtering and 30-day forecast card rendering.
    - [ ] Update `web/src/components/Navbar.tsx` and `web/src/App.tsx`.
- [ ] Task: Rebuild embedded frontend (`web/dist`), run full Go & Web test suites, and restart `--demo` server on port `8080`
    - [ ] Run `go test ./...` and `npm --prefix web test -- --run && npm --prefix web run build`.
    - [ ] Restart `--demo` server on port `8080` and verify Phase 3 features end-to-end.
