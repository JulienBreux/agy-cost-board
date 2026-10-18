# Getting Started & GCP Setup

This guide walks you through building `agy-cost-board` from source, exploring the 100% offline Synthetic Demo Lab (`--demo`), and connecting to your Google Cloud & BigQuery environment using the built-in 5-point diagnostic and provisioning engine.

---

## 1. Prerequisites

To build `agy-cost-board` from source, you need:
- **Go 1.27+** (compiled with `CGO_ENABLED=0` for a standalone static binary)
- **Node.js 22+ & npm** (to compile the embedded React 19 + Tailwind CSS web frontend)
- **Google Cloud CLI (`gcloud`)** *(optional—only required when connecting to live GCP projects)*

---

## 2. Build from Source

Clone the repository and run `make build` to compile both the React web assets (`web/dist`) and the standalone Go executable (`./bin/agy-cost-board`):

```bash
git clone https://github.com/JulienBreux/agy-cost-board.git
cd agy-cost-board
make build
```

You can verify the compiled binary version:
```bash
./bin/agy-cost-board --version
```

---

## 3. Run the 100% Offline Synthetic Demo Lab (`--demo`)

You can evaluate every feature of `agy-cost-board`—including proportional billing attribution, Weekly Allowance Pools (`Standard` & `Plus`), Fair-Share vs. When-Used PAYG chargeback, 30-day window-aware forecasting, Cloud Monitoring quota meter comparison, multi-project filtering (`gcp-ai-platform-prod`, `gcp-finops-analytics`, `gcp-agent-sandbox`), and shelfware reclamation—without any Google Cloud credentials or network access.

> [!NOTE]
> Model names featured in the demo data (e.g., `gemini-4.0-pro`, `gemini-3.8-flash`, `gemini-4-argon`, `claude-5.5-sonnet`) are purely synthetic and used for simulation and demonstration purposes.

```bash
# 1. Print proportional per-developer cost attribution table
./bin/agy-cost-board cost --demo

# 2. Switch PAYG chargeback mode between Fair Share (default) and When Used
./bin/agy-cost-board cost --demo --payg-mode=when_used

# 3. Filter metrics to a specific monitored GCP project
./bin/agy-cost-board cost --demo --project=gcp-ai-platform-prod

# 4. View the ranked developer leaderboard (spend, WTU, cache hit rate, top model)
./bin/agy-cost-board leaderboard --demo

# 5. Inspect Gemini Enterprise license utilization, tiers, and shelfware savings
./bin/agy-cost-board license --demo

# 6. Drill down into an individual developer's consumption & budget pacing
./bin/agy-cost-board user alex.rivera@acme.corp --demo

# 7. Launch the interactive full-screen Terminal UI (TUI)
./bin/agy-cost-board tui --demo

# 8. Start the embedded React 19 Web Dashboard & MCP Server on http://localhost:8080
./bin/agy-cost-board serve --demo --port 8080
```

---

## 4. Connecting to Google Cloud & BigQuery

In production mode, `agy-cost-board` correlates two BigQuery tables:
1. **Cloud Logging Telemetry Table**: Stores `InferenceResponseLog` records exported from Cloud Logging via a Log Sink.
2. **Cloud Billing Export Table**: Standard or Detailed BigQuery Cloud Billing export (`gcp_billing_export_v1_*`).

### Step 1: Authenticate with Google Cloud ADC
Ensure your local environment (or Cloud Run service account) has valid Application Default Credentials:
```bash
gcloud auth application-default login
```

### Step 2: Run the 5-Point Automated Diagnostic Engine (`setup` / `doctor`)
`agy-cost-board` includes an automated verification engine that checks 5 critical infrastructure prerequisites:

1. **Google Cloud ADC & Project**: Verifies active credentials and target GCP project binding.
2. **GCP IAM Permissions**: Verifies `roles/bigquery.jobUser` and `roles/bigquery.dataViewer` permissions.
3. **Cloud Logging Sink**: Inspects Cloud Logging sinks for the `InferenceResponseLog` ingestion filter.
4. **Cloud Billing Export**: Validates the BigQuery Cloud Billing export dataset and table schema.
5. **Telemetry Pipeline Probe**: Executes a live dry-run/probe query against the telemetry dataset.

Run the diagnostic check against your GCP project:
```bash
./bin/agy-cost-board setup --project=YOUR_GCP_PROJECT_ID
```

Or run a quick health check at any time:
```bash
./bin/agy-cost-board doctor --project=YOUR_GCP_PROJECT_ID
```

### Step 3: Auto-Provision Missing Infrastructure (`setup --create`)
If your project does not yet have a BigQuery dataset or Cloud Logging sink for Antigravity telemetry, `agy-cost-board` can provision them automatically (or preview the exact changes with `--dry-run`):

```bash
# Preview infrastructure provisioning without making changes
./bin/agy-cost-board setup --project=YOUR_GCP_PROJECT_ID --create --dry-run

# Provision the BigQuery telemetry dataset and Cloud Logging sink
./bin/agy-cost-board setup --project=YOUR_GCP_PROJECT_ID --create
```

### Step 4: Save Configuration (`.agy-cost-board.yaml`)
Persist your verified configuration to `.agy-cost-board.yaml` so you don't need to pass flags on every command:

```bash
./bin/agy-cost-board setup \
  --project=YOUR_GCP_PROJECT_ID \
  --telemetry-table=YOUR_GCP_PROJECT_ID.antigravity_telemetry.logs \
  --billing-table=YOUR_GCP_PROJECT_ID.billing_export.gcp_billing_export_v1_XXXXXX \
  --seat-quota=50 \
  --save
```

Example `.agy-cost-board.yaml`:
```yaml
project_id: your-gcp-project-id
monitored_projects:
  - your-gcp-project-id
  - your-secondary-ai-project
telemetry_table: your-gcp-project-id.antigravity_telemetry.logs
billing_table: your-gcp-project-id.billing_export.gcp_billing_export_v1_XXXXXX
seat_quota: 50
seat_price: 45.0
license_locations:
  - global
  - eu
  - us
pool_mode: pooled
pool_count_basis: assigned
payg_attribution_mode: fair_share
port: 8080
mcp_enabled: true
```

### Step 5: Launch against Production BigQuery
Once `.agy-cost-board.yaml` is saved (or `AGY_COST_BOARD_*` environment variables are set):

```bash
./bin/agy-cost-board serve --port 8080
```

---

## Next Steps
- Learn how proportional cost attribution, WTU weekly pools, and 30-day forecasting work in [**Architecture & FinOps Model**](./architecture-and-finops.md).
- Explore all Web UI tabs, TUI shortcuts, CLI flags, and REST API endpoints in [**Interfaces, CLI & API Reference**](./interfaces-and-reference.md).
