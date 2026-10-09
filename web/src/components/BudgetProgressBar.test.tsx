import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { BudgetProgressBar } from './BudgetProgressBar';
import { UserBudgetMetrics } from '../api';

describe('BudgetProgressBar Component', () => {
  const onTrackBudget: UserBudgetMetrics = {
    threshold: 100,
    current_spend: 35,
    utilization_percent: 35,
    projected_month_end_spend: 70,
    projected_overage: 0,
    on_track: true,
  };

  const overBudget: UserBudgetMetrics = {
    threshold: 100,
    current_spend: 85,
    utilization_percent: 85,
    projected_month_end_spend: 140,
    projected_overage: 40,
    on_track: false,
  };

  it('renders budget progress when on track', () => {
    render(
      <BudgetProgressBar
        budget={onTrackBudget}
        currency="USD"
        onUpdateBudget={vi.fn()}
      />
    );

    expect(screen.getByText('Monthly Budget & Quota Driving')).toBeInTheDocument();
    expect(screen.getByText('$35.00')).toBeInTheDocument();
    expect(screen.getByText('of $100.00')).toBeInTheDocument();
    expect(screen.getByText('On Track')).toBeInTheDocument();
    expect(screen.getByText(/projected \$70.00 at month end/i)).toBeInTheDocument();
  });

  it('renders alert when projected to exceed threshold', () => {
    render(
      <BudgetProgressBar
        budget={overBudget}
        currency="USD"
        onUpdateBudget={vi.fn()}
      />
    );

    expect(screen.getByText('Risk of Overage')).toBeInTheDocument();
    expect(screen.getByText(/projected \$140.00 at month end/i)).toBeInTheDocument();
    expect(screen.getByText(/over budget by \$40.00/i)).toBeInTheDocument();
  });

  it('allows editing and updating budget threshold', () => {
    const handleUpdate = vi.fn();
    render(
      <BudgetProgressBar
        budget={onTrackBudget}
        currency="USD"
        onUpdateBudget={handleUpdate}
      />
    );

    const editButton = screen.getByRole('button', { name: /set budget/i });
    fireEvent.click(editButton);

    const input = screen.getByPlaceholderText('100');
    fireEvent.change(input, { target: { value: '150' } });

    const saveButton = screen.getByRole('button', { name: /save/i });
    fireEvent.click(saveButton);

    expect(handleUpdate).toHaveBeenCalledWith(150);
  });
});
