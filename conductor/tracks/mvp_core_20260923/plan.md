# Implementation Plan: MVP Track - Dual-Mode CLI & Cloud Run Web Dashboard

## Phase 1: Project Scaffolding & Core Domain Models
- [x] Task: Project Scaffolding & Dependency Initialization
  - [x] Initialize Go module (`agy-ge-board`) and standard directory layout (`cmd/`, `internal/`, `web/`)
  - [x] Configure linting and test tooling adhering to `workflow.md` and `code_styleguides/`
- [x] Task: Domain Models & Attribution Data Types
  - [x] Write unit tests for domain entities (`UserTokenShare`, `AllocatedUserCost`, `LicenseStatus`)
  - [x] Implement domain structs, JSON/CSV serialization, and attribution math validation
- [x] Task: Phase Verification & Checkpoint (Refer to workflow.md)

## Phase 2: BigQuery Client & Attribution Engine (TDD)
- [x] Task: Synthetic Mock / Demo Data Engine
  - [x] Write tests for synthetic telemetry and billing generator
  - [x] Implement `--demo` fixture engine generating 30 days of realistic multi-user/multi-model logs
- [x] Task: Proportional Cost Attribution Calculator
  - [x] Write unit tests for proportional cost formula, edge cases (zero tokens, missing SKUs), and TTL caching
  - [x] Implement attribution engine and in-memory thread-safe cache
- [x] Task: Live Google Cloud BigQuery Client
  - [x] Write unit tests for SQL query builders and credential validation
  - [x] Implement BigQuery client integration for Cloud Logging `InferenceResponseLog` and Cloud Billing exports
- [x] Task: Phase Verification & Checkpoint (Refer to workflow.md)

## Phase 3: CLI Subcommands & Interactive TUI (TDD)
- [x] Task: Cobra Command Hierarchy & Formatters
  - [x] Write unit tests for table, JSON, and CSV formatters
  - [x] Implement Cobra root command with `cost`, `license`, `user`, and `doctor` subcommands
- [x] Task: Bubbletea Interactive TUI
  - [x] Write unit tests for TUI state transitions, table sorting, and filtering
  - [x] Implement interactive terminal dashboard with Lipgloss styling
- [x] Task: Phase Verification & Checkpoint (Refer to workflow.md)

## Phase 4: Embedded Web Dashboard & REST API (TDD)
- [x] Task: Chi HTTP REST API Handlers
  - [x] Write unit tests for API endpoints (`/api/v1/metrics/overview`, `/api/v1/costs/users`, `/api/v1/licenses/status`)
  - [x] Implement Chi HTTP router, CORS/compression middleware, and graceful shutdown on SIGTERM
- [ ] Task: React Frontend SPA Development
  - [ ] Build React 18/19 SPA with Tailwind CSS, Recharts, and Lucide icons
  - [ ] Implement 3 core views: Overview KPIs, User Cost Attribution table, and License Governance
- [ ] Task: Binary Embedding & `serve` Subcommand
  - [ ] Write unit tests verifying embedded static file serving via Go `embed.FS`
  - [ ] Connect `serve` command to launch HTTP server on `$PORT` with fallback routing
- [ ] Task: Phase Verification & Checkpoint (Refer to workflow.md)

## Phase 5: Containerization & End-to-End Verification
- [ ] Task: Multi-Stage Dockerfile & Packaging
  - [ ] Create multi-stage Dockerfile (Node frontend build -> Go compiler -> distroless static runtime)
  - [ ] Verify container builds cleanly and runs locally under `$PORT`
- [ ] Task: End-to-End Validation & Documentation
  - [ ] Write end-to-end integration tests covering CLI outputs and API responses
  - [ ] Create comprehensive README with Cloud Run deployment guide and BigQuery log sink setup instructions
- [ ] Task: Phase Verification & Checkpoint (Refer to workflow.md)
