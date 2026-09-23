# Specification: Setup & Telemetry Verification Engine (`setup_cmd_20260923`)

## 1. Overview
This track introduces the `setup` command (`agy-ge-board setup`), configuration persistence, and Web Dashboard integration. Based on the Google Cloud reference architecture article [*"Per-user cost attribution for Antigravity with BigQuery"*](https://medium.com/google-cloud/per-user-cost-attribution-for-antigravity-with-bigquery-3e98fd997c58), this capability allows users and administrators to verify, diagnose, and optionally provision all Google Cloud prerequisites needed for Antigravity inference logging and billing attribution.

## 2. Functional Requirements

### 2.1 CLI `setup` Subcommand
- **Prerequisite & Environment Inspection:**
  - Check Google Cloud Application Default Credentials (ADC) or active `gcloud` configuration.
  - Detect and validate active GCP project ID.
  - Verify IAM permissions for calling identity (`roles/bigquery.jobUser`, `roles/bigquery.dataViewer`, `roles/logging.configWriter`).
- **BigQuery Telemetry & Logging Sink Verification:**
  - Verify existence of BigQuery dataset for Antigravity telemetry.
  - Verify Cloud Logging sink filter targeting `InferenceResponseLog` (`jsonPayload.log_type="InferenceResponseLog"`).
  - Verify sink destination, service account writer identity, and dataset write permissions.
- **Cloud Billing Export Verification:**
  - Verify GCP Cloud Billing export table (`gcp_billing_export_v1_*`) existence and schema queryability.
- **Sample Attribution Pipeline Execution:**
  - Run a lightweight validation query checking for existing `InferenceResponseLog` records in BigQuery.
  - Validate model token extraction (`labels.user_id`, `labels.model`, `metadata.totalTokenCount`).
- **Resource Provisioning Mode (`--create` / `--dry-run`):**
  - `--create`: Automatically create missing BigQuery datasets and Cloud Logging sink.
  - `--dry-run`: Output the exact `gcloud` CLI commands and BigQuery DDL without executing changes.
- **Local Configuration Persistence (`--save` / auto-save):**
  - Save verified configuration to `.agy-ge-board.yaml` (or `.env`).
  - Automatically load `.agy-ge-board.yaml` across all commands (`cost`, `license`, `user`, `serve`, `tui`) with precedence: Flags > Env Vars > Config File > Demo Fallback.

### 2.2 REST API & Embedded Web Dashboard Integration
- **REST Endpoint:**
  - `GET /api/v1/setup/status`: Returns JSON diagnostics, check states (`ok`, `warning`, `error`), and remediation snippets.
- **Web Dashboard Setup Tab:**
  - Dedicated **Setup & Health** tab in the React 19 SPA.
  - Interactive status cards for ADC, BigQuery sinks, Billing export, and Telemetry ingest.
  - 1-click re-check button and copyable `gcloud` fix commands.

## 3. Non-Functional Requirements & Safety
- **Safe & Idempotent:** Provisioning never overwrites existing datasets or sinks; repeated execution is safe.
- **Fast Execution:** Concurrent checks completing within 3-5 seconds.
- **Demo Mode Support:** `setup --demo` returns 100% green simulated health checks for offline local evaluation.
- **High Test Coverage:** Target $\ge 80\%$ statement coverage with unit and integration tests.

## 4. Acceptance Criteria
1. `agy-ge-board setup --demo` outputs formatted diagnostic checklist with all checks passing.
2. `agy-ge-board setup --dry-run` outputs exact `gcloud` commands to configure sink and dataset.
3. `agy-ge-board setup --save` creates `.agy-ge-board.yaml` respected by all subcommands.
4. `GET /api/v1/setup/status` returns full diagnostic report in JSON.
5. React web dashboard includes Setup & Health tab displaying interactive status cards and remediation snippets.
6. Test coverage $\ge 80\%$ across new code.
