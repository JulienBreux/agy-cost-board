# Integrations: Antigravity MCP Server, Plugin & CLI Statusline

`agy-cost-board` integrates directly into the **Antigravity IDE** and **Antigravity CLI (`agy`)** via a built-in **Model Context Protocol (MCP `2025-03-26`)** server, a ready-to-use **Antigravity Plugin** bundle, and a dynamic **CLI Statusline** script.

---

## 1. Antigravity MCP Server & Plugin Integration

`agy-cost-board` includes a built-in **Model Context Protocol (MCP)** server (`2025-03-26`) and a native **Antigravity Plugin** bundle (`.agents/plugins/agy-cost-board/`), allowing AI agents in Antigravity IDE and CLI to query real-time AI costs, Weekly Allowance Pool status, 30-day spillover forecasts, Cloud Monitoring quota meter drift, and Gemini Enterprise seat quotas directly.

### Exposed MCP Tools

| Tool Name | Arguments | Description |
| :--- | :--- | :--- |
| `get_overview_metrics` | `days?: number` (default `30`), `pool_mode?: "pooled"\|"per_user"`, `pool_basis?: "assigned"\|"purchased"`, `project?: string` | Organization-wide spend, WTU, Included vs. PAYG split, 7-day Weekly Allowance Pool (with Cloud Monitoring quota meter sync), 30-day window-aware forecast, token cache efficiency, model effort breakdown, and monitored projects. |
| `get_license_status` | `days?: number` (default `30`), `pool_mode?: "pooled"\|"per_user"`, `pool_basis?: "assigned"\|"purchased"`, `project?: string` | Gemini Enterprise seat quota, `Standard`/`Plus` tier breakdown, regional store distribution (`global`, `eu`, `us`), active/at-risk/dormant seats, and reclaimable license savings. |
| `get_users_costs` | `days?: number` (default `30`), `model?: string`, `payg_mode?: "fair_share"\|"when_used"`, `project?: string` | Proportional per-user, per-model, per-day cost attribution rows reconciled with Cloud Billing (`fair_share` or `when_used` PAYG chargeback). |
| `get_user_dashboard` | `user_id: string`, `days?: number`, `monthly_budget?: number` | Personal developer burn rate, projected month-end spend, budget pacing, cache hit rate, and FinOps model/effort optimization recommendations. |

### Dual Transport Modes

1. **Streamable HTTP (`POST /mcp`)**: Automatically mounted when running `agy-cost-board serve`:
   ```json
   {
     "mcpServers": {
       "agy-cost-board": {
         "url": "http://localhost:8080/mcp"
       }
     }
   }
   ```
2. **Stdio CLI (`agy-cost-board mcp`)**: Zero-port subprocess mode for local Antigravity sessions:
   ```json
   {
     "mcpServers": {
       "agy-cost-board": {
         "command": "agy-cost-board",
         "args": ["mcp", "--demo"],
         "env": {
           "AGY_COST_BOARD_MCP_ENABLED": "true"
         }
       }
     }
   }
   ```

### Disabling MCP (`AGY_COST_BOARD_MCP_ENABLED`)

MCP is **enabled by default** (`true`). To disable the MCP server (causing `POST /mcp` and `GET /mcp` to return `404 Not Found` and `agy-cost-board mcp` to exit cleanly), set `AGY_COST_BOARD_MCP_ENABLED=false`:

```bash
# Disable via environment variable
AGY_COST_BOARD_MCP_ENABLED=false ./bin/agy-cost-board serve --demo

# Or via CLI flag
./bin/agy-cost-board serve --mcp-enabled=false
```

### Antigravity Plugin Bundle (`.agents/plugins/agy-cost-board/`)

The repository includes a pre-packaged Antigravity plugin containing:
- `.agents/plugins/agy-cost-board/rules/AGENTS.md`: Governance rule instructing agents to query live telemetry before estimating spend or seat quotas.
- `.agents/plugins/agy-cost-board/skills/cost-and-quota-advisor/SKILL.md`: Specialized FinOps advisor skill for analyzing WTU consumption, dormant licenses, and personal budgets.

---

## 2. Antigravity CLI Statusline Spend Integration

Display individual developer AI spend directly in the Antigravity CLI status bar (e.g., ` · cost €18` or ` · cost $123`) using a dynamic, fail-open statusline script served directly by `agy-cost-board`.

<p align="center">
  <img src="assets/statusline.png" alt="Antigravity CLI Status Bar with Cost Attribution" width="100%" />
  <br />
  <em><b>Antigravity CLI Status Bar:</b> Live developer spend (<code>· cost €18</code> / <code>· cost $18</code>) displayed seamlessly at the end of the statusline.</em>
</p>

### One-Line Quick Install

Download and configure the script directly from your running `agy-cost-board` instance:

```bash
curl -sS https://<YOUR_AGY_COST_BOARD_URL>/statusline.sh -o ~/.gemini/antigravity-cli/statusline.sh && chmod +x ~/.gemini/antigravity-cli/statusline.sh
```

### Antigravity CLI Configuration

Add or update the `statusline` setting in `~/.gemini/antigravity-cli/settings.json`:

```json
{
  "statusline": "~/.gemini/antigravity-cli/statusline.sh"
}
```

Or export the environment variable in your shell profile (`~/.zshrc` or `~/.bashrc`):

```bash
export AGY_STATUSLINE="$HOME/.gemini/antigravity-cli/statusline.sh"
```

### Download Customization (Query Parameters)

| Parameter | Default | Description |
| :--- | :--- | :--- |
| `user` | *(auto)* | Pre-bake your developer email into the script (e.g., `?user=alex.rivera@acme.corp`) |
| `days` | `30` | Lookback window in days for spend aggregation (e.g., `?days=14`) |
| `ttl` | `300` | Local spend cache time-to-live in seconds (e.g., `?ttl=600` for 10 minutes) |

Example:
```bash
curl -sS "https://<YOUR_AGY_COST_BOARD_URL>/statusline.sh?user=alex.rivera@acme.corp&days=30&ttl=300" \
  -o ~/.gemini/antigravity-cli/statusline.sh && chmod +x ~/.gemini/antigravity-cli/statusline.sh
```

### User Identity Resolution Hierarchy

When executing, the statusline script resolves the developer's identity in order:
1. `AGY_COST_BOARD_USER` environment variable
2. Pre-baked `user` query parameter
3. Active Google Cloud account (`gcloud config get-value account`)
4. Git author email (`git config user.email`)
5. Local machine username (`$USER`)

### Runtime Environment Variable Overrides

| Environment Variable | Default | Description |
| :--- | :--- | :--- |
| `AGY_COST_BOARD_USER` | *(detected)* | Force a specific developer email or identity |
| `AGY_COST_BOARD_URL` | *(baked URL)* | Override the `agy-cost-board` server base URL |
| `AGY_COST_BOARD_DAYS` | `30` | Number of days for spend calculation |
| `AGY_COST_BOARD_CACHE_TTL` | `300` | Local cache TTL in seconds |
| `AGY_COST_BOARD_DISABLED` | `false` | Set to `true` or `1` to hide the cost suffix |

### Cloud Run IAM Authentication & Fail-Open Guarantee

- **Cloud Run IAM Bearer Auth**: When deployed to Cloud Run with `--no-allow-unauthenticated`, the script automatically obtains an identity token via `gcloud auth print-identity-token` and sends `Authorization: Bearer <token>`.
- **Zero Terminal Latency**: Spend is read synchronously from `/tmp/agy_cost_<hash>.cache` and refreshed asynchronously in a background subshell when the TTL expires.
- **Fail-Open Resilience**: If offline or if the server times out (`1s` timeout), the script outputs the standard Antigravity statusline without blocking or failing.
