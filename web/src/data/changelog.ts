export interface ChangelogItem {
  title?: string;
  description: string;
}

export interface ChangelogRelease {
  version: string;
  date: string;
  title: string;
  items: ChangelogItem[];
}

export const CHANGELOG_DATA: ChangelogRelease[] = [
  {
    version: 'v0.10.3',
    date: 'October 18, 2026',
    title: 'Native Billing Currency Preservation Across Project Filters',
    items: [
      {
        title: 'Project-Scoped Currency Inheritance & BigQuery Currency Fallback',
        description:
          'Preserved native billing currency (EUR, GBP, JPY, CAD, AUD, USD) when filtering by a GCP project with zero project-scoped billing rows or when zero AI SKU rows match in the window by falling back to the billing account export currency.',
      },
      {
        title: 'Dynamic Currency Symbols in Forecast & Weekly Pool Cards',
        description:
          'Replaced remaining hardcoded USD labels in the Linear Regression Forecast unit toggle and Weekly Allowance Pool subtitle with dynamic currency symbols and formatting.',
      },
    ],
  },
  {
    version: 'v0.10.2',
    date: 'October 18, 2026',
    title: 'Project Filter Resilience & My Consumption Zero-Activity Support',
    items: [
      {
        title: 'Null-Safe Empty Project Cost Serialization & Frontend Guards',
        description:
          'Initialized attributed cost slices in the attribution engine so projects with zero telemetry logs serialize as [] instead of null, and added defensive Array.isArray guards across API clients and table views.',
      },
      {
        title: 'Project-Scoped My Consumption Dashboard',
        description:
          'Propagated the active project filter to the My Consumption dashboard and user activity feeds, with graceful zero-spend fallback when a developer has no activity in the selected GCP project.',
      },
    ],
  },
  {
    version: 'v0.10.1',
    date: 'October 18, 2026',
    title: 'BigQuery Schema Resilience, Cloud Logging & Distributed Tracing',
    items: [
      {
        title: 'Schema-Safe BigQuery Telemetry Extraction',
        description:
          'Fixed Cloud Logging BigQuery sink compatibility where resource.labels uses a global STRUCT<location, resource_container> schema without a .project_id subfield by using TO_JSON_STRING(...) + JSON_VALUE(...) with automatic resource_container extraction.',
      },
      {
        title: 'Google Cloud Logging & Cloud Trace Correlation',
        description:
          'Added structured JSON/text logging (log/slog) with automatic Cloud Run detection, X-Cloud-Trace-Context and W3C traceparent header propagation, X-Trace-Id / X-Request-Id response headers, and trace-correlated HTTP error payloads.',
      },
      {
        title: 'Span Instrumentation & Shared Telemetry Cache',
        description:
          'Instrumented BigQuery queries, attribution engine calculations, and MCP tool calls with child trace spans and introduced a shared raw-telemetry cache to eliminate redundant BigQuery queries on cold-cache overview requests.',
      },
    ],
  },
  {
    version: 'v0.10.0',
    date: 'October 18, 2026',
    title: 'FinOps Intelligence Suite: WTU, Shared Pool Simulator, Chargeback, Forecasting & Multi-Project',
    items: [
      {
        title: 'Weighted Token Units (WTU), Cache Hit Rate & Effort Telemetry',
        description:
          'Added Weighted Token Units (1x Input + 0.1x Cached + 3x Thinking + 5x Output), context caching efficiency and USD savings, thinking effort tier breakdowns (Minimal/Low/Medium/High), and client distribution across Web, CLI, TUI, and MCP.',
      },
      {
        title: 'Included Seat Cost vs. PAYG Overage & Weekly Allowance Pool Simulator',
        description:
          'Separated fixed Gemini Enterprise subscription costs from variable Vertex AI Pay-As-You-Go overages, with an interactive Weekly Allowance Pool gauge and What-If seat simulator (+5/+10/-5 seats).',
      },
      {
        title: 'Fair-Share vs. When-Used Chargeback & Multi-Tier Licenses',
        description:
          'Added live Fair-Share (proportional WTU) vs. When-Used (chronological post-exhaustion) PAYG attribution modes, CSV chargeback export, and multi-tier/multi-location license governance (Enterprise, Standard, Trial across us/eu/global).',
      },
      {
        title: '30-Day Window-Aware Spend Forecast & Cloud Monitoring Meter',
        description:
          'Added 30-day projected total spend (fixed renewal + daily PAYG velocity × 30) with confidence intervals, active days, optional Cloud Monitoring token_consumption cross-check badge, and multi-project selector.',
      },
      {
        title: 'Streamlined README & Modular Documentation Hub',
        description:
          'Refactored README.md into a concise landing page and introduced a comprehensive 6-part documentation suite in docs/ covering Getting Started, Architecture & FinOps, Interfaces & Reference, Integrations, and Deployment.',
      },
    ],
  },
  {
    version: 'v0.9.0',
    date: 'October 18, 2026',
    title: 'Developer Leaderboard & Two-Line Navigation Bar',
    items: [
      {
        title: 'Developer Consumption Leaderboard',
        description:
          'New Leaderboard view featuring Top 3 Podium cards, ranked developer spend and token tables, active day counts, cost efficiency per 1K tokens, and stacked model mix bars.',
      },
      {
        title: 'Refined Two-Line Navigation Bar',
        description:
          'Dedicated second navigation row housing Overview, Leaderboard, My Consumption, Attributed Costs, Seat Governance, and Setup & Health tabs while keeping brand and global controls on the top bar.',
      },
      {
        title: 'CLI & TUI Leaderboard Support',
        description:
          'Added the agy-cost-board leaderboard CLI subcommand (table, json, csv) and interactive [2] LEADERBOARD tab in the Bubble Tea terminal UI.',
      },
    ],
  },
  {
    version: 'v0.8.0',
    date: 'October 18, 2026',
    title: 'Unified AGY_COST_BOARD_ Environment Variable Standard',
    items: [
      {
        title: 'Strict AGY_COST_BOARD_ Prefix',
        description:
          'Standardized all configuration and statusline environment variables under the AGY_COST_BOARD_ prefix (including AGY_COST_BOARD_MCP_ENABLED, AGY_COST_BOARD_PROJECT_ID, and AGY_COST_BOARD_PORT).',
      },
      {
        title: 'Legacy Environment Variable Removal',
        description:
          'Removed unprefixed legacy environment variable fallbacks (such as _MCP_ENABLED, GCP_PROJECT, PROJECT_ID, PORT, and AGY_COST_*) for clean, collision-free configuration.',
      },
    ],
  },
  {
    version: 'v0.7.0',
    date: 'October 18, 2026',
    title: 'Model Context Protocol (MCP) Server & Antigravity Quota Plugin',
    items: [
      {
        title: 'Dual-Transport MCP Server',
        description:
          'Expose FinOps cost attribution, Gemini Enterprise seat quotas, and personal budget tools via Streamable HTTP (/mcp) and stdio (agy-cost-board mcp).',
      },
      {
        title: 'Runtime _MCP_ENABLED Toggle',
        description:
          'Control MCP server availability dynamically using the _MCP_ENABLED environment variable or --mcp-enabled CLI flag.',
      },
      {
        title: 'Antigravity Native Plugin',
        description:
          'Bundled .agents/plugins/agy-cost-board plugin with MCP configuration, FinOps governance rules, and the cost-and-quota-advisor skill.',
      },
      {
        title: 'Live Quota & MCP Setup Modal',
        description:
          'Interactive Navbar MCP status badge and modal displaying live Gemini Enterprise seat utilization, active developer quotas, and one-click Antigravity config snippets.',
      },
    ],
  },
  {
    version: 'v0.6.1',
    date: 'October 16, 2026',
    title: 'Navigation layout polish & responsive header alignment',
    items: [
      {
        title: 'Single-Line Navigation Tabs',
        description:
          'Prevent navigation items, brand header, and theme selector controls from wrapping across multiple lines on compact viewports.',
      },
      {
        title: 'Brand Header Alignment',
        description:
          'Improved vertical alignment and line-height spacing for the application title, subtitle, and horizontally scrollable navigation bar.',
      },
    ],
  },
  {
    version: 'v0.6.0',
    date: 'October 13, 2026',
    title: 'Native Currency Support & Release Changelog Popup',
    items: [
      {
        title: 'Multi-Currency Engine',
        description:
          'Automatically extract and resolve native ISO 4217 billing currencies (EUR, GBP, JPY, CAD, AUD, USD) directly from GCP Cloud Billing exports.',
      },
      {
        title: 'CLI & TUI Formatting',
        description:
          'Display dynamic currency symbols, localized separators, and zero-decimal currencies across terminal tables and interactive Bubble Tea dashboards.',
      },
      {
        title: 'Statusline Spend Script',
        description:
          'Serve customized developer statusline scripts with real-time currency formatting and subshell background refresh.',
      },
      {
        title: 'What\'s New Changelog Popup',
        description:
          'Instant notification popup highlighting new features and changes across versions with persistent cookie and localStorage tracking.',
      },
    ],
  },
  {
    version: 'v0.5.1',
    date: 'October 10, 2026',
    title: 'Version derivation and release tag links',
    items: [
      {
        title: 'Version Synchronization',
        description:
          'Dynamically load application build metadata and version info via the /api/v1/version endpoint.',
      },
      {
        title: 'Release Tag Resolution',
        description:
          'Cleanly parse semver core so snapshot and dirty builds link directly to corresponding GitHub release tags.',
      },
    ],
  },
  {
    version: 'v0.5.0',
    date: 'October 10, 2026',
    title: 'Dynamic footer with project details and repository links',
    items: [
      {
        title: 'Dynamic Footer',
        description:
          'Added responsive footer displaying GitHub repository links, build version, and author credits.',
      },
      {
        title: 'Theme Integration',
        description:
          'Polished dark and light theme styles with WCAG-compliant contrast ratios across all viewports.',
      },
    ],
  },
  {
    version: 'v0.4.0',
    date: 'October 10, 2026',
    title: 'Statusline integration help box and rendering preview',
    items: [
      {
        title: 'Statusline Integration Box',
        description:
          'Interactive configuration instructions and one-click shell snippets in the Personal Consumption tab.',
      },
      {
        title: 'Terminal Rendering Preview',
        description:
          'Live visual simulation of terminal status bar output showing developer token spend.',
      },
    ],
  },
  {
    version: 'v0.3.0',
    date: 'October 10, 2026',
    title: 'Multi-arch container releases and Cloud Run deployment',
    items: [
      {
        title: 'GHCR Container Registry',
        description:
          'Automated multi-architecture container images published to GitHub Packages for linux/amd64 and linux/arm64.',
      },
      {
        title: 'Cloud Run One-Click Deploy',
        description:
          'Distroless minimal container runtime with app.json manifest for instant browser-based provisioning.',
      },
    ],
  },
  {
    version: 'v0.2.0',
    date: 'October 9, 2026',
    title: 'Cross-platform statusline performance and caching',
    items: [
      {
        title: 'Subshell Asynchronous Refresh',
        description:
          'Zero terminal latency through non-blocking background subshell data fetching.',
      },
      {
        title: 'Cross-Platform File Cache',
        description:
          'Optimized cache file mtime evaluation compatible with macOS stat and Linux GNU stat.',
      },
    ],
  },
  {
    version: 'v0.1.0',
    date: 'October 9, 2026',
    title: 'Initial Release: AI Agent Cost & Usage Monitoring Board',
    items: [
      {
        title: 'Antigravity Cost Attribution',
        description:
          'Proportional mathematical cost attribution reconciling InferenceResponseLog telemetry with GCP billing invoices.',
      },
      {
        title: 'Gemini Enterprise License Governance',
        description:
          'Track assigned vs active seats and detect dormant users to optimize enterprise seat quotas.',
      },
      {
        title: 'Dual-Mode Architecture',
        description:
          'Execute rich CLI subcommands or serve an interactive React dashboard from a single Go binary.',
      },
    ],
  },
];
