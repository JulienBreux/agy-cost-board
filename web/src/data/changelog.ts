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
