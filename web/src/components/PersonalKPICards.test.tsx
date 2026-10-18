import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { PersonalKPICards } from './PersonalKPICards';
import { UserConsumptionKPIs } from '../api';

describe('PersonalKPICards Component', () => {
  const mockKpis: UserConsumptionKPIs = {
    mtd_spend: 34.5,
    daily_burn_rate: 1.15,
    weekly_burn_rate: 8.05,
    total_user_tokens: 8500000,
    org_spend_share_percent: 4.8,
  };

  it('renders all personal KPI metrics properly', () => {
    render(<PersonalKPICards kpis={mockKpis} currency="USD" />);

    expect(screen.getByText('Month-to-Date Spend')).toBeInTheDocument();
    expect(screen.getByText('$34.50')).toBeInTheDocument();

    expect(screen.getByText('Weekly Burn Rate')).toBeInTheDocument();
    expect(screen.getByText('$8.05')).toBeInTheDocument();

    expect(screen.getByText('Personal Inference Tokens')).toBeInTheDocument();
    expect(screen.getByText('8.50M')).toBeInTheDocument();

    expect(screen.getByText('Org Spend Share')).toBeInTheDocument();
    expect(screen.getByText('4.8%')).toBeInTheDocument();
  });

  it('renders safely without throwing when kpis is undefined', () => {
    render(<PersonalKPICards kpis={undefined as any} currency="USD" />);

    expect(screen.getByText('Month-to-Date Spend')).toBeInTheDocument();
    expect(screen.getAllByText('$0.00')).toHaveLength(2);
    expect(screen.getByText('Weekly Burn Rate')).toBeInTheDocument();
    expect(screen.getByText('Personal Inference Tokens')).toBeInTheDocument();
    expect(screen.getByText('0.00M')).toBeInTheDocument();
    expect(screen.getByText('Org Spend Share')).toBeInTheDocument();
    expect(screen.getByText('0.0%')).toBeInTheDocument();
  });

  it('renders EUR formatted values properly', () => {
    render(<PersonalKPICards kpis={mockKpis} currency="EUR" />);

    expect(screen.getByText('€34.50')).toBeInTheDocument();
    expect(screen.getByText('€8.05')).toBeInTheDocument();
    expect(screen.getByText(/€1\.15\/day average velocity/)).toBeInTheDocument();
  });

  it('renders Included vs PAYG coexistence and WTU + Cache Hit % when provided', () => {
    const kpisWithFinOps: UserConsumptionKPIs = {
      ...mockKpis,
      included_cost: 22.0,
      payg_cost: 12.5,
      total_wtu: 6400000,
      cache_hit_rate: 79.2,
      top_effort: 'high',
    };

    render(<PersonalKPICards kpis={kpisWithFinOps} currency="USD" />);

    expect(screen.getByText(/Included: \$22\.00/)).toBeInTheDocument();
    expect(screen.getByText(/PAYG: \$12\.50/)).toBeInTheDocument();
    expect(screen.getByText(/6\.40M WTU/)).toBeInTheDocument();
    expect(screen.getByText(/79\.2% cache/)).toBeInTheDocument();
  });
});

