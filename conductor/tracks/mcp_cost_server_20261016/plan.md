# Implementation Plan: MCP Server for Cost & Quota Tools with Antigravity Plugin & UI Integration

## Phase 1: Configuration & Core MCP JSON-RPC Engine (`internal/mcp`) [checkpoint: dc0bd9e]
- [x] e3f3699 Task: Unit Tests for `_MCP_ENABLED` Configuration & MCP Protocol Engine (TDD Red)
    - [x] Add tests in `internal/config/config_test.go` verifying default `MCPEnabled: true` and overrides via `AGY_COST_BOARD_MCP_ENABLED=false`, `_MCP_ENABLED=false`, and `MCP_ENABLED=false`
    - [x] Create `internal/mcp/server_test.go` testing JSON-RPC 2.0 `initialize`, `notifications/initialized`, `ping`, `tools/list`, and `tools/call` across all 4 tools (`get_overview_metrics`, `get_license_status`, `get_users_costs`, `get_user_dashboard`)
    - [x] Test error handling for unknown methods, invalid tool arguments, and missing `user_id`
- [x] dc0bd9e Task: Implement Configuration Toggle & Core MCP Server (`internal/mcp`) (TDD Green)
    - [x] Add `MCPEnabled bool` to `config.Config` in `internal/config/config.go` with Viper env bindings (`AGY_COST_BOARD_MCP_ENABLED`, `_MCP_ENABLED`, `MCP_ENABLED`) and default `true`
    - [x] Implement `internal/mcp/server.go` with JSON-RPC 2.0 request/response types, tool schemas, and tool handlers backed by `attribution.Engine`
    - [x] Implement both `ServeStdio(ctx, io.Reader, io.Writer)` and `ServeHTTP(http.ResponseWriter, *http.Request)` on `mcp.Server`
    - [x] Verify all unit tests pass with >80% coverage
- [x] dc0bd9e Task: Conductor - User Manual Verification 'Phase 1: Configuration & Core MCP JSON-RPC Engine (`internal/mcp`)' (Protocol in workflow.md)

## Phase 2: HTTP `/mcp` Endpoint & CLI `agy-cost-board mcp` Subcommand [checkpoint: 2b01c9b]
- [x] ec09955 Task: HTTP & CLI Integration Tests for MCP Transports and `_MCP_ENABLED` Toggle (TDD Red)
    - [x] Add HTTP tests in `internal/server/mcp_test.go` verifying `POST /mcp` executes `initialize`, `tools/list`, and `tools/call` when `MCPEnabled` is `true`, and returns `404 Not Found` when `MCPEnabled` is `false`
    - [x] Assert `/api/v1/version` (or `/api/v1/mcp/status`) exposes `mcp_enabled` boolean for UI discovery
    - [x] Add CLI tests in `internal/cli/mcp_test.go` verifying `agy-cost-board mcp --demo` processes JSON-RPC messages over stdin/stdout and returns an error when `--mcp-enabled=false` or `_MCP_ENABLED=false`
- [x] 2b01c9b Task: Wire HTTP `/mcp` Route and `agy-cost-board mcp` Cobra Command (TDD Green)
    - [x] Update `internal/server/server.go` to mount `/mcp` when `MCPEnabled` is enabled and expose `mcp_enabled` in server version/status metadata
    - [x] Create `internal/cli/mcp.go` implementing `agy-cost-board mcp` stdio transport and register `--mcp-enabled` persistent flag in `internal/cli/root.go`
    - [x] Run Go test suite (`go test ./...`) and verify >80% coverage
- [x] 2b01c9b Task: Conductor - User Manual Verification 'Phase 2: HTTP `/mcp` Endpoint & CLI `agy-cost-board mcp` Subcommand' (Protocol in workflow.md)

## Phase 3: Antigravity Plugin Bundle & Web UI Quota/MCP Integration
- [x] e1246b6 Task: Frontend Component Tests & Plugin Manifest Validation (TDD Red)
    - [x] Create frontend tests in `web/src/components/McpPluginModal.test.tsx` (or equivalent integration view test) verifying rendering of MCP status badge (`Enabled`/`Disabled`), quota telemetry preview, and copyable `mcp_config.json` snippets for Streamable HTTP and Stdio
    - [x] Add test assertions verifying header button triggers the Antigravity Plugin / MCP Quota modal
- [~] Task: Scaffold Antigravity Plugin & Implement Web UI Integration (TDD Green)
    - [ ] Create `.agents/plugins/agy-cost-board/plugin.json`, `.agents/plugins/agy-cost-board/mcp_config.json`, `.agents/plugins/agy-cost-board/rules/AGENTS.md`, and `.agents/plugins/agy-cost-board/skills/cost-and-quota-advisor/SKILL.md`
    - [ ] Implement the Antigravity MCP & Quota Plugin integration component in `web/src/` and wire it into the main application header/dashboard
    - [ ] Update `README.md`, `conductor/product.md`, and `conductor/tech-stack.md` to document the MCP server, `_MCP_ENABLED` variable, and Antigravity plugin
    - [ ] Run full frontend and backend test suites and linters
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Antigravity Plugin Bundle & Web UI Quota/MCP Integration' (Protocol in workflow.md)
