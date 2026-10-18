---
name: cost-and-quota-advisor
description: Queries the agy-cost-board MCP server to inspect organization AI costs, Gemini Enterprise seat quotas, dormant licenses, and personal developer consumption budgets. Activate when the user asks about AI spend, token usage, seat quotas, or FinOps optimization.
---

# Cost & Quota Advisor Skill (`agy-cost-board`)

Use this skill to inspect real-time Antigravity and Gemini Enterprise cost attribution and seat quota telemetry via the `agy-cost-board` Model Context Protocol (MCP) server.

## Available MCP Tools

1. **`get_overview_metrics`** (`days?: number`)
   - Returns organization-wide total billed cost, total tokens, active developer count, seat quota utilization percentage, daily spend trends, and per-model cost breakdown.
2. **`get_license_status`** (`days?: number`)
   - Returns Gemini Enterprise seat quota (`seat_quota`), `assigned_seats`, `active_seats`, `dormant_seats`, `utilization_pct`, `estimated_monthly_savings`, and the list of dormant users eligible for seat reclamation.
3. **`get_users_costs`** (`days?: number`, `model?: string`)
   - Returns proportional per-user, per-model, per-day cost attribution rows reconciled between BigQuery inference telemetry and GCP Billing exports.
4. **`get_user_dashboard`** (`user_id: string`, `days?: number`, `monthly_budget?: number`)
   - Returns personal consumption driving KPIs for a developer (`daily_burn_rate`, `weekly_burn_rate`, `projected_month_end_spend`, `budget_consumed_pct`, `budget_status`) and FinOps optimization recommendations.

## Recommended Response Format

When presenting quota or cost status to the user in Antigravity IDE or CLI, render a structured summary:

```markdown
### 📊 Antigravity & Gemini Enterprise Quota Status (30d)
| Metric | Value | Status |
| :--- | :--- | :--- |
| **Seat Quota Utilization** | `8 / 10 seats` (`████████░░ 80%`) | 🟢 Healthy |
| **Dormant Seats** | `1 seat` (Est. savings: `€38.00/mo`) | 🟡 Reclaimable |
| **Total Attributed Spend** | `€420.50` (`15.4M tokens`) | 🟢 On Track |
```

## Troubleshooting

- If MCP tools are unavailable, verify that `AGY_COST_BOARD_MCP_ENABLED` is not set to `false` in the server environment or `mcp_config.json`.
- To connect to a remote Cloud Run instance over Streamable HTTP instead of local `stdio`, configure `~/.gemini/antigravity-cli/mcp_config.json`:
  ```json
  {
    "mcpServers": {
      "agy-cost-board": {
        "url": "https://<your-cloud-run-url>/mcp"
      }
    }
  }
  ```
