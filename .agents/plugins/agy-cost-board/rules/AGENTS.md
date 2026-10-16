# Antigravity Cost & Seat Quota Governance Rule

When the user asks about AI inference spend, token consumption, Gemini Enterprise seat quotas, dormant seats, or personal budget pacing:

1. **Prefer Live Telemetry via MCP**: Always query the `agy-cost-board` MCP server tools before estimating costs or seat usage:
   - `get_overview_metrics`: High-level organization spend, total tokens, active developers, seat utilization, and model distribution.
   - `get_license_status`: Gemini Enterprise seat quota, assigned/active/dormant seats, and potential license savings.
   - `get_users_costs`: Per-developer and per-model proportional cost allocations.
   - `get_user_dashboard`: Personal consumption driving metrics, burn rate, monthly budget status, and model optimization recommendations for a specific `user_id`.
2. **Quota & Cost Presentation**: Format quota and cost summaries using clear Markdown tables with visual progress bars (e.g., `████████░░ 80%`) and actionable optimization insights.
3. **Configuration Toggle**: Remember that the MCP server can be toggled at runtime via `_MCP_ENABLED=true|false` (or `--mcp-enabled`).
