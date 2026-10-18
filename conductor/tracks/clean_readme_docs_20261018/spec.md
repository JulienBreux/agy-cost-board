# Specification: Clean README and Create Structured Documentation in `docs/`

## 1. Overview
Currently, `README.md` is a 452-line monolithic document containing product marketing, mathematical formulas, CLI/API reference tables, MCP server configuration, Statusline script setup, and Cloud Run deployment instructions, while `docs/` only houses image assets (`docs/assets/`).

This track refactors `README.md` into a crisp, high-impact landing page (~120–160 lines) and extracts comprehensive, well-structured documentation into a dedicated `docs/` knowledge base, fully updated with the latest `v0.10.0` capabilities (Weighted Token Units / WTU, Token Cache Hits, Thinking Effort levels, Client Telemetry, Included vs. PAYG coexistence, and the Developer Leaderboard).

## 2. Functional Requirements

### 2.1 Streamlined `README.md` (~120–160 lines)
1. **Hero & Badges**: Project title, badges (Go, License, CI, Release, Cloud Run Ready, Run on Google Cloud button), 2-sentence elevator pitch, and link to the foundational Medium article.
2. **Dashboard Visuals**: Embedded screenshots (`docs/assets/overview.png`, `docs/assets/consumption.png`) with concise captions.
3. **Why `agy-cost-board` & Core Superpowers**:
   - Concise comparison table ("The Difference at a Glance") and brief bulleted highlights covering:
     - Mathematical Proportional Attribution & Reconciled Billing
     - WTU, Included Seat Pool vs. PAYG Overage, Cache Hit Rate & Thinking Effort (`v0.10.0`)
     - Gemini Enterprise Shelfware Reclamation & Developer Leaderboard
     - Triple-Mode Single Binary (CLI, TUI, React 19 Web SPA) & Native Multi-Currency
     - Antigravity MCP Server, Plugin & CLI Statusline Integration
4. **60-Second Quick Start**: Build/install and immediate `--demo` commands (`cost`, `leaderboard`, `license`, `tui`, `serve`), plus 1-click Cloud Run deploy button.
5. **Documentation Hub Map**: Clear table/links directing readers to each guide in `docs/`.
6. **Community & License**: Links to `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `SECURITY.md`, `MAINTAINERS.md`, and `LICENSE`.

### 2.2 Comprehensive Documentation Suite in `docs/`
1. **`docs/README.md` (Documentation Hub & Index)**:
   - Central table of contents, persona-based reading paths (Platform/FinOps Engineers, Engineering Managers, Developers), and architecture overview diagram.
2. **`docs/getting-started.md` (Installation, Demo Lab & GCP Setup)**:
   - Building from source (`make build`), running the 100% offline Synthetic Demo Lab (`--demo`), and connecting to Google Cloud & BigQuery using the 5-point diagnostic & provisioning engine (`setup`, `setup --create`, `setup --save`, `doctor`).
3. **`docs/architecture-and-finops.md` (Proportional Attribution, WTU & FinOps Model)**:
   - Proportional billing reconciliation formula (`Allocated Cost(u, m, d)`), Dual Cost Model coexistence (Reconciled Cloud Billing vs. Metered WTU / Included Seat Pool vs. PAYG Overage), Token Cache Hit Rate & Thinking Effort levels (`minimal`..`max`), Client Telemetry (`vscode`, `jetbrains`, `agy-cli`), Gemini Enterprise Shelfware Reclamation (`Active`, `At-Risk`, `Dormant`), and Native Multi-Currency (ISO-4217 zero-FX ingestion).
4. **`docs/interfaces-and-reference.md` (Web UI, TUI, CLI & REST API Reference)**:
   - Web Dashboard guide (Two-line navigation: `Overview`, `Leaderboard`, `My Consumption`, `Attributed Costs`, `Seat Governance`, `Setup & Health`, plus "What's New" release modal and "MCP & Quotas" modal).
   - Interactive TUI guide (`agy-cost-board tui` tabs and keybindings).
   - Complete CLI Command Reference (`setup`, `cost`, `leaderboard`, `license`, `user`, `doctor`, `tui`, `serve`, `mcp`) and global flags.
   - Complete REST API Reference (`/healthz`, `/api/v1/metrics/overview`, `/api/v1/costs/users`, `/api/v1/leaderboard`, `/api/v1/licenses/status`, `/api/v1/users/{id}`, `/api/v1/setup/status`, `/api/v1/mcp/status`, `/mcp`, `/statusline.sh`).
5. **`docs/integrations.md` (Antigravity MCP Server, Plugin & CLI Statusline)**:
   - Model Context Protocol (MCP `2025-03-26`) Server: exposed tools (`get_overview_metrics`, `get_license_status`, `get_users_costs`, `get_user_dashboard`), Streamable HTTP (`POST /mcp`) vs. Stdio (`agy-cost-board mcp`), `AGY_COST_BOARD_MCP_ENABLED` toggle, and `.agents/plugins/agy-cost-board/` plugin bundle.
   - Antigravity CLI Statusline Spend Integration: 1-line install, `~/.gemini/antigravity-cli/settings.json` configuration, query parameters (`user`, `days`, `ttl`), identity resolution hierarchy, env var overrides, Cloud Run IAM Bearer auth, and fail-open caching architecture.
6. **`docs/deployment.md` (Cloud Run Deployment & Configuration Reference)**:
   - One-click Cloud Run Button (`app.json`), declarative Knative deployment (`deploy.yaml`), manual `gcloud builds submit` & `gcloud run deploy`, IAM / IAP access patterns, and complete `AGY_COST_BOARD_*` environment variable & `.agy-cost-board.yaml` configuration reference.

## 3. Acceptance Criteria
- `README.md` is concise (~120–160 lines), visually scannable, preserves key badges and screenshots, and links cleanly to all `docs/*.md` files.
- All 6 markdown files in `docs/` are created, accurate, cross-linked with working relative links, and reflect `v0.10.0` features.
- No broken links or missing documentation topics from the original `README.md`.

## 4. Out of Scope
- Modifying application runtime logic (Go backend or React frontend code), except if any doc tests or linters require updates.
