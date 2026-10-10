# Implementation Plan: Cloud Run One-Click Deploy Button

## Phase 1: Cloud Run Button Configuration (`app.json` + Unit Tests)
- [x] Task: Unit Tests for Cloud Run Button Configuration (TDD Red) [19a22a7]
    - [x] Create test in `internal/config` or `tests/` validating `app.json` exists at repo root
    - [x] Assert valid JSON syntax, required top-level attributes (`name`, `description`, `options`, `env`)
    - [x] Assert option constraints: `allow-unauthenticated: false`, `port: 8080`, `memory: "512Mi"`, `cpu: "1"`
    - [x] Assert environment variable mappings match configuration expectations
- [ ] Task: Implement `app.json` Specification (TDD Green)
    - [ ] Create `app.json` at repository root compliant with Cloud Run Button tech specification
    - [ ] Configure `name`, `options`, and environment parameters (`PROJECT_ID`, `AGY_COST_BOARD_DEMO`, etc.)
    - [ ] Run unit tests to verify green status
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Cloud Run Button Configuration (`app.json` + Unit Tests)' (Protocol in workflow.md)

## Phase 2: Documentation & One-Click Deployment Guide (`README.md`)
- [ ] Task: Automated README Cloud Run Button Validation Test (TDD Red)
    - [ ] Add test asserting `README.md` contains the official Cloud Run badge and target URL
- [ ] Task: Update `README.md` with Badge & Deployment Guide (TDD Green)
    - [ ] Add `[![Run on Google Cloud](https://deploy.cloud.run/button.svg)](https://deploy.cloud.run)` to top badge row
    - [ ] Add dedicated section "Deploy to Cloud Run in One Click" detailing prerequisites, button action, parameters, and IAM access instructions
    - [ ] Run test suite to verify green status
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Documentation & One-Click Deployment Guide (`README.md`)' (Protocol in workflow.md)

## Phase 3: End-to-End Build & Validation
- [ ] Task: Validate Dockerfile & Cloud Run Compatibility
    - [ ] Verify root `Dockerfile` compatibility with Cloud Run Button build flow (multi-stage build, port binding)
    - [ ] Run full test suite (`make test` or `go test ./...`) and lint checks (`make lint`)
- [ ] Task: Conductor - User Manual Verification 'Phase 3: End-to-End Build & Validation' (Protocol in workflow.md)
