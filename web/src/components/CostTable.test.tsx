import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { CostTable } from './CostTable';
import { AllocatedUserCost } from '../api';

const mockCosts: AllocatedUserCost[] = [
  {
    usage_date: '2026-10-18',
    user_id: 'sophia.chen@google.com',
    model: 'gemini-2.5-pro',
    user_tokens: 1500000,
    wtu: 2100000,
    cache_hit_rate: 64.2,
    total_model_tokens: 3000000,
    token_share: 0.5,
    allocated_cost: 45.0,
    included_cost: 30.0,
    payg_cost: 15.0,
    fair_share_included_cost: 30.0,
    fair_share_payg_cost: 15.0,
    when_used_included_cost: 18.0,
    when_used_payg_cost: 27.0,
    currency: 'EUR',
  },
];

describe('CostTable Component', () => {
  it('renders attributed cost rows and toggles between Fair Share and When Used chargeback modes', () => {
    const onSelectUser = vi.fn();
    render(<CostTable costs={mockCosts} onSelectUser={onSelectUser} />);

    expect(screen.getByText('sophia.chen@google.com')).toBeInTheDocument();
    expect(screen.getByText('€45.00')).toBeInTheDocument();
    expect(screen.getByText(/Inc: €30\.00/)).toBeInTheDocument();
    expect(screen.getByText(/PAYG: €15\.00/)).toBeInTheDocument();

    // Switch to When Used
    const whenUsedBtn = screen.getByRole('button', { name: /when used/i });
    fireEvent.click(whenUsedBtn);

    expect(screen.getByText(/Inc: €18\.00/)).toBeInTheDocument();
    expect(screen.getByText(/PAYG: €27\.00/)).toBeInTheDocument();

    // Switch back to Fair Share
    const fairShareBtn = screen.getByRole('button', { name: /fair share/i });
    fireEvent.click(fairShareBtn);

    expect(screen.getByText(/Inc: €30\.00/)).toBeInTheDocument();
    expect(screen.getByText(/PAYG: €15\.00/)).toBeInTheDocument();

    // Click Export CSV
    const exportBtn = screen.getByRole('button', { name: /export csv/i });
    expect(exportBtn).toBeInTheDocument();
    fireEvent.click(exportBtn);
  });
});
