import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { KPICards } from './KPICards';
import { OverviewMetrics } from '../api';

describe('KPICards Component', () => {
  const baseMetrics: OverviewMetrics = {
    total_billed_cost: 1250.5,
    total_tokens: 45000000,
    active_users_count: 12,
    seat_quota: 20,
    assigned_seats: 15,
    utilization_pct: 75.0,
    estimated_monthly_savings: 150.0,
    daily_trends: [],
    model_breakdown: {},
    currency: 'USD',
  };

  it('renders USD formatted cost and savings by default', () => {
    render(<KPICards metrics={baseMetrics} />);

    expect(screen.getByText('Total Billed Cost')).toBeInTheDocument();
    expect(screen.getByText('$1,250.50')).toBeInTheDocument();
    expect(screen.getByText('$150.00')).toBeInTheDocument();
  });

  it('renders EUR formatted cost and savings when currency is EUR', () => {
    render(<KPICards metrics={{ ...baseMetrics, currency: 'EUR' }} />);

    expect(screen.getByText('€1,250.50')).toBeInTheDocument();
    expect(screen.getByText('€150.00')).toBeInTheDocument();
  });

  it('renders GBP formatted cost and savings when currency is GBP', () => {
    render(<KPICards metrics={{ ...baseMetrics, currency: 'GBP' }} />);

    expect(screen.getByText('£1,250.50')).toBeInTheDocument();
    expect(screen.getByText('£150.00')).toBeInTheDocument();
  });
});
