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
});
