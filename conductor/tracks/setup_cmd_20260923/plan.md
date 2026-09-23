# Implementation Plan: Setup Command & Telemetry Verification Engine (`setup_cmd_20260923`)

## Phase 1: Diagnostics & Verification Engine (`internal/setup`)
- [ ] Task: Diagnostics Domain Models & Checker Interface
  - [ ] Write unit tests for CheckResult, DiagnosticReport, and CheckStatus evaluation
  - [ ] Implement diagnostic types, status codes (OK, WARNING, ERROR), and report aggregators in `internal/domain`
- [ ] Task: GCP Environment & ADC Authentication Check
  - [ ] Write unit tests for ADC & gcloud credential detection
  - [ ] Implement ADC credential validator with project ID extraction and permission checking
- [ ] Task: BigQuery Telemetry Dataset & Cloud Logging Sink Checker
  - [ ] Write unit tests for BigQuery sink and dataset inspection
  - [ ] Implement checker querying Cloud Logging sink destination and BigQuery dataset existence
- [ ] Task: Cloud Billing Export Table Checker
  - [ ] Write unit tests for billing export dataset/table validation
  - [ ] Implement query probe verifying `gcp_billing_export_v1_*` presence and queryability
- [ ] Task: Telemetry Pipeline Validation Probe
  - [ ] Write unit tests for sample query executing token & user extraction
  - [ ] Implement validation probe querying for recent `InferenceResponseLog` records
- [ ] Task: Phase Verification & Checkpoint (Refer to workflow.md)

## Phase 2: CLI `setup` Subcommand & Configuration Persistence
- [ ] Task: Configuration File Manager (`.agy-ge-board.yaml`)
  - [ ] Write unit tests for config file reading, writing, and env/flag precedence
  - [ ] Implement configuration loader and writer in `internal/config`
- [ ] Task: CLI `setup` Command Implementation
  - [ ] Write unit tests for `agy-ge-board setup` with table/JSON output
  - [ ] Implement `setup` command with `--save`, `--dry-run`, and `--demo` flags
- [ ] Task: Resource Provisioning Generator (`--create` / `--dry-run`)
  - [ ] Write unit tests for `gcloud` command generation and dataset/sink provisioning logic
  - [ ] Implement BigQuery dataset creation and Cloud Logging sink creation helpers
- [ ] Task: Integrate Config Loader into All Existing CLI Commands
  - [ ] Write unit tests verifying CLI subcommands (`cost`, `license`, `user`, `serve`, `tui`) inherit settings from `.agy-ge-board.yaml`
  - [ ] Update `root.go` to auto-load configuration file if flags are omitted
- [ ] Task: Phase Verification & Checkpoint (Refer to workflow.md)

## Phase 3: REST API & Web Dashboard Integration
- [ ] Task: REST API `/api/v1/setup/status`
  - [ ] Write unit tests for HTTP GET `/api/v1/setup/status`
  - [ ] Implement Chi router endpoint returning diagnostic JSON report
- [ ] Task: React 19 Setup & Health Dashboard Tab
  - [ ] Implement `SetupView.tsx` component with interactive check cards, badges, and remediation commands
  - [ ] Integrate Setup tab into navigation bar and main view router
  - [ ] Rebuild frontend bundle (`npm run build`) and update `web/dist`
- [ ] Task: Phase Verification & Checkpoint (Refer to workflow.md)

## Phase 4: End-to-End Validation & Documentation
- [ ] Task: End-to-End Integration Tests
  - [ ] Write end-to-end integration tests for `setup` CLI and `/api/v1/setup/status` API endpoint
  - [ ] Verify test suite coverage remains $\ge 80\%$ with zero race conditions
- [ ] Task: Documentation & Setup Runbook Updates
  - [ ] Update `README.md` with `setup` command documentation, `--dry-run`, `--save`, and step-by-step verification guide
- [ ] Task: Phase Verification & Checkpoint (Refer to workflow.md)
