# Product Guidelines: agy-cost-board

## 1. Visual & UI Design System
- **Aesthetic:** Modern, high-density FinOps dashboard inspired by Google Cloud Console clarity, utilizing Tailwind / modern CSS with seamless Dark and Light theme support.
- **Layout:** High information density without visual clutter. Prioritize KPI scorecards (total spend, active seats, idle licenses, avg cost per seat) above tabular and graphical views.
- **Color Palette:** Clean semantic colors indicating health:
  - Spend / Costs: Blue & Indigo accents
  - License Quotas: Emerald (healthy/available), Amber (near quota / threshold), Rose (over quota / budget alert)
  - Dark mode with high contrast ratios (WCAG AA compliant) for long developer and operator sessions.

## 2. CLI & Terminal User Experience (TUI)
- **Dual Terminal Experience:**
  - **Interactive TUI Mode:** Terminal dashboard with keyboard navigation (arrow keys, `j`/`k`, `/` filter, `q` quit), collapsible tree views, and live sorting.
  - **Scriptable Non-Interactive Mode:** Deterministic execution with `--json`, `--csv`, or `--plain` flags for CI/CD pipelines, cron jobs, and Unix pipes (`jq`, `awk`).
- **Telemetry & Feedback:** Progress spinners for BigQuery queries with execution timing, row count indicators, and clear dry-run summaries.

## 3. Voice, Tone & Educational Guidance
- **Instructive & Guiding:** Clear, accessible explanations rather than cryptic numbers. Demystify the proportional cost attribution formula:
  $$\text{User Cost} = \frac{\text{User Tokens}}{\text{Project Model Tokens}} \times \text{Net Invoiced SKU Cost}$$
- **Contextual Tooltips & Help Text:** Web UI tooltips and CLI `--help` explanations detailing Antigravity telemetry concepts (e.g., `InferenceResponseLog`, prompt caching, active vs assigned seats).
- **Consistent Nomenclature:** Use standard Google Cloud terminology (`BigQuery Dataset`, `Cloud Billing Export`, `Inference Telemetry`, `SKU`, `ADC`).

## 4. Diagnostics, Error Handling & Resilience
- **Actionable Diagnostic Errors:** When a BigQuery query fails due to missing IAM permissions, provide explicit remediation commands (e.g., `gcloud projects add-iam-policy-binding ... --role="roles/bigquery.dataViewer"`).
- **Proactive Prerequisites Check:** An `agy-cost-board doctor` / diagnostic check to validate GCP ADC credentials, BigQuery sink connectivity, and dataset accessibility.
- **Graceful Demo Fallback:** When GCP credentials are not detected or `--demo` is passed, offer instant demonstration mode with realistic synthetic datasets to explore the UI and TUI offline.

## 5. Performance & Resource Constraints
- Single-binary execution with minimal memory footprint (<50MB idle in Cloud Run).
- Fast cold-start latency (<1s) optimized for Cloud Run scale-to-zero.
- Smart client-side and server-side in-memory caching for repetitive BigQuery analytics queries.
