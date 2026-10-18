# agy-cost-board

[![Go Version](https://img.shields.io/badge/Go-1.27+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)
[![CI](https://github.com/JulienBreux/agy-cost-board/actions/workflows/ci.yml/badge.svg)](https://github.com/JulienBreux/agy-cost-board/actions/workflows/ci.yml)
[![Release](https://github.com/JulienBreux/agy-cost-board/actions/workflows/release.yml/badge.svg)](https://github.com/JulienBreux/agy-cost-board/actions/workflows/release.yml)
[![Cloud Run](https://img.shields.io/badge/Cloud_Run-Ready-4285F4?style=flat&logo=googlecloud)](https://cloud.google.com/run)
[![Run on Google Cloud](https://deploy.cloud.run/button.svg)](https://deploy.cloud.run)

**`agy-cost-board`** is an enterprise-grade AI FinOps engine, Weekly Allowance Pool simulator, and developer telemetry board built for **Google Antigravity** and **Gemini Enterprise**.

It reconciles Google Cloud Billing exports with BigQuery inference logs (`InferenceResponseLog`) and Cloud Monitoring quota meters (`agentic_wtu_tier_*_pp_weekly_regional`) to provide exact per-developer cost attribution, Weighted Token Unit (WTU) weekly pool governance, 30-day spillover forecasting, and shelfware license reclamation in a single static Go binary (CLI, Bubble Tea TUI, embedded React 19 SPA, and MCP Server).

Based on the Google Cloud architecture article [**"Per-user cost attribution for Antigravity with BigQuery"**](https://medium.com/google-cloud/per-user-cost-attribution-for-antigravity-with-bigquery-3e98fd997c58).

<p align="center">
  <img src="docs/assets/overview.png" alt="agy-cost-board FinOps Overview Dashboard" width="100%" />
  <br />
  <em><b>Executive FinOps Overview:</b> Reconciled Cloud Billing spend, 7-day Weekly Allowance Pool drawdown, 30-day spillover forecast, and model effort repartition.</em>
</p>

<p align="center">
  <img src="docs/assets/consumption.png" alt="agy-cost-board Developer Self-Service Dashboard" width="100%" />
  <br />
  <em><b>Developer Self-Service ("My Consumption"):</b> Personal burn rate, monthly budget tracking, prompt cache efficiency, and model optimization insights.</em>
</p>

---

## Why `agy-cost-board`?

| Capability | Traditional Cloud Billing | `agy-cost-board` |
| :--- | :--- | :--- |
| **Per-Developer Attribution** | Aggregate pooled SKU lines; zero engineer visibility | **Reconciled Cloud Billing attribution + Fair-Share vs. When-Used PAYG chargeback** |
| **Weekly Allowance Pools** | No visibility into 7-day rolling WTU seat pools | **7-day WTU Pool simulator (`Standard` & `Plus` tiers) + Cloud Monitoring meter sync** |
| **30-Day Spend Forecasting** | Naïve linear projections that ignore weekly pool resets | **Window-aware 30-day forecast predicting Included vs. Spillover PAYG WTU** |
| **Token & Effort Analytics** | Raw token counts without cache or reasoning context | **Token Share vs. Cost Share (Cached Input, Thinking), Effort (`minimal`..`max`) & Clients** |
| **License Governance** | Blind seat renewals ($45/seat/mo shelfware) | **Multi-location (`global`,`eu`,`us`) & multi-tier dormancy audit with ROI savings** |
| **Interfaces & Agent Integration** | Static spreadsheets or heavy multi-service stacks | **Single static binary (`CGO_ENABLED=0`): CLI, TUI, React 19 Web SPA, MCP & Statusline** |
| **Multi-Project & Currency** | Single-project USD-only assumptions | **Multi-project switcher (`--monitored-projects`) & native ISO-4217 currency (zero FX)** |

### Core Superpowers
- **📐 Reconciled Billing + WTU Coexistence**: Allocates exact Cloud Billing invoice costs (`Allocated Cost(u, m, d)`) alongside metered Weighted Token Units (**WTU**, `$2 / 1M WTU`, `50 WTU` minimum call floor) split into **Included Seat Quota** vs. **PAYG Overage**.
- **🌊 7-Day Weekly Allowance Pool & Fair-Share Chargeback**: Models floating 7-day `Standard` (`$2.50/wk`) and `Plus` (`$3.75/wk`) pools with interactive **Pooled vs. Per-User** and **Assigned vs. Purchased** What-If controls, **Cloud Monitoring Quota Meter** drift reconciliation, and **Fair Share vs. When Used** PAYG chargeback with 1-click CSV export.
- **📈 30-Day Window-Aware Forecast & Multi-Project Support**: Projects 7-day trailing WTU velocity against daily pool capacity across 30 days and filters unified or per-project boards across multiple GCP projects.
- **🧠 Cache Hit Rate, Thinking Effort & Leaderboard**: Tracks `cachedContentTokenCount`, reasoning effort levels (`minimal` through `max`), client surfaces (`vscode`, `jetbrains`, `agy-cli`), and ranks developers on an interactive **Leaderboard** and **My Consumption** cockpit.
- **🤖 Built-in MCP Server (`2025-03-26`), Plugin & CLI Statusline**: Exposes real-time FinOps tools over Streamable HTTP (`POST /mcp`) and `stdio` (`agy-cost-board mcp`), plus a fail-open CLI statusline (`GET /statusline.sh`).

---

## Quick Start (60 Seconds)

### 1. Build from Source
Requires **Go 1.27+** and **Node.js 22+**:
```bash
git clone https://github.com/JulienBreux/agy-cost-board.git
cd agy-cost-board
make build
```

### 2. Explore Immediately in Offline Demo Mode (`--demo`)
No Google Cloud credentials or network access required *(synthetic model names like `gemini-4.0-pro` are for demonstration only)*:
```bash
# Print proportional per-developer cost attribution (Fair Share or When Used)
./bin/agy-cost-board cost --demo
./bin/agy-cost-board cost --demo --payg-mode=when_used

# View ranked developer leaderboard & Gemini Enterprise seat governance
./bin/agy-cost-board leaderboard --demo
./bin/agy-cost-board license --demo

# Launch the interactive Terminal UI (TUI) or Web Dashboard (http://localhost:8080)
./bin/agy-cost-board tui --demo
./bin/agy-cost-board serve --demo --port 8080
```

### 3. Deploy to Cloud Run in One Click
Verify your GCP project and BigQuery tables (`./bin/agy-cost-board setup --project=YOUR_PROJECT_ID --save`), deploy declaratively via [`deploy.yaml`](deploy.yaml), or deploy in one click (and access securely via `gcloud run services proxy agy-cost-board --project=YOUR_PROJECT_ID` or Cloud IAM/IAP):

[![Run on Google Cloud](https://deploy.cloud.run/button.svg)](https://deploy.cloud.run)

---

## 📚 Documentation Hub (`docs/`)

Full documentation is organized in the [**`docs/`**](./docs/README.md) directory:

| Guide | Description |
| :--- | :--- |
| [**Documentation Index & Architecture**](./docs/README.md) | Persona-based reading paths and end-to-end system architecture diagram. |
| [**Getting Started & GCP Setup**](./docs/getting-started.md) | Building from source, `--demo` lab walkthrough, and 5-point `setup` / `doctor` provisioning. |
| [**Architecture & FinOps Model**](./docs/architecture-and-finops.md) | Proportional billing math, WTU metering, 7-day Weekly Pools, Fair-Share vs. When-Used, 30-day forecast, Cloud Monitoring meter sync, cache/effort analytics, and ISO-4217 currency. |
| [**Interfaces, CLI & API Reference**](./docs/interfaces-and-reference.md) | Web UI guide, Bubble Tea TUI shortcuts, CLI commands & flags, and REST API endpoints. |
| [**Integrations: MCP, Plugin & Statusline**](./docs/integrations.md) | Model Context Protocol (`2025-03-26`) server, Antigravity plugin, and `/statusline.sh` setup. |
| [**Cloud Run Deployment & Configuration**](./docs/deployment.md) | One-click Cloud Run deploy, `deploy.yaml`, IAM/IAP security, and `AGY_COST_BOARD_*` reference. |

---

## Development & Testing

```bash
make test       # Run Go & Frontend unit tests with race detection
make lint       # Run static analysis
make build-web  # Rebuild embedded React 19 frontend (web/dist)
make build      # Compile static binary (./bin/agy-cost-board)
```

---

## Community & License

- 🤝 [**Contributing Guide**](CONTRIBUTING.md) · 📜 [**Code of Conduct**](CODE_OF_CONDUCT.md) · 🔒 [**Security Policy**](SECURITY.md) · 👥 [**Maintainers**](MAINTAINERS.md)
- Licensed under the [**Apache 2.0 License**](LICENSE).
