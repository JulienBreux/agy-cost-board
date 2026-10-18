import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { OptimizationCard } from './OptimizationCard';
import type { OptimizationTip } from '../api';

const mockTips: OptimizationTip[] = [
  {
    type: 'model_switch',
    title: 'Switch Routine Tasks to Flash',
    description: 'You used Gemini Pro for 65% of your queries. Switching routine coding assistance to Gemini Flash could cut model cost by 40%.',
    potential_savings_monthly: 24.50,
    severity: 'warning'
  },
  {
    type: 'caching',
    title: 'Enable Context Caching',
    description: 'High repetitive token prompt volume detected. Caching system instructions can reduce prompt costs up to 75%.',
    potential_savings_monthly: 12.00,
    severity: 'info'
  }
];

describe('OptimizationCard', () => {
  it('renders optimization tips and total potential monthly savings', () => {
    render(<OptimizationCard tips={mockTips} currency="$" />);

    expect(screen.getByText('Optimization Insights')).toBeInTheDocument();
    expect(screen.getByText('Switch Routine Tasks to Flash')).toBeInTheDocument();
    expect(screen.getByText('Enable Context Caching')).toBeInTheDocument();

    // Check potential total savings calculation: 24.50 + 12.00 = $36.50
    expect(screen.getByText(/\$36\.50/)).toBeInTheDocument();

    // Check individual savings badges
    expect(screen.getByText('Save $24.50/mo')).toBeInTheDocument();
    expect(screen.getByText('Save $12.00/mo')).toBeInTheDocument();
  });

  it('renders clean empty state when there are no tips', () => {
    render(<OptimizationCard tips={[]} />);

    expect(screen.getByText(/Your consumption is well-optimized/i)).toBeInTheDocument();
  });

  it('triggers onApplyTip when action button is clicked', () => {
    const onApply = vi.fn();
    render(<OptimizationCard tips={mockTips} onApplyTip={onApply} />);

    const buttons = screen.getAllByRole('button', { name: /apply|learn more/i });
    expect(buttons.length).toBeGreaterThan(0);
    fireEvent.click(buttons[0]);

    expect(onApply).toHaveBeenCalledWith(mockTips[0]);
  });

  it('renders EUR formatted savings properly', () => {
    render(<OptimizationCard tips={mockTips} currency="EUR" />);

    expect(screen.getByText(/\+€36\.50/)).toBeInTheDocument();
    expect(screen.getByText('Save €24.50/mo')).toBeInTheDocument();
    expect(screen.getByText('Save €12.00/mo')).toBeInTheDocument();
  });

  it('formats effort-tier-downgrade and prompt-cache-warming tip types with human-readable labels', () => {
    const finopsTips: OptimizationTip[] = [
      {
        type: 'effort-tier-downgrade',
        title: 'Right-size High/Max Thinking Effort',
        description: '42% of your requests run at high/max thinking effort.',
        potential_savings_monthly: 18.2,
        severity: 'warning',
      },
      {
        type: 'prompt-cache-warming',
        title: 'Improve Context Cache Reuse',
        description: 'Your cache hit rate is 48% vs org 75%.',
        potential_savings_monthly: 9.5,
        severity: 'info',
      },
    ];

    render(<OptimizationCard tips={finopsTips} currency="USD" />);

    expect(screen.getByText('Thinking Effort')).toBeInTheDocument();
    expect(screen.getByText('Context Caching')).toBeInTheDocument();
  });
});

