# Specification: MVP Track - Dual-Mode CLI & Cloud Run Web Dashboard for Gemini Enterprise & Antigravity

## 1. Overview & Objectives
Build the core MVP of `agy-ge-board`, a single-binary application deployable to Cloud Run and operable via CLI/TUI. It extracts Antigravity inference telemetry from Cloud Logging in BigQuery, computes exact per-user proportional cost allocations against GCP Cloud Billing export datasets, monitors Gemini Enterprise seat quotas, and flags dormant or inactive license holders.

## 2. Functional Requirements

### 2.1 BigQuery Telemetry & Cost Engine
- Ingest `InferenceResponseLog` records from BigQuery log sinks.
- Parse `labels.user_id`, `labels.model`, and `metadata.totalTokenCount`.
- Query Cloud Billing export tables for Gemini and Antigravity SKUs.
- Calculate proportional user cost:
  $$\text{User Cost} = \frac{\text{User Tokens}}{\text{Project Model Tokens}} \times \text{Net Invoiced SKU Cost}$$
- Provide an in-memory TTL cache to optimize query performance and reduce BigQuery analysis costs.

### 2.2 License Management & Dormant Seat Governance
- Calculate active user counts within configurable time windows (e.g., 7, 14, 30 days).
- Compare active users against a configurable project seat quota (`--seat-quota`).
- Classify seats as Active, At-Risk (low token consumption), or Dormant (zero activity in window).
- Compute potential monthly savings from reclaiming dormant licenses.

### 2.3 CLI & TUI Interface
- Command structure built with Cobra:
  - `serve`: Starts the HTTP API and embedded Web UI server, listening on `$PORT` (default: 8080).
  - `cost`: Outputs per-user cost attribution table (supports `--days`, `--model`, `--format=table|json|csv`).
  - `license`: Displays license quota status, assigned vs active seats, and dormant users.
  - `user <email>`: Detailed breakdown of an individual developer's token history, model breakdown, and cost.
  - `doctor`: Validates ADC credentials, BigQuery connectivity, and configured dataset permissions.
- Interactive TUI dashboard (via Bubbletea/Lipgloss) launched when running `cost` or `license` with `--tui`.

### 2.4 Web Dashboard & REST API
- Embedded React/Vite/Tailwind SPA served directly from Go binary via `embed.FS`.
- 3 Core Views:
  1. **Overview Dashboard:** Top KPI cards (Total Billed AI Spend, Active Developers, License Utilization %, Potential Dormant Savings) and 30-day daily cost/token trend line charts.
  2. **Cost Attribution Table:** Searchable, sortable table by user email, total tokens, proportional share, and allocated cost ($) with model filter dropdown.
  3. **License Governance:** Donut visualizer for seat quotas, list of dormant users with last-active timestamps, and seat reclamation recommendations.
- REST API Endpoints:
  - `GET /api/v1/health`
  - `GET /api/v1/metrics/overview`
  - `GET /api/v1/costs/users`
  - `GET /api/v1/licenses/status`

### 2.5 Demonstration & Offline Mode
- Provide a `--demo` flag that initializes rich synthetic 30-day telemetry and billing datasets.
- Automatically activated if no Google Cloud credentials are found and running in local development mode.

## 3. Non-Functional & Architecture Requirements
- **Single Binary:** Fully self-contained Go executable embedding the compiled React SPA.
- **Containerization:** Multi-stage Dockerfile producing a minimal, secure distroless/scratch container for Google Cloud Run.
- **Portability:** Reads `$PORT` for Cloud Run compliance, adheres to graceful SIGTERM shutdown.
- **Quality & Testing:** Adherence to strict TDD (Red/Green/Refactor) with >80% test coverage on Go core logic and React UI components.

## 4. Acceptance Criteria
- [ ] Running `./agy-ge-board cost --demo` outputs formatted tabular cost attribution in the terminal.
- [ ] Running `./agy-ge-board cost --demo --format=json` outputs valid JSON parseable by `jq`.
- [ ] Running `./agy-ge-board license --demo` shows seat utilization and identifies dormant users.
- [ ] Running `./agy-ge-board serve --demo` launches HTTP server on `$PORT` and serves the React dashboard at `http://localhost:$PORT`.
- [ ] Navigating between Overview, Cost Attribution, and License views in the web UI renders data and charts seamlessly.
- [ ] Docker container builds cleanly and runs locally with `docker run -e PORT=8080 -p 8080:8080 <image> serve --demo`.
- [ ] Go test suite passes with >80% coverage.

## 5. Out of Scope for MVP
- Direct write-back/mutation to Google Workspace Admin SDK (automated revoking of seats).
- Real-time WebSockets streaming (batch BigQuery queries with caching are sufficient).
- Multi-cloud billing providers (AWS/Azure).
