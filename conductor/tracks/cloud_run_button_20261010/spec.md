# Specification: Cloud Run One-Click Deploy Button

## 1. Overview
Implement a one-click deployment experience for Google Cloud Run using the official Google Cloud Run Button technology specification (https://github.com/GoogleCloudPlatform/cloud-run-button). This enables developers and platform administrators to launch an instance of `agy-cost-board` directly into their Google Cloud Project with a single click from the GitHub repository, pre-configured with secure IAM authentication and standard runtime parameters.

## 2. Functional Requirements

### 2.1 Cloud Run Button Configuration (`app.json`)
- Add an `app.json` configuration file at the repository root compliant with the Cloud Run Button specification:
  - **name:** `agy-cost-board`
  - **description:** "Per-user AI cost attribution and seat management dashboard for Google Antigravity & Gemini Enterprise"
  - **options:**
    - `allow-unauthenticated`: `false` (enforce IAM authentication for enterprise security)
    - `port`: `8080` (matches internal Go HTTP server port and Dockerfile `ENV PORT=8080`)
    - `memory`: `"512Mi"` (efficient baseline footprint)
    - `cpu`: `"1"`
  - **env:**
    - `PROJECT_ID`: Description "Google Cloud Project ID hosting BigQuery telemetry and billing tables", required: `false`
    - `AGY_COST_BOARD_DEMO`: Description "Enable demo mode with simulated cost and telemetry data (true/false)", value: `"false"`, required: `false`
    - `AGY_COST_BOARD_TELEMETRY_TABLE`: Description "BigQuery telemetry table path (project.dataset.table)", required: `false`
    - `AGY_COST_BOARD_BILLING_TABLE`: Description "BigQuery billing export table path (project.dataset.table)", required: `false`

### 2.2 Documentation & Quickstart Integration (`README.md`)
- Add the official Cloud Run Button badge to the repository header badge row:
  - Markdown: `[![Run on Google Cloud](https://deploy.cloud.run/button.svg)](https://deploy.cloud.run)`
  - URL target: `https://deploy.cloud.run` (fork-friendly automatic referrer detection)
- Add a dedicated **Deploy to Cloud Run in One Click** section in `README.md`:
  - Visual button display.
  - Prerequisites checklist (Google Cloud Project, billing enabled, Cloud Run & Cloud Build APIs enabled, required BigQuery viewer permissions).
  - Walkthrough of deployment prompts configured via `app.json`.
  - Secure access guide explaining how to invoke or access the IAM-protected Cloud Run service using `gcloud run services proxy` or authorized browser credentials.

### 2.3 Automated Validation & Verification
- Unit test in Go to validate:
  - `app.json` file existence and JSON parsing integrity.
  - Sizing options (`memory`, `cpu`, `port`, `allow-unauthenticated: false`).
  - Environment variable definitions and matching application configuration bindings.
- Documentation link/syntax test ensuring badge markdown syntax and target URL are well-formed.

## 3. Non-Functional Requirements
- **Security:** Cloud Run service defaults to private access (`allow-unauthenticated: false`), adhering to enterprise security baselines.
- **Portability:** Use referrer-based button URL `https://deploy.cloud.run` so downstream forks and mirrors deploy cleanly without hardcoded repository URLs.
- **Maintainability:** Full automated test coverage ensuring future updates to `app.json` or `README.md` do not break Cloud Run Button compatibility.

## 4. Acceptance Criteria
- [ ] `app.json` exists at root and conforms to Cloud Run Button schema.
- [ ] Service options specify IAM authentication (`allow-unauthenticated: false`), 512Mi memory, 1 vCPU, and port 8080.
- [ ] Configurable environment variables (`PROJECT_ID`, `AGY_COST_BOARD_DEMO`, `AGY_COST_BOARD_TELEMETRY_TABLE`, `AGY_COST_BOARD_BILLING_TABLE`) are defined with descriptive text.
- [ ] `README.md` features the Cloud Run Button badge in both the header and a dedicated quickstart deployment section.
- [ ] Automated tests validate `app.json` schema and README links.
- [ ] Docker build and application test suite pass without error.
