import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { LicenseTable } from './LicenseTable';
import { LicenseGovernance } from '../api';

const mockGovernance: LicenseGovernance = {
  seat_quota: 15,
  assigned_seats: 12,
  active_seats: 9,
  dormant_seats: 3,
  standard_seats: 8,
  plus_seats: 4,
  trial_seats: 2,
  locations: ['global', 'eu', 'us'],
  utilization_pct: 80.0,
  estimated_monthly_savings: 135.0,
  currency: 'USD',
  weekly_pool: {
    window_start: '2026-10-12',
    window_end: '2026-10-18',
    reset_in_hours: 48,
    reset_countdown: '2d 0h',
    pool_mode: 'pooled',
    seat_basis: 'assigned',
    pool_limit_usd: 35.0,
    pool_limit_wtu: 17500000,
    included_used_usd: 24.5,
    included_used_wtu: 12250000,
    payg_overage_usd: 2.0,
    payg_overage_wtu: 1000000,
    remaining_usd: 10.5,
    remaining_wtu: 5250000,
    utilization_pct: 70.0,
  },
  dormant_users: [
    {
      user_id: 'dormant.dev@google.com',
      status: 'dormant',
      tier: 'standard',
      weekly_allowance_usd: 2.5,
      weekly_allowance_wtu: 1250000,
      location: 'eu',
      is_trial: true,
      trial_ends_at: '2026-10-30',
      last_activity: '',
      total_tokens_in_window: 0,
    },
  ],
  all_seats: [
    {
      user_id: 'alex.turner@google.com',
      status: 'active',
      tier: 'plus',
      weekly_allowance_usd: 3.75,
      weekly_allowance_wtu: 1875000,
      location: 'us',
      is_trial: false,
      last_activity: '2026-10-18T10:00:00Z',
      total_tokens_in_window: 1450000,
    },
    {
      user_id: 'dormant.dev@google.com',
      status: 'dormant',
      tier: 'standard',
      weekly_allowance_usd: 2.5,
      weekly_allowance_wtu: 1250000,
      location: 'eu',
      is_trial: true,
      trial_ends_at: '2026-10-30',
      last_activity: '',
      total_tokens_in_window: 0,
    },
  ],
};

describe('LicenseTable', () => {
  it('renders WeeklyAllowancePoolCard, Multi-Tier summary, Location badges, and Free Trial badges', () => {
    const onSelectUser = vi.fn();
    render(<LicenseTable governance={mockGovernance} onSelectUser={onSelectUser} />);

    // Weekly Allowance Pool card
    expect(screen.getByText(/Weekly Allowance Pool/i)).toBeInTheDocument();

    // Multi-Tier & Trial summary badges
    expect(screen.getAllByText(/Plus/i).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/Standard/i).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/Free Trial/i).length).toBeGreaterThan(0);

    // Location badges
    expect(screen.getAllByText('EU').length).toBeGreaterThan(0);
    expect(screen.getAllByText('US').length).toBeGreaterThan(0);

    // Clicking a user row triggers onSelectUser
    fireEvent.click(screen.getByText('alex.turner@google.com'));
    expect(onSelectUser).toHaveBeenCalledWith('alex.turner@google.com');
  });
});
