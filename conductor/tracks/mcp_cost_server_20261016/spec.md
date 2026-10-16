# Specification: MCP Server for Cost & Quota Tools with Antigravity Plugin & UI Integration

## 1. Overview
Implement a Model Context Protocol (MCP) server built directly into the `agy-cost-board` single binary, exposing cost attribution and Gemini Enterprise quota/license tools to AI agents (specifically Google Antigravity). The MCP server supports both **Streamable HTTP** (`POST /mcp` on `agy-cost-board serve`) and **Stdio** (`agy-cost-board mcp` CLI subcommand) transports, and can be toggled via the `AGY_COST_BOARD_MCP_ENABLED` / `_MCP_ENABLED` environment variable (`true` by default). Additionally, provide a ready-to-use **Antigravity Plugin** bundle and a **Web UI integration** displaying quota telemetry and one-click MCP/Plugin setup instructions.

## 2. Functional Requirements

### 2.1 Configuration & Environment Toggle (`_MCP_ENABLED`)
- Extend `config.Config` with `MCPEnabled bool` (`mapstructure:"mcp_enabled" yaml:"mcp_enabled" json:"mcp_enabled"`).
- Default `mcp_enabled` to `true`.
- Bind environment variables in Viper: `AGY_COST_BOARD_MCP_ENABLED`, `_MCP_ENABLED`, and `MCP_ENABLED`.
- Add persistent CLI flag `--mcp-enabled` (default `true`) on the root/serve/mcp commands.
- Expose `mcp_enabled` status in `/api/v1/version` or `/healthz` metadata so the Web UI can detect whether the HTTP MCP endpoint is active on the server.

### 2.2 Dual-Transport MCP Server (`internal/mcp`)
- Implement a lightweight, specification-compliant JSON-RPC 2.0 MCP server (`2025-03-26` Streamable HTTP & Stdio protocol) in `internal/mcp` without heavy external framework bloat:
  - Support standard MCP lifecycle methods: `initialize`, `notifications/initialized`, `ping`, `tools/list`, and `tools/call`.
  - **Streamable HTTP Transport (`/mcp`)**:
    - Mounted on the `chi` router in `internal/server/server.go` when `MCPEnabled` is `true`.
    - When `MCPEnabled` is `false`, requests to `/mcp` return `404 Not Found` (with JSON error `{"error": "MCP server is disabled"}`).
  - **Stdio Transport (`agy-cost-board mcp`)**:
    - New Cobra subcommand `agy-cost-board mcp` in `internal/cli/mcp.go` reading JSON-RPC messages line-by-line from `os.Stdin` and writing responses to `os.Stdout` (logging diagnostics solely to `os.Stderr`).
    - When `MCPEnabled` is `false`, `agy-cost-board mcp` exits immediately with an informative error on `stderr`.

### 2.3 MCP Cost & Quota Tools
All tools delegate to `attribution.Engine` and return structured JSON content inside MCP text content blocks (with `isError: true` on engine/validation errors):
1. `get_overview_metrics`:
   - **Inputs**: `days` (optional integer, default `30`).
   - **Output**: Project-wide total billed cost, currency, total tokens, active users, seat quota, and per-model cost breakdown.
2. `get_license_status`:
   - **Inputs**: `days` (optional integer, default `30`).
   - **Output**: Gemini Enterprise seat quota, assigned seats, active vs. dormant seats, utilization percentage, wasted monthly spend, and seat status list.
3. `get_users_costs`:
   - **Inputs**: `days` (optional integer, default `30`), `model` (optional string filter).
   - **Output**: Per-user proportional cost attribution records (`user_id`, `model`, `token_count`, `proportional_share`, `attributed_cost`, `currency`).
4. `get_user_dashboard`:
   - **Inputs**: `user_id` (required string), `days` (optional integer, default `30`), `monthly_budget` (optional number, default `150.0`).
   - **Output**: User consumption driving dashboard including total spend, budget burn rate, projected monthly spend, quota health status, model breakdown, efficiency tips, and recent activity.

### 2.4 Antigravity Plugin Bundle (`.agents/plugins/agy-cost-board/`)
- Create a native Antigravity plugin package under `.agents/plugins/agy-cost-board/`:
  - `plugin.json`: Plugin manifest (`{"name": "agy-cost-board"}`).
  - `mcp_config.json`: Pre-configured MCP server definition using `agy-cost-board mcp` (or Streamable HTTP URL template).
  - `rules/AGENTS.md`: Contextual instructions teaching Antigravity when and how to invoke the `agy-cost-board` MCP tools for cost and quota inquiries.
  - `skills/cost-and-quota-advisor/SKILL.md`: On-demand skill guiding the agent to inspect user/project quotas, analyze spend velocity, and render formatted quota progress bars and FinOps recommendations inside Antigravity.

### 2.5 Web UI Integration (Antigravity Plugin & Quota Display)
- Enhance the React Web UI (`web/src/`) with an **Antigravity Plugin & MCP Quota Integration** experience:
  - Add a dedicated trigger/modal or panel in the header/dashboard (alongside the CLI Statusline integration) showing:
    - Live MCP server status badge (`Enabled` / `Disabled` via `_MCP_ENABLED`).
    - Interactive preview of the Antigravity Quota & Spend widget/output.
    - Copy-pasteable `mcp_config.json` snippets for both **Remote Streamable HTTP** (`/mcp` on Cloud Run) and **Local Stdio** (`agy-cost-board mcp`), plus Antigravity plugin installation instructions.

## 3. Non-Functional Requirements
- **Zero Stdout Pollution in Stdio Mode**: `agy-cost-board mcp` must strictly reserve `stdout` for JSON-RPC 2.0 payloads.
- **High Test Coverage**: >80% unit test coverage across `internal/mcp`, `internal/config`, `internal/cli`, `internal/server`, and frontend components.
- **Demo Mode Compatibility**: Works seamlessly with `--demo` synthetic telemetry when GCP credentials are not configured.

## 4. Acceptance Criteria
- [ ] `AGY_COST_BOARD_MCP_ENABLED=false` or `_MCP_ENABLED=false` disables both `/mcp` HTTP endpoint and `agy-cost-board mcp` CLI command.
- [ ] `agy-cost-board mcp --demo` responds to `initialize`, `tools/list`, and `tools/call` for all 4 tools over `stdin`/`stdout`.
- [ ] `POST /mcp` on `agy-cost-board serve` responds to MCP JSON-RPC requests when enabled and returns `404` when disabled.
- [ ] `.agents/plugins/agy-cost-board/` contains valid `plugin.json`, `mcp_config.json`, `rules/AGENTS.md`, and `skills/cost-and-quota-advisor/SKILL.md`.
- [ ] Web UI displays MCP server status, quota widget preview, and copyable Antigravity MCP / Plugin configurations.

## 5. Out of Scope
- Legacy MCP HTTP+SSE (`2024-11-05`) stateful session transport (Antigravity uses stateless/Streamable HTTP on `/mcp` and `stdio`).
- Mutating GCP billing budgets or revoking IAM seats directly via MCP tools (tools are strictly read-only FinOps observability).
