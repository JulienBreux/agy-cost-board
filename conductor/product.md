# Product Definition: agy-cost-board

## Initial Concept
I want to create an application based on this article: https://medium.com/google-cloud/per-user-cost-attribution-for-antigravity-with-bigquery-3e98fd997c58.
The idea is to build a web application that can be deployed to Cloud Run (a Docker container) and also to be able to interact with it via the CLI, with the same features in the same binary.
The main goal is to track the allocation of Gemini Enterprise licenses for a project, as well as the costs per user.

## Executive Summary & Vision
`agy-cost-board` is a single-binary cloud solution delivering both a modern web dashboard (deployable seamlessly to Cloud Run) and an expressive CLI for FinOps, engineering managers, and platform teams. It provides transparent, per-user cost attribution for Antigravity developer usage and tracks Google Gemini Enterprise license allocations. By reconciling Cloud Logging inference telemetry (`InferenceResponseLog` token shares) with Google Cloud Billing BigQuery exports using a proportional allocation model, `agy-cost-board` eliminates AI cost opacity while identifying unused licenses and budget overruns.

## Target Personas
1. **Platform & FinOps Engineers:** Oversee Google Cloud AI spend across projects, reconcile invoiced billing SKUs against actual developer consumption, establish budget alert thresholds, and detect idle or wasted seats.
2. **Engineering Managers & Leads:** Track team-level consumption trends, understand project burn rates, optimize license assignments, and forecast future AI capacity requirements.
3. **Developers:** Inspect individual token usage metrics and verify active license status directly from terminal or web interface.

## Key Capabilities & Features

### 1. Antigravity Inference Telemetry & Token Accounting
- Ingests and analyzes `InferenceResponseLog` records from BigQuery log sinks.
- Extracts user identities (`labels.user_id`), model classifications (`labels.model`), and token volumes (`metadata.totalTokenCount`).
- Calculates daily and per-model proportional token shares across projects.

### 2. Proportional Billing Reconciliation & Per-User Cost Attribution
- Integrates with GCP Cloud Billing export datasets in BigQuery.
- Matches net billed costs for Gemini and Antigravity SKUs against user token proportions, guaranteeing 100% mathematical reconciliation with the GCP invoice.
- Surfaces historical trends, average cost per active user, and cost breakdowns by model (e.g., Gemini 1.5 Pro vs. Flash).

### 3. Gemini Enterprise License Management
- Tracks assigned vs. consumed Gemini Enterprise seats per project.
- Identifies inactive or dormant license holders to enable seat reclamation and cost optimization.
- Generates threshold alerts when project allocations or user budgets approach predetermined limits.

### 4. Unified Dual-Mode Architecture (Single Binary)
- **CLI Subcommands:** Provides rapid commands such as `serve`, `cost`, `license`, `user`, and `init` with formatted table, JSON, and CSV outputs.
- **Web Interface (Cloud Run):** Triggered via `serve` (binding to `$PORT`), offering an intuitive analytics dashboard with charts, KPI summary cards, and user drill-downs.
- Packaged as a lightweight, scratch/distroless Docker container ready for Google Cloud Run deployment.

### 5. Seamless Ingestion & Local Developer Experience
- **Authentication:** Native support for Google Cloud Application Default Credentials (ADC) and Workload Identity Federation.
- **SQL / Infrastructure Automation:** Pre-packaged SQL views and DDL/sinks setup helpers.
- **Mock / Demo Fixture Mode:** Built-in synthetic datasets allowing instant offline local testing and UI demonstration without GCP billing credentials.
