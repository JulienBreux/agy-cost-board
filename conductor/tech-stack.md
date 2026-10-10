# Technology Stack: agy-cost-board

## 1. Core Architecture: Single-Binary Dual Mode
- **Language:** Go (Golang 1.23+)
- **Distribution:** Single statically-linked binary embedding all frontend assets via Go `embed.FS`, executable on macOS/Linux/Windows and packaged as a minimal scratch/distroless container for Google Cloud Run.

## 2. CLI & Terminal Interface
- **Command Router:** `github.com/spf13/cobra` for command hierarchy (`serve`, `cost`, `license`, `user`, `doctor`).
- **Terminal UI (TUI):** `github.com/charmbracelet/bubbletea` and `github.com/charmbracelet/lipgloss` for interactive keyboard-driven dashboards, tables, and spinners.
- **Output Formatting:** Native support for ANSI color tables, `--json`, and `--csv` for scriptable pipelines.

## 3. Web Backend & APIs
- **HTTP Routing:** `github.com/go-chi/chi/v5` - lightweight, idiomatic, fast HTTP router.
- **Asset Embedding:** Go standard library `embed` package bundling the compiled React SPA.
- **Configuration & Environment:** Support for standard Cloud Run `$PORT`, standard GCP environment variables, and YAML/flag overrides.

## 4. Frontend Web Application
- **Framework & Build:** React with TypeScript, bundled using Vite.
- **Styling:** Tailwind CSS with first-class Dark and Light theme support matching modern FinOps aesthetics.
- **Data Visualizations:** Recharts for proportional spend charts, daily cost trends, and license quota donut meters.
- **Icons:** `lucide-react`.

## 5. Data Engine & Cloud Integration
- **BigQuery Client:** `cloud.google.com/go/bigquery` with Google Cloud Application Default Credentials (ADC) and Workload Identity.
- **Data Allocation Model:** Proportional attribution SQL joining Cloud Logging `InferenceResponseLog` telemetry with Cloud Billing export tables.
- **Caching:** Thread-safe in-memory TTL cache to eliminate redundant BigQuery queries and minimize analysis costs.
- **Mock / Demo Engine:** Built-in synthetic telemetry generator enabling full local execution and testing without GCP credentials (`--demo`).

## 6. Containerization & Deployment
- **Container Build:** Multi-stage Dockerfile (Stage 1: Node.js frontend build; Stage 2: Go compiler; Stage 3: `gcr.io/distroless/static` or Alpine minimal runtime).
- **Deployment Target:** Google Cloud Run (stateless, scale-to-zero, HTTPS termination, IAM authenticated).
- **One-Click Deployment:** Cloud Run Button integration (`app.json`) enabling browser-based provisioning via `https://deploy.cloud.run` with configurable environment variables and secure-by-default IAM authentication.

## 7. Developer Tooling & Integrations
- **Antigravity CLI Statusline:** Dynamic POSIX/bash 3.2+ compatible statusline script served via `GET /statusline.sh`, with zero-latency `/tmp` caching, background asynchronous subshell refresh, stampede locking, and Cloud Run IAM Bearer token authentication.
