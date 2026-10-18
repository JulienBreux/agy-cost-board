# Implementation Plan: Developer Leaderboard & Two-Line Navigation Bar

## Phase 1: Domain Model, Attribution Engine & REST API (`internal/domain`, `internal/attribution`, `internal/server`)
- [ ] Task: Define `LeaderboardEntry` model and implement `GetLeaderboard` in `internal/attribution`
    - [ ] Write unit tests in `internal/attribution/engine_test.go` for `GetLeaderboard(ctx, days)` verifying rank ordering, spend share %, token share %, active days count, cost per 1k tokens, top model, model mix breakdown, and currency propagation
    - [ ] Add `LeaderboardEntry` struct in `internal/domain/models.go` and implement `GetLeaderboard` in `internal/attribution/engine.go`
- [ ] Task: Expose `GET /api/v1/leaderboard` REST endpoint in `internal/server`
    - [ ] Write HTTP handler unit tests in `internal/server/server_test.go` for `GET /api/v1/leaderboard?days=30`
    - [ ] Register `/api/v1/leaderboard` in `internal/server/server.go` and implement handler
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Domain Model, Attribution Engine & REST API' (Protocol in workflow.md)

## Phase 2: CLI Subcommand (`agy-cost-board leaderboard`) & Interactive TUI Tab (`internal/cli`, `internal/tui`)
- [ ] Task: Implement `agy-cost-board leaderboard` CLI subcommand
    - [ ] Write unit tests in `internal/cli/root_test.go` (and `tests/e2e_test.go`) for `leaderboard` subcommand in `table`, `json`, and `csv` formats
    - [ ] Create `internal/cli/leaderboard.go` and register it on `rootCmd` with `--days` and `--format` support
- [ ] Task: Add interactive Leaderboard view to Bubbletea TUI (`internal/tui`)
    - [ ] Write unit tests in `internal/tui/tui_test.go` verifying tab navigation to Leaderboard, cursor movement, column sorting, and `Enter` drill-down
    - [ ] Update `internal/tui/model.go` to include the `Leaderboard` tab, ranked table rendering, and developer drill-down
- [ ] Task: Conductor - User Manual Verification 'Phase 2: CLI Subcommand & Interactive TUI Tab' (Protocol in workflow.md)

## Phase 3: Two-Line Web Navigation & React Leaderboard View (`web/src/`)
- [ ] Task: Refine `Navbar.tsx` into a two-line header and add `leaderboard` tab routing in `App.tsx`
    - [ ] Write unit tests in `web/src/components/Navbar.test.tsx` verifying the two-line layout, tab order (`Overview`, `Leaderboard`, `My Consumption`, `Attributed Costs`, `Seat Governance`, `Setup & Health`), and tab selection callbacks
    - [ ] Update `web/src/components/Navbar.tsx` to render Brand + Global Controls on Line 1 and the 6 navigation tabs on Line 2
- [ ] Task: Implement `LeaderboardView.tsx` with Top 3 Podium Cards and Ranked Table
    - [ ] Write unit tests in `web/src/components/LeaderboardView.test.tsx` and `web/src/App.test.tsx` covering podium cards, search filter, column sorting (Spend, Tokens, Active Days, Efficiency), model mix bar, and user drill-down modal
    - [ ] Add `LeaderboardEntry` and `fetchLeaderboard` in `web/src/api.ts` (with fallback/aggregation support), implement `web/src/components/LeaderboardView.tsx`, and wire it into `web/src/App.tsx` (including `?tab=leaderboard` URL sync)
    - [ ] Rebuild embedded frontend assets (`npm --prefix web run build`) and update `README.md` and `conductor/product.md`
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Two-Line Web Navigation & React Leaderboard View' (Protocol in workflow.md)
