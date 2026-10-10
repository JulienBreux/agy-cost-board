# Implementation Plan: Antigravity CLI Statusline Spend Integration

## Phase 1: Statusline Template & HTTP Endpoint Serving
- [x] a890e6a Task: Unit & HTTP Tests for `/statusline.sh` Endpoint (TDD Red)
    - [ ] Create tests in `internal/server/statusline_test.go` requesting `GET /statusline.sh`
    - [ ] Assert status 200 OK and `Content-Type: text/x-shellscript; charset=utf-8`
    - [ ] Assert template substitution of server base URL (`Host`, `X-Forwarded-Proto`, or `X-Forwarded-Host`)
    - [ ] Assert template substitution of query parameters (`?user=...`, `?days=...`, `?ttl=...`)
- [x] 7a51dc3 Task: Implement Statusline Script Template & Handler (TDD Green)
    - [x] Embed the base `statusline.sh` template into `internal/server`
    - [x] Implement `handleStatuslineScript` in `internal/server/server.go` registering `GET /statusline.sh`
    - [x] Perform dynamic template substitution for host URL, pre-configured user, default days, and cache TTL
    - [x] Run unit tests to verify green status
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Statusline Template & HTTP Endpoint Serving' (Protocol in workflow.md)

## Phase 2: Statusline Script Logic, Cost Attribution & Caching
- [ ] Task: Integration & Script Execution Tests (TDD Red)
    - [ ] Create tests executing the rendered bash script against a mock HTTP server
    - [ ] Validate bash script syntax using `bash -n`
    - [ ] Validate simulated stdin JSON input from `agy` CLI
    - [ ] Assert output string contains ` · cost $...` at the end of the statusline
    - [ ] Validate cache file creation in `/tmp/` and background async refresh behavior
- [ ] Task: Implement Statusline Script Logic & Cost Formatting (TDD Green)
    - [ ] Integrate user spend resolution logic calling `/api/v1/users/{id}?days={days}`
    - [ ] Implement user detection hierarchy (`AGY_COST_USER` -> query param -> `gcloud` -> `git` -> `$USER`)
    - [ ] Implement zero-latency local caching in `/tmp/` with configurable `AGY_COST_CACHE_TTL`
    - [ ] Implement Cloud Run IAM authentication header support (`gcloud auth print-identity-token`)
    - [ ] Verify script execution tests pass
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Statusline Script Logic, Cost Attribution & Caching' (Protocol in workflow.md)

## Phase 3: Documentation & End-to-End Validation
- [ ] Task: Documentation & Quickstart Integration
    - [ ] Update `README.md` with "Antigravity CLI Statusline Integration" section
    - [ ] Include one-line `curl` download command and `agy` configuration snippet
    - [ ] Run full test suite (`make test`) and linter (`make lint`)
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Documentation & End-to-End Validation' (Protocol in workflow.md)
