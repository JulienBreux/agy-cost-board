import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { WeeklyAllowancePoolCard } from './WeeklyAllowancePoolCard';
import { WeeklyAllowancePool } from '../api';

const mockPool: WeeklyAllowancePool = {
  window_start: '2026-10-12',
  window_end: '2026-10-18',
  reset_in_hours: 54,
  reset_countdown: '2d 6h',
  pool_mode: 'pooled',
  seat_basis: 'assigned',
  pool_limit_usd: 32.5,
  pool_limit_wtu: 16250000,
  included_used_usd: 26.0,
  included_used_wtu: 13000000,
  payg_overage_usd: 4.5,
  payg_overage_wtu: 2250000,
  remaining_usd: 6.5,
  remaining_wtu: 3250000,
  utilization_pct: 80.0,
  windows: [
    {
      window_label: 'Oct 05 - Oct 11',
      window_start: '2026-10-05',
      window_end: '2026-10-11',
      pool_limit_usd: 32.5,
      pool_limit_wtu: 16250000,
      included_used_usd: 29.0,
      included_used_wtu: 14500000,
      payg_overage_usd: 1.2,
      payg_overage_wtu: 600000,
      utilization_pct: 89.2,
    },
    {
      window_label: 'Oct 12 - Oct 18',
      window_start: '2026-10-12',
      window_end: '2026-10-18',
      pool_limit_usd: 32.5,
      pool_limit_wtu: 16250000,
      included_used_usd: 26.0,
      included_used_wtu: 13000000,
      payg_overage_usd: 4.5,
      payg_overage_wtu: 2250000,
      utilization_pct: 80.0,
    },
  ],
  scenarios: {
    pooled_assigned: {
      pool_mode: 'pooled',
      seat_basis: 'assigned',
      pool_limit_usd: 32.5,
      pool_limit_wtu: 16250000,
      included_used_usd: 26.0,
      included_used_wtu: 13000000,
      payg_overage_usd: 4.5,
      payg_overage_wtu: 2250000,
      remaining_usd: 6.5,
      remaining_wtu: 3250000,
      utilization_pct: 80.0,
    },
    pooled_purchased: {
      pool_mode: 'pooled',
      seat_basis: 'purchased',
      pool_limit_usd: 45.0,
      pool_limit_wtu: 22500000,
      included_used_usd: 30.5,
      included_used_wtu: 15250000,
      payg_overage_usd: 0.0,
      payg_overage_wtu: 0,
      remaining_usd: 14.5,
      remaining_wtu: 7250000,
      utilization_pct: 67.8,
    },
    per_user_assigned: {
      pool_mode: 'per_user',
      seat_basis: 'assigned',
      pool_limit_usd: 32.5,
      pool_limit_wtu: 16250000,
      included_used_usd: 21.0,
      included_used_wtu: 10500000,
      payg_overage_usd: 9.5,
      payg_overage_wtu: 4750000,
      remaining_usd: 11.5,
      remaining_wtu: 5750000,
      utilization_pct: 64.6,
    },
    per_user_purchased: {
      pool_mode: 'per_user',
      seat_basis: 'purchased',
      pool_limit_usd: 45.0,
      pool_limit_wtu: 22500000,
      included_used_usd: 21.0,
      included_used_wtu: 10500000,
      payg_overage_usd: 9.5,
      payg_overage_wtu: 4750000,
      remaining_usd: 24.0,
      remaining_wtu: 12000000,
      utilization_pct: 46.7,
    },
  },
};

describe('WeeklyAllowancePoolCard', () => {
  it('renders nothing when pool is undefined', () => {
    const { container } = render(<WeeklyAllowancePoolCard pool={undefined} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders 7-day drawdown bar, reset countdown, and weekly window history', () => {
    render(<WeeklyAllowancePoolCard pool={mockPool} currency="USD" />);

    expect(screen.getByText(/Weekly Allowance Pool/i)).toBeInTheDocument();
    expect(screen.getByText(/2d 6h/i)).toBeInTheDocument();
    expect(screen.getAllByText('80.0%').length).toBeGreaterThan(0);
    expect(screen.getByText('$26.00')).toBeInTheDocument();
    expect(screen.getAllByText(/\$32\.50/).length).toBeGreaterThan(0);
    expect(screen.getByText('Oct 05 - Oct 11')).toBeInTheDocument();
    expect(screen.getByText('Oct 12 - Oct 18')).toBeInTheDocument();
  });

  it('switches What-If scenarios interactively between Pooled / Per-User and Assigned / Purchased', () => {
    render(<WeeklyAllowancePoolCard pool={mockPool} currency="USD" />);

    // Switch to Purchased Quota
    const purchasedBtn = screen.getByRole('button', { name: /Purchased Quota/i });
    fireEvent.click(purchasedBtn);

    expect(screen.getByText('67.8%')).toBeInTheDocument();
    expect(screen.getByText('$30.50')).toBeInTheDocument();
    expect(screen.getAllByText(/\$45\.00/).length).toBeGreaterThan(0);

    // Switch to Per-User Cap
    const perUserBtn = screen.getByRole('button', { name: /Per-User Cap/i });
    fireEvent.click(perUserBtn);

    expect(screen.getByText('46.7%')).toBeInTheDocument();
    expect(screen.getByText('$21.00')).toBeInTheDocument();
    expect(screen.getByText('$9.50')).toBeInTheDocument();
  });

  it('renders Cloud Monitoring Quota Meter vs. Log Estimate reconciliation banner when monitoring_meter is provided', () => {
    const poolWithMeter: WeeklyAllowancePool = {
      ...mockPool,
      monitoring_meter: {
        metric_type: 'discoveryengine.googleapis.com/quota/agentic_wtu_tier_pp_weekly_regional',
        metered_used_wtu: 15523500,
        metered_used_usd: 31.05,
        log_estimated_wtu: 15250000,
        log_estimated_usd: 30.5,
        drift_wtu: 273500,
        drift_pct: 1.79,
        status: 'synced',
        last_synced_at: '2026-10-18T12:00:00Z',
      },
    };

    render(<WeeklyAllowancePoolCard pool={poolWithMeter} currency="USD" />);

    expect(screen.getByText(/Cloud Monitoring Quota Meter/i)).toBeInTheDocument();
    expect(
      screen.getByText(/discoveryengine\.googleapis\.com\/quota\/agentic_wtu_tier_pp_weekly_regional/i)
    ).toBeInTheDocument();
    expect(screen.getByText(/Synced/i)).toBeInTheDocument();
    expect(screen.getByText(/\+1\.79%/i)).toBeInTheDocument();
  });
});
