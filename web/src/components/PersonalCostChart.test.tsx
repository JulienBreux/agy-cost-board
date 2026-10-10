import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { PersonalCostChart } from './PersonalCostChart';
import { PersonalDailyTrend, ModelDetail } from '../api';

describe('PersonalCostChart Component', () => {
  const mockTrends: PersonalDailyTrend[] = [
    {
      date: '2026-10-07',
      cost: 1.25,
      total_tokens: 300000,
      models: { 'gemini-1.5-pro': 0.95, 'gemini-1.5-flash': 0.3 },
    },
    {
      date: '2026-10-08',
      cost: 2.5,
      total_tokens: 650000,
      models: { 'gemini-1.5-pro': 1.9, 'gemini-1.5-flash': 0.6 },
    },
    {
      date: '2026-10-09',
      cost: 0.85,
      total_tokens: 220000,
      models: { 'gemini-1.5-pro': 0.65, 'gemini-1.5-flash': 0.2 },
    },
  ];

  const mockModelDistribution: Record<string, ModelDetail> = {
    'gemini-1.5-pro': { cost: 3.5, tokens: 900000, share: 76.1 },
    'gemini-1.5-flash': { cost: 1.1, tokens: 270000, share: 23.9 },
  };

  it('renders daily spend trajectory and model distribution breakdown', () => {
    render(
      <PersonalCostChart
        trends={mockTrends}
        modelDistribution={mockModelDistribution}
        currency="USD"
      />
    );

    expect(screen.getByText('Daily Spend Trajectory')).toBeInTheDocument();
    expect(screen.getByText('Personal Model Breakdown')).toBeInTheDocument();

    expect(screen.getByText('gemini-1.5-pro')).toBeInTheDocument();
    expect(screen.getByText('$3.50')).toBeInTheDocument();
    expect(screen.getByText('(76.1%)')).toBeInTheDocument();

    expect(screen.getByText('gemini-1.5-flash')).toBeInTheDocument();
    expect(screen.getByText('$1.10')).toBeInTheDocument();
    expect(screen.getByText('(23.9%)')).toBeInTheDocument();
  });

  it('renders empty state when no daily trend data exists', () => {
    render(
      <PersonalCostChart
        trends={[]}
        modelDistribution={{}}
        currency="USD"
      />
    );

    expect(screen.getByText(/no daily spend recorded in this period/i)).toBeInTheDocument();
    expect(screen.getByText(/no model activity recorded/i)).toBeInTheDocument();
  });

  it('updates display when hovering on a trend bar', () => {
    render(
      <PersonalCostChart
        trends={mockTrends}
        modelDistribution={mockModelDistribution}
        currency="USD"
      />
    );

    // Initial total in banner and donut center
    expect(screen.getAllByText('$4.60').length).toBeGreaterThanOrEqual(1);

    // Hover on second bar
    const bars = screen.getAllByTestId('trend-bar');
    expect(bars).toHaveLength(3);
    fireEvent.mouseEnter(bars[1]);

    expect(screen.getAllByText('2026-10-08').length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText('$2.50')).toBeInTheDocument();
    expect(screen.getByText(/650,000 tokens/i)).toBeInTheDocument();

    // Check Day Repartition pills
    expect(screen.getByText(/day repartition:/i)).toBeInTheDocument();
    expect(screen.getByText('$1.90')).toBeInTheDocument();
    expect(screen.getByText('$0.60')).toBeInTheDocument();
  });

  it('switches between stacked and aggregate chart modes', () => {
    render(
      <PersonalCostChart
        trends={mockTrends}
        modelDistribution={mockModelDistribution}
        currency="USD"
      />
    );

    // Click Aggregate
    const aggregateBtn = screen.getByRole('button', { name: /aggregate/i });
    fireEvent.click(aggregateBtn);

    expect(screen.getAllByText('Aggregate').length).toBeGreaterThanOrEqual(2);

    // Click Stacked
    const stackedBtn = screen.getByRole('button', { name: /stacked/i });
    fireEvent.click(stackedBtn);

    expect(screen.getByText('Model Breakdown')).toBeInTheDocument();
  });

  it('switches between Cost and Tokens metric modes', () => {
    render(
      <PersonalCostChart
        trends={mockTrends}
        modelDistribution={mockModelDistribution}
        currency="USD"
      />
    );

    const tokensBtn = screen.getByRole('button', { name: /tokens/i });
    fireEvent.click(tokensBtn);

    // Center donut should show token count (900k + 270k = 1.17M)
    expect(screen.getByText('1.17M tokens')).toBeInTheDocument();
  });

  it('highlights model when hovered in the breakdown list', () => {
    render(
      <PersonalCostChart
        trends={mockTrends}
        modelDistribution={mockModelDistribution}
        currency="USD"
      />
    );

    const modelItem = screen.getByTestId('model-item-gemini-1.5-pro');
    fireEvent.mouseEnter(modelItem);

    // Center of donut should now display the focused model percentage
    expect(screen.getByText('76.1%')).toBeInTheDocument();

    fireEvent.mouseLeave(modelItem);
    expect(screen.getByText('Total Spend')).toBeInTheDocument();
  });
});
