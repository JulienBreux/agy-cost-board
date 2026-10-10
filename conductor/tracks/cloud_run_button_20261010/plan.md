# Implementation Plan: Cloud Run One-Click Deploy Button

## Phase 1: Cloud Run Button Configuration (`app.json` + Unit Tests) [checkpoint: 10c9617]
- [x] Task: Unit Tests for Cloud Run Button Configuration (TDD Red) [19a22a7]
    - [x] Create test in `internal/config` or `tests/` validating `app.json` exists at repo root
    - [x] Assert valid JSON syntax, required top-level attributes (`name`, `description`, `options`, `env`)
    - [x] Assert option constraints: `allow-unauthenticated: false`, `port: 8080`, `memory: "512Mi"`, `cpu: "1"`
    - [x] Assert environment variable mappings match configuration expectations
- [x] Task: Implement `app.json` Specification (TDD Green) [10c9617]
    - [x] Create `app.json` at repository root compliant with Cloud Run Button tech specification
    - [x] Configure `name`, `options`, and environment parameters (`PROJECT_ID`, `AGY_COST_BOARD_DEMO`, etc.)
    - [x] Run unit tests to verify green status
- [x] Task: Conductor - User Manual Verification 'Phase 1: Cloud Run Button Configuration (`app.json` + Unit Tests)' (Protocol in workflow.md) [10c9617]

## Phase 2: Documentation & One-Click Deployment Guide (`README.md`) [checkpoint: fc3cf86]
- [x] Task: Automated README Cloud Run Button Validation Test (TDD Red) [9bc43f3]
    - [x] Add test asserting `README.md` contains the official Cloud Run badge and target URL
- [x] Task: Update `README.md` with Badge & Deployment Guide (TDD Green) [fc3cf86]
    - [x] Add `[![Run on Google Cloud](https://deploy.cloud.run/button.svg)](https://deploy.cloud.run)` to top badge row
    - [x] Add dedicated section "Deploy to Cloud Run in One Click" detailing prerequisites, button action, parameters, and IAM access instructions
    - [x] Run test suite to verify green status
- [x] Task: Conductor - User Manual Verification 'Phase 2: Documentation & One-Click Deployment Guide (`README.md`)' (Protocol in workflow.md) [fc3cf86]

## Phase 3: End-to-End Build & Validation [checkpoint: 0532aaa]
- [x] Task: Validate Dockerfile & Cloud Run Compatibility [0532aaa]
    - [x] Verify root `Dockerfile` compatibility with Cloud Run Button build flow (multi-stage build, port binding)
    - [x] Run full test suite (`make test` or `go test ./...`) and lint checks (`make lint`)
- [x] Task: Conductor - User Manual Verification 'Phase 3: End-to-End Build & Validation' (Protocol in workflow.md) [0532aaa]
