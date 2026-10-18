# Implementation Plan: Clean README and Create Structured Documentation in `docs/`

## Phase 1: Core Documentation Hub, Getting Started & FinOps Architecture (`docs/`)
- [x] Task: Create `docs/README.md` (Documentation Hub & Index)
    - [x] Draft overview, persona-based navigation paths (Platform/FinOps, Engineering Managers, Developers), Mermaid system architecture diagram, and complete index table linking to all guides in `docs/`.
- [x] Task: Create `docs/getting-started.md` (Installation, Demo Lab & GCP Setup)
    - [x] Document prerequisites, building from source (`make build`), exploring the 100% offline Synthetic Demo Lab (`--demo`), and connecting to Google Cloud & BigQuery with the 5-point automated verification and provisioning engine (`setup`, `setup --create`, `setup --save`, `doctor`).
- [x] Task: Create `docs/architecture-and-finops.md` (Proportional Attribution, WTU & FinOps Model)
    - [x] Document the Mathematical Proportional Attribution formula (`Allocated Cost(u, m, d)`), Dual Cost Model coexistence (`v0.10.0`: Reconciled Cloud Billing vs. Metered WTU / Included Seat Allowance vs. PAYG Overage), Token Cache Hit Rate & Thinking Effort levels (`minimal`..`max`), Client Telemetry (`vscode`, `jetbrains`, `agy-cli`), Gemini Enterprise Shelfware Reclamation, and Native Multi-Currency (ISO-4217 zero-FX ingestion).
- [x] Task: Conductor - User Manual Verification 'Phase 1: Core Documentation Hub, Getting Started & FinOps Architecture (`docs/`)' (Protocol in workflow.md)

## Phase 2: Interfaces, Integrations & Deployment Guides (`docs/`)
- [~] Task: Create `docs/interfaces-and-reference.md` (Web UI, TUI, CLI & REST API Reference)
    - [ ] Document the Web Dashboard (two-line header navigation: `Overview`, `Leaderboard`, `My Consumption`, `Attributed Costs`, `Seat Governance`, `Setup & Health`, plus "What's New" and "MCP & Quotas" modals), Interactive TUI (`agy-cost-board tui`), complete CLI Command Reference table & flags, and REST API Reference table (`v0.10.0` fields included).
- [ ] Task: Create `docs/integrations.md` (Antigravity MCP Server, Plugin & CLI Statusline)
    - [ ] Document the Model Context Protocol (`2025-03-26`) Server (exposed tools, Streamable HTTP `POST /mcp`, Stdio `agy-cost-board mcp`, `AGY_COST_BOARD_MCP_ENABLED` toggle, and `.agents/plugins/agy-cost-board/` plugin bundle) and the Antigravity CLI Statusline Spend Integration (`GET /statusline.sh`, query params, user resolution hierarchy, env vars, Cloud Run IAM Bearer auth, and fail-open caching).
- [ ] Task: Create `docs/deployment.md` (Cloud Run Deployment & Configuration Reference)
    - [ ] Document One-Click Cloud Run Button deployment (`app.json`), declarative Knative deployment (`deploy.yaml`), manual `gcloud builds submit` & `gcloud run deploy`, IAM / IAP access patterns, and complete `AGY_COST_BOARD_*` environment variable & `.agy-cost-board.yaml` configuration reference.
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Interfaces, Integrations & Deployment Guides (`docs/`)' (Protocol in workflow.md)

## Phase 3: Streamline `README.md` & Verify Documentation Links
- [ ] Task: Refactor `README.md` into a concise, high-impact landing page (~120–160 lines)
    - [ ] Keep Hero & badges, screenshots (`docs/assets/overview.png`, `docs/assets/consumption.png`), concise "Why agy-cost-board?" & "The Difference at a Glance" table, brief `v0.10.0`-updated Superpowers summary, 60-Second Quick Start, Documentation Map linking to `docs/*.md`, Development & Testing commands, and Community & License links.
- [ ] Task: Verify relative links, image references, and formatting across `README.md` and `docs/*.md`
    - [ ] Validate that all relative links between `README.md`, `docs/*.md`, `docs/assets/*.png`, and community files (`CONTRIBUTING.md`, `SECURITY.md`, etc.) resolve cleanly and `make test` passes.
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Streamline `README.md` & Verify Documentation Links' (Protocol in workflow.md)
