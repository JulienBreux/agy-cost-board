import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { McpPluginModal } from './McpPluginModal';
import { LicenseGovernance, OverviewMetrics } from '../api';

const mockOverview: OverviewMetrics = {
  total_billed_cost: 420.5,
  total_tokens: 15400000,
  active_users_count: 8,
  seat_quota: 10,
  assigned_seats: 9,
  utilization_pct: 80.0,
  estimated_monthly_savings: 38.0,
  daily_trends: [],
  model_breakdown: {},
  currency: 'EUR',
};

const mockGovernance: LicenseGovernance = {
  seat_quota: 10,
  assigned_seats: 9,
  active_seats: 8,
  dormant_seats: 1,
  utilization_pct: 80.0,
  estimated_monthly_savings: 38.0,
  dormant_users: [],
  currency: 'EUR',
};

describe('McpPluginModal Component', () => {
  it('does not render when isOpen is false', () => {
    render(
      <McpPluginModal
        isOpen={false}
        onClose={vi.fn()}
        mcpEnabled={true}
        overview={mockOverview}
        governance={mockGovernance}
      />
    );

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('renders MCP active badge, quota telemetry widget, tools list, and config snippets when open', () => {
    render(
      <McpPluginModal
        isOpen={true}
        onClose={vi.fn()}
        mcpEnabled={true}
        overview={mockOverview}
        governance={mockGovernance}
      />
    );

    expect(screen.getByRole('dialog')).toBeInTheDocument();
    expect(screen.getByText(/Antigravity Plugin & MCP Server/i)).toBeInTheDocument();
    expect(screen.getByTestId('mcp-status-badge')).toHaveTextContent(/Enabled/i);

    // Quota & Seat telemetry preview
    expect(screen.getByText(/8 \/ 10/)).toBeInTheDocument();
    expect(screen.getByText(/80%/)).toBeInTheDocument();

    // MCP Tools list
    expect(screen.getByText('get_overview_metrics')).toBeInTheDocument();
    expect(screen.getByText('get_license_status')).toBeInTheDocument();
    expect(screen.getByText('get_users_costs')).toBeInTheDocument();
    expect(screen.getByText('get_user_dashboard')).toBeInTheDocument();

    // Default Streamable HTTP snippet
    expect(screen.getByTestId('mcp-config-snippet').textContent).toContain('/mcp');
  });

  it('renders disabled badge and _MCP_ENABLED hint when mcpEnabled is false', () => {
    render(
      <McpPluginModal
        isOpen={true}
        onClose={vi.fn()}
        mcpEnabled={false}
        overview={mockOverview}
        governance={mockGovernance}
      />
    );

    expect(screen.getByTestId('mcp-status-badge')).toHaveTextContent(/Disabled/i);
    expect(screen.getByText(/_MCP_ENABLED/i)).toBeInTheDocument();
  });

  it('switches between Streamable HTTP and Stdio mcp_config.json snippets', () => {
    render(
      <McpPluginModal
        isOpen={true}
        onClose={vi.fn()}
        mcpEnabled={true}
        overview={mockOverview}
        governance={mockGovernance}
      />
    );

    const stdioTab = screen.getByRole('button', { name: /stdio/i });
    fireEvent.click(stdioTab);

    expect(screen.getByTestId('mcp-config-snippet').textContent).toContain('"command": "agy-cost-board"');
    expect(screen.getByTestId('mcp-config-snippet').textContent).toContain('"_MCP_ENABLED": "true"');
  });

  it('calls onClose when Escape key is pressed or close button is clicked', () => {
    const onClose = vi.fn();
    render(
      <McpPluginModal
        isOpen={true}
        onClose={onClose}
        mcpEnabled={true}
        overview={mockOverview}
        governance={mockGovernance}
      />
    );

    fireEvent.keyDown(window, { key: 'Escape' });
    expect(onClose).toHaveBeenCalledTimes(1);
  });
});
