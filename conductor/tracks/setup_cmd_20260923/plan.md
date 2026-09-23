# Implementation Plan: Setup Command & Telemetry Verification Engine (`setup_cmd_20260923`)

## Phase 1: Diagnostics & Verification Engine (`internal/setup`)
- [x] Task: Diagnostics Domain Models & Checker Interface
  - [x] Write unit tests for CheckResult, DiagnosticReport, and CheckStatus evaluation
  - [x] Implement diagnostic types, status codes (OK, WARNING, ERROR), and report aggregators in `internal/domain`
- [x] Task: GCP Environment & ADC Authentication Check
  - [x] Write unit tests for ADC & gcloud credential detection
  - [x] Implement ADC credential validator with project ID extraction and permission checking
- [x] Task: BigQuery Telemetry Dataset & Cloud Logging Sink Checker
  - [x] Write unit tests for BigQuery sink and dataset inspection
  - [x] Implement checker querying Cloud Logging sink destination and BigQuery dataset existence
- [x] Task: Cloud Billing Export Table Checker
  - [x] Write unit tests for billing export dataset/table validation
  - [x] Implement query probe verifying `gcp_billing_export_v1_*` presence and queryability
- [x] Task: Telemetry Pipeline Validation Probe
  - [x] Write unit tests for sample query executing token & user extraction
  - [x] Implement validation probe querying for recent `InferenceResponseLog` records
- [x] Task: Phase Verification & Checkpoint (Refer to workflow.md)

## Phase 2: CLI `setup` Subcommand & Configuration Persistence
- [x] Task: Configuration File Manager (`.agy-ge-board.yaml`)
  - [x] Write unit tests for config file reading, writing, and env/flag precedence
  - [x] Implement configuration loader and writer in `internal/config`
- [x] Task: CLI `setup` Command Implementation
  - [x] Write unit tests for `agy-ge-board setup` with table/JSON output
  - [x] Implement `setup` command with `--save`, `--dry-run`, and `--demo` flags
- [x] Task: Resource Provisioning Generator (`--create` / `--dry-run`)
  - [x] Write unit tests for `gcloud` command generation and dataset/sink provisioning logic
  - [x] Implement BigQuery dataset creation and Cloud Logging sink creation helpers
- [x] Task: Integrate Config Loader into All Existing CLI Commands
  - [x] Write unit tests verifying CLI subcommands (`cost`, `license`, `user`, `serve`, `tui`) inherit settings from `.agy-ge-board.yaml`
  - [x] Update `root.go` to auto-load configuration file if flags are omitted
- [x] Task: Phase Verification & Checkpoint (Refer to workflow.md)

## Phase 3: REST API & Web Dashboard Integration
- [x] Task: REST API `/api/v1/setup/status`
  - [x] Write unit tests for HTTP GET `/api/v1/setup/status`
  - [x] Implement Chi router endpoint returning diagnostic JSON report
- [x] Task: React 19 Setup & Health Dashboard Tab
  - [x] Implement `SetupView.tsx` component with interactive check cards, badges, and remediation commands
  - [x] Integrate Setup tab into navigation bar and main view router
  - [x] Rebuild frontend bundle (`npm run build`) and update `web/dist`
- [x] Task: Phase Verification & Checkpoint (Refer to workflow.md)

## Phase 4: End-to-End Validation & Documentation
- [x] Task: End-to-End Integration Tests
  - [x] Write end-to-end integration tests for `setup` CLI and `/api/v1/setup/status` API endpoint
  - [x] Verify test suite coverage remains $\ge 80\%$ with zero race conditions
- [x] Task: Documentation & Setup Runbook Updates
  - [x] Update `README.md` with `setup` command documentation, `--dry-run`, `--save`, and step-by-step verification guide
- [x] Task: Phase Verification & Checkpoint (Refer to workflow.md)
