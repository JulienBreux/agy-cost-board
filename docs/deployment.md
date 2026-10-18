# Cloud Run Deployment & Configuration Reference

`agy-cost-board` is compiled as a zero-CGO static binary (`CGO_ENABLED=0`) with embedded React 19 web assets, making it ideal for serverless deployment on **Google Cloud Run**.

---

## 1. Deploy to Cloud Run in One Click (`app.json`)

Deploy `agy-cost-board` directly to your Google Cloud project with one click using the Cloud Run Button:

[![Run on Google Cloud](https://deploy.cloud.run/button.svg)](https://deploy.cloud.run)

### Prerequisites
1. A Google Cloud project with billing enabled.
2. The Cloud Run API (`run.googleapis.com`) and Cloud Build API (`cloudbuild.googleapis.com`) enabled.
3. BigQuery Data Viewer (`roles/bigquery.dataViewer`) and Job User (`roles/bigquery.jobUser`) permissions on your telemetry and billing export datasets.

---

## 2. Declarative Deployment via `deploy.yaml`

`deploy.yaml` provides a Knative-compatible declarative Service manifest (`serving.knative.dev/v1`) for GitOps, Terraform, and continuous delivery:

```bash
# 1. Substitute your GCP project ID and deploy declaratively
sed 's/PROJECT_ID/'"$(gcloud config get-value project)"'/g' deploy.yaml | \
  gcloud run services replace - --region=us-central1

# 2. Or apply deploy.yaml directly once configured
gcloud run services replace deploy.yaml --region=us-central1
```

---

## 3. Manual Deployment via `gcloud`

Build and deploy manually using Google Cloud Build and Cloud Run:

```bash
export AGY_COST_BOARD_PROJECT_ID="YOUR_GCP_PROJECT_ID"
export IMAGE="gcr.io/${AGY_COST_BOARD_PROJECT_ID}/agy-cost-board:latest"

# 1. Build and push container image
gcloud builds submit --tag ${IMAGE} .

# 2. Deploy to Cloud Run
gcloud run deploy agy-cost-board \
  --image=${IMAGE} \
  --project=${AGY_COST_BOARD_PROJECT_ID} \
  --region=us-central1 \
  --platform=managed \
  --no-allow-unauthenticated \
  --set-env-vars="AGY_COST_BOARD_PROJECT_ID=${AGY_COST_BOARD_PROJECT_ID},AGY_COST_BOARD_TELEMETRY_TABLE=${AGY_COST_BOARD_PROJECT_ID}.antigravity_telemetry.logs,AGY_COST_BOARD_BILLING_TABLE=${AGY_COST_BOARD_PROJECT_ID}.billing_export.gcp_billing_export_v1_XXXXXX,AGY_COST_BOARD_SEAT_QUOTA=50"
```

---

## 4. Securing & Accessing Your Cloud Run Service

By default, `agy-cost-board` should be deployed with **IAM authentication required** (`--no-allow-unauthenticated`) to protect enterprise telemetry and billing data:

- **Authenticated Local Proxy (`gcloud run services proxy`)**:
  ```bash
  gcloud run services proxy agy-cost-board --project=YOUR_GCP_PROJECT_ID --region=us-central1
  ```
- **Identity-Aware Proxy (IAP) & Cloud IAM**:
  Grant authorized FinOps leads, engineering managers, and developers the **Cloud Run Invoker** (`roles/run.invoker`) role in Cloud IAM or expose the service behind Google Cloud Identity-Aware Proxy (IAP).
- **Statusline CLI Bearer Token Support**:
  The dynamic `/statusline.sh` script automatically detects authenticated Cloud Run endpoints and attaches `Authorization: Bearer $(gcloud auth print-identity-token)`.

---

## 5. Complete Configuration Reference

`agy-cost-board` resolves configuration in the following precedence order (highest to lowest):
1. **CLI Flags** (`--project`, `--monitored-projects`, `--pool-mode`, etc.)
2. **Environment Variables** (`AGY_COST_BOARD_*` and Cloud Run `PORT`)
3. **Configuration File** (`.agy-cost-board.yaml`)
4. **Built-in Defaults**

| Environment Variable | YAML Key (`.agy-cost-board.yaml`) | CLI Flag | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `AGY_COST_BOARD_PROJECT_ID` | `project_id` | `--project` | `""` | Primary Google Cloud Project ID hosting BigQuery jobs (or active project filter) |
| `AGY_COST_BOARD_MONITORED_PROJECTS` | `monitored_projects` | `--monitored-projects` | `[]` | Comma-separated list (or YAML array) of monitored GCP Project IDs |
| `AGY_COST_BOARD_TELEMETRY_TABLE` | `telemetry_table` | `--telemetry-table` | `""` | Full BigQuery table path (`project.dataset.table`) for `InferenceResponseLog` |
| `AGY_COST_BOARD_BILLING_TABLE` | `billing_table` | `--billing-table` | `""` | Full BigQuery table path (`project.dataset.table`) for Cloud Billing export |
| `AGY_COST_BOARD_SEAT_QUOTA` | `seat_quota` | `--seat-quota` | `50` | Purchased Gemini Enterprise license seat quota |
| `AGY_COST_BOARD_SEAT_PRICE` | `seat_price` | `--seat-price` | `45.0` | Reference monthly seat price used for shelfware savings calculation |
| `AGY_COST_BOARD_LICENSE_LOCATIONS` | `license_locations` | `--license-locations` | `["global","eu","us"]` | Discovery Engine locations queried for license seats |
| `AGY_COST_BOARD_POOL_MODE` | `pool_mode` | `--pool-mode` | `"pooled"` | Weekly allowance pool mode (`pooled` or `per_user`) |
| `AGY_COST_BOARD_POOL_COUNT_BASIS` | `pool_count_basis` | `--pool-basis` | `"assigned"` | Weekly allowance pool seat basis (`assigned` or `purchased`) |
| `AGY_COST_BOARD_PAYG_ATTRIBUTION_MODE` | `payg_attribution_mode` | `--payg-mode` | `"fair_share"` | PAYG overage attribution mode (`fair_share` or `when_used`) |
| `AGY_COST_BOARD_PORT` / `PORT` | `port` | `--port` | `8080` | HTTP listen port for `agy-cost-board serve` |
| `AGY_COST_BOARD_DEMO` | `demo_mode` | `--demo` | `false` | Run with 100% offline synthetic multi-project demonstration data |
| `AGY_COST_BOARD_MCP_ENABLED` | `mcp_enabled` | `--mcp-enabled` | `true` | Enable or disable the Model Context Protocol (`2025-03-26`) server |
| `AGY_COST_BOARD_LOG_LEVEL` | `log_level` | `--log-level` | `"info"` | Log verbosity level (`debug`, `info`, `warn`, `error`) |
| `AGY_COST_BOARD_LOG_FORMAT` | `log_format` | `--log-format` | `"auto"` | Log output format (`auto`: Cloud Logging `json` on Cloud Run, `text` locally) |
