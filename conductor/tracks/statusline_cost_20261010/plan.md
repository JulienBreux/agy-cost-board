# Implementation Plan: Antigravity CLI Statusline Spend Integration

## Phase 1: Statusline Template & HTTP Endpoint Serving [checkpoint: 7a51dc3]
- [x] a890e6a Task: Unit & HTTP Tests for `/statusline.sh` Endpoint (TDD Red)
    - [x] Create tests in `internal/server/statusline_test.go` requesting `GET /statusline.sh`
    - [x] Assert status 200 OK and `Content-Type: text/x-shellscript; charset=utf-8`
    - [x] Assert template substitution of server base URL (`Host`, `X-Forwarded-Proto`, or `X-Forwarded-Host`)
    - [x] Assert template substitution of query parameters (`?user=...`, `?days=...`, `?ttl=...`)
- [x] 7a51dc3 Task: Implement Statusline Script Template & Handler (TDD Green)
    - [x] Embed the base `statusline.sh` template into `internal/server`
    - [x] Implement `handleStatuslineScript` in `internal/server/server.go` registering `GET /statusline.sh`
    - [x] Perform dynamic template substitution for host URL, pre-configured user, default days, and cache TTL
    - [x] Run unit tests to verify green status
- [x] 7a51dc3 Task: Conductor - User Manual Verification 'Phase 1: Statusline Template & HTTP Endpoint Serving' (Protocol in workflow.md)

## Phase 2: Statusline Script Logic, Cost Attribution & Caching [checkpoint: 0a2c07f]
- [x] a4f03b8 Task: Integration & Script Execution Tests (TDD Red)
    - [x] Create tests executing the rendered bash script against a mock HTTP server
    - [x] Validate bash script syntax using `bash -n`
    - [x] Validate simulated stdin JSON input from `agy` CLI
    - [x] Assert output string contains ` · cost $...` at the end of the statusline
    - [x] Validate cache file creation in `/tmp/` and background async refresh behavior
- [x] 0a2c07f Task: Implement Statusline Script Logic & Cost Formatting (TDD Green)
    - [x] Integrate user spend resolution logic calling `/api/v1/users/{id}?days={days}`
    - [x] Implement user detection hierarchy (`AGY_COST_USER` -> query param -> `gcloud` -> `git` -> `$USER`)
    - [x] Implement zero-latency local caching in `/tmp/` with configurable `AGY_COST_CACHE_TTL`
    - [x] Implement Cloud Run IAM authentication header support (`gcloud auth print-identity-token`)
    - [x] Verify script execution tests pass
- [x] 0a2c07f Task: Conductor - User Manual Verification 'Phase 2: Statusline Script Logic, Cost Attribution & Caching' (Protocol in workflow.md)

## Phase 3: Documentation & End-to-End Validation
- [x] 06937f2 Task: Documentation & Quickstart Integration
    - [x] Update `README.md` with "Antigravity CLI Statusline Integration" section
    - [x] Include one-line `curl` download command and `agy` configuration snippet
    - [x] Run full test suite (`make test`) and linter (`make lint`)
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Documentation & End-to-End Validation' (Protocol in workflow.md)
