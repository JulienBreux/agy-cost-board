# Interfaces, CLI & API Reference

`agy-cost-board` ships three full-featured interfaces in a single static Go binary: an embedded **React 19 Web SPA**, an interactive **Bubble Tea Terminal UI (TUI)**, and a scriptable **CLI & REST API**.

---

## 1. Web Dashboard (`agy-cost-board serve`)

Start the HTTP server (default port `8080`):
```bash
./bin/agy-cost-board serve --demo --port 8080
```

### Two-Line Header Navigation
1. **Top Bar (Platform Governance & Global Controls)**:
   - **Seat Governance**: Inspect Gemini Enterprise license tiers (`Standard` / `Plus`), regional stores (`global`, `eu`, `us`), trial expiration badges, and shelfware savings.
   - **Setup & Health**: Real-time 5-point GCP telemetry & BigQuery diagnostic checklist.
   - **GCP Project Selector** *(when multiple projects are monitored)*: Filter the entire dashboard to `All Projects (N)` or an individual GCP project (`?project=gcp-ai-platform-prod`).
   - **Lookback Window**: Switch between `Last 7 Days`, `Last 14 Days`, `Last 30 Days`, and `Last 90 Days`.
   - **MCP & Quotas**: Inspect live seat quota bars, verify `AGY_COST_BOARD_MCP_ENABLED`, and copy `mcp_config.json` snippets.
   - **What's New (✨)**: Open the release changelog modal (`agy_last_seen_version` cookie + `localStorage`).
   - **Theme Toggle**: Switch between Dark and Light themes.
2. **Second Bar (FinOps & Developer Analytics)**:
   - **Overview**:
     - **Executive KPIs**: Total Billed Cost, Included Pool vs. PAYG Overage, Active Developers, Prompt Cache Hit Rate %, and Reclaimable Seat Savings.
     - **Weekly Allowance Pool & 7-Day Rolling Windows**: Interactive What-If toggles (`Pooled` vs. `Per-User`, `Assigned` vs. `Purchased`), Cloud Monitoring Quota Meter reconciliation (`metered_wtu` vs. `log_estimated_wtu`), and 7-day window history.
     - **30-Day Window-Aware Usage & Spend Forecast**: Projects daily WTU against daily pool capacity (`weekly_pool_limit / 7`) to forecast 30-day Included vs. Spillover PAYG WTU and cost.
     - **Daily Reconciled Spend & Model Breakdown**: Stacked daily spend chart (`Included` vs. `PAYG` or by model).
     - **Where Tokens & Money Go / Models & Thinking Effort / Antigravity Clients**: Compares Token Share % vs. Cost Share % across Uncached Input, Cached Input, Output, and Thinking tokens; breaks down model effort (`minimal`..`max`) and client surfaces (`vscode`, `jetbrains`, `agy-cli`).
   - **Leaderboard**:
     - Top-3 developer podium and ranked table with **Total Cost**, **Included / PAYG Split**, **WTU**, **Total Tokens**, **Cache Hit Rate %**, and **Top Model**.
     - Interactive **Fair Share / When Used** PAYG attribution toggle and **Export CSV** button.
   - **My Consumption**:
     - Developer self-service cockpit (`?tab=my-consumption&user=...`) with personal month-to-date spend, 7-day burn rate, customizable monthly budget progress bar, model/effort optimization recommendations, and recent activity log.
   - **Attributed Cost**:
     - Granular per-user, per-model, per-day reconciled cost table with **Fair Share / When Used** toggle, model filter, and **Export CSV** for FinOps chargeback.

---

## 2. Interactive Terminal UI (`agy-cost-board tui`)

Launch the full-screen keyboard-driven terminal dashboard:
```bash
./bin/agy-cost-board tui --demo
```

### Keybindings
| Key | Action |
| :--- | :--- |
| `Tab` / `1`–`4` | Switch between **Overview** (`1`), **Attributed Costs** (`2`), **License Governance** (`3`), and **Setup & Health** (`4`) |
| `r` | Refresh telemetry and billing data from BigQuery |
| `q` / `Ctrl+C` | Exit the TUI |

---

## 3. CLI Command Reference

| Command | Description | Example |
| :--- | :--- | :--- |
| `setup` | Run 5-point GCP diagnostic and optionally create dataset/sink (`--create`) or save `.agy-cost-board.yaml` (`--save`) | `agy-cost-board setup --project=my-prj --save` |
| `cost` | Print per-developer, per-model, per-day attributed costs (`table`, `json`, or `csv`) | `agy-cost-board cost --days=30 --payg-mode=fair_share` |
| `leaderboard` | Print ranked developer leaderboard with WTU, Included/PAYG split, and Cache Hit % | `agy-cost-board leaderboard --days=30` |
| `license` | Analyze Gemini Enterprise seat utilization (`Standard`/`Plus`), weekly pool status, and dormant seats | `agy-cost-board license --days=30 --pool-mode=pooled` |
| `user <email>` | Inspect an individual developer's spend, WTU, cache efficiency, and model breakdown | `agy-cost-board user alex.rivera@acme.corp` |
| `doctor` | Quick 5-point connectivity and configuration health check | `agy-cost-board doctor` |
| `tui` | Launch interactive Bubble Tea terminal dashboard | `agy-cost-board tui --demo` |
| `serve` | Start HTTP server with embedded React 19 SPA, REST API, Statusline, and `/mcp` | `agy-cost-board serve --port=8080` |
| `mcp` | Start Model Context Protocol (MCP `2025-03-26`) JSON-RPC server over `stdio` | `agy-cost-board mcp --demo` |

### Global & Command Flags

| Flag | Environment Variable | Default | Description |
| :--- | :--- | :--- | :--- |
| `--project` | `AGY_COST_BOARD_PROJECT_ID` | `""` | Active GCP Project ID (or filter to a specific monitored project) |
| `--monitored-projects` | `AGY_COST_BOARD_MONITORED_PROJECTS` | `""` | Comma-separated list of monitored GCP Project IDs |
| `--telemetry-table` | `AGY_COST_BOARD_TELEMETRY_TABLE` | `""` | Full BigQuery table path for `InferenceResponseLog` records |
| `--billing-table` | `AGY_COST_BOARD_BILLING_TABLE` | `""` | Full BigQuery table path for Cloud Billing export |
| `--seat-quota` | `AGY_COST_BOARD_SEAT_QUOTA` | `50` | Total purchased Gemini Enterprise seat quota |
| `--seat-price` | `AGY_COST_BOARD_SEAT_PRICE` | `45.0` | Reference monthly price per seat for shelfware ROI calculation |
| `--license-locations` | `AGY_COST_BOARD_LICENSE_LOCATIONS` | `global,eu,us` | Comma-separated Discovery Engine locations to scan for licenses |
| `--pool-mode` | `AGY_COST_BOARD_POOL_MODE` | `pooled` | Weekly allowance pool mode (`pooled` or `per_user`) |
| `--pool-basis` | `AGY_COST_BOARD_POOL_COUNT_BASIS` | `assigned` | Seat count basis for weekly pool (`assigned` or `purchased`) |
| `--payg-mode` | `AGY_COST_BOARD_PAYG_ATTRIBUTION_MODE` | `fair_share` | PAYG overage attribution mode (`fair_share` or `when_used`) |
| `--days` | — | `30` | Lookback window in days (`7`, `14`, `30`, `90`) |
| `--format` | — | `table` | CLI output format (`table`, `json`, `csv`) |
| `--demo` | `AGY_COST_BOARD_DEMO` | `false` | Use offline synthetic multi-project demonstration data |
| `--mcp-enabled` | `AGY_COST_BOARD_MCP_ENABLED` | `true` | Enable or disable the built-in MCP server (`/mcp` and `mcp` subcommand) |

---

## 4. REST API Reference

All REST endpoints return `application/json` (except `/statusline.sh` which returns `text/x-shellscript`).

| Method | Endpoint | Query Parameters | Description |
| :--- | :--- | :--- | :--- |
| `GET` | `/healthz` | — | Liveness and readiness health check (`{"status":"ok"}`) |
| `GET` | `/api/v1/version` | — | Application semantic version, git commit, and build date |
| `GET` | `/api/v1/metrics/overview` | `days`, `pool_mode`, `pool_basis`, `project` | Organization KPIs, `weekly_pool` (with `monitoring_meter`), `forecast` (30-day window-aware projection), `token_cost_shares`, `model_effort_stats`, `client_breakdown`, and `monitored_projects` |
| `GET` | `/api/v1/costs/users` | `days`, `model`, `payg_mode`, `project` | Per-developer, per-model, per-day attributed costs (`fair_share` or `when_used`) |
| `GET` | `/api/v1/leaderboard` | `days`, `payg_mode`, `project` | Ranked developer leaderboard with WTU, Included/PAYG split, and Cache Hit % |
| `GET` | `/api/v1/licenses/status` | `days`, `pool_mode`, `pool_basis`, `project` | Seat governance, `tier_breakdown`, `location_breakdown`, `weekly_pool`, and dormant seats |
| `GET` | `/api/v1/users/{id}` | `days`, `budget` | Individual developer summary, burn rate, budget status, and model breakdown |
| `GET` | `/api/v1/setup/status` | — | Real-time 5-point GCP telemetry & BigQuery diagnostic report |
| `GET` | `/api/v1/mcp/status` | — | MCP server status (`mcp_enabled`), protocol version (`2025-03-26`), and exposed tool list |
| `POST` | `/mcp` | — | Model Context Protocol (`2025-03-26`) Streamable HTTP JSON-RPC endpoint |
| `GET` | `/statusline.sh` | `user`, `days`, `ttl` | Dynamic Antigravity CLI statusline script with fail-open spend display |
