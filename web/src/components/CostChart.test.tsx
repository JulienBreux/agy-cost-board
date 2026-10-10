import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { CostChart } from './CostChart';
import { DailyTrend, ModelDetail } from '../api';

describe('CostChart Component', () => {
  const mockTrends: DailyTrend[] = [
    {
      date: '2026-10-07',
      cost: 10.5,
      tokens: 250000,
      active_users: 3,
      by_model: {
        'gemini-1.5-pro': 6.0,
        'gemini-1.5-flash': 4.5,
      },
    },
    {
      date: '2026-10-08',
      cost: 20.0,
      tokens: 500000,
      active_users: 5,
      by_model: {
        'gemini-1.5-pro': 12.0,
        'gemini-1.5-flash': 8.0,
      },
    },
    {
      date: '2026-10-09',
      cost: 15.0,
      tokens: 350000,
      active_users: 4,
      by_model: {
        'gemini-1.5-pro': 9.0,
        'gemini-1.5-flash': 6.0,
      },
    },
  ];

  const mockModelBreakdown: Record<string, ModelDetail> = {
    'gemini-1.5-pro': { cost: 27.0, tokens: 660000, share: 60.0 },
    'gemini-1.5-flash': { cost: 18.5, tokens: 440000, share: 40.0 },
  };

  it('renders daily spend chart and model repartition donut card', () => {
    render(
      <CostChart
        trends={mockTrends}
        modelBreakdown={mockModelBreakdown}
        currency="USD"
      />
    );

    expect(screen.getByText('Daily Reconciled Cloud Spend')).toBeInTheDocument();
    expect(screen.getByText('Model Repartition')).toBeInTheDocument();
    expect(screen.getByText('2 Models')).toBeInTheDocument();

    // Verify models in breakdown
    expect(screen.getByText('gemini-1.5-pro')).toBeInTheDocument();
    expect(screen.getByText('$27.00')).toBeInTheDocument();
    expect(screen.getByText('(59.3%)')).toBeInTheDocument();

    expect(screen.getByText('gemini-1.5-flash')).toBeInTheDocument();
    expect(screen.getByText('$18.50')).toBeInTheDocument();
    expect(screen.getByText('(40.7%)')).toBeInTheDocument();

    // Verify total in donut center and summary banner
    expect(screen.getByText('Total Spend')).toBeInTheDocument();
    expect(screen.getAllByText('$45.50').length).toBeGreaterThanOrEqual(1);
  });

  it('renders friendly empty states when no trend or model data is provided', () => {
    render(<CostChart trends={[]} modelBreakdown={{}} />);

    expect(screen.getByText(/no daily spend activity recorded in this period/i)).toBeInTheDocument();
    expect(screen.getByText(/no ai model consumption recorded in this period/i)).toBeInTheDocument();
  });

  it('updates dynamic summary and day repartition pills when hovering on a trend bar', () => {
    render(
      <CostChart
        trends={mockTrends}
        modelBreakdown={mockModelBreakdown}
        currency="USD"
      />
    );

    const bars = screen.getAllByTestId('trend-bar');
    expect(bars).toHaveLength(3);

    // Hover on second day (2026-10-08)
    fireEvent.mouseEnter(bars[1]);

    expect(screen.getAllByText('2026-10-08').length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText('(5 active users)')).toBeInTheDocument();
    expect(screen.getByText('$20.00')).toBeInTheDocument();
    expect(screen.getByText('500,000')).toBeInTheDocument();

    // Check Day Repartition pills
    expect(screen.getByText(/day repartition:/i)).toBeInTheDocument();
    expect(screen.getByText('$12.00')).toBeInTheDocument();
    expect(screen.getByText('$8.00')).toBeInTheDocument();

    // Mouse leave restores window total
    fireEvent.mouseLeave(bars[1]);
    expect(screen.getByText(/window total/i)).toBeInTheDocument();
  });

  it('switches between stacked and aggregate chart modes', () => {
    render(
      <CostChart
        trends={mockTrends}
        modelBreakdown={mockModelBreakdown}
        currency="USD"
      />
    );

    // Initially in Stacked mode
    expect(screen.getByText('Model Breakdown')).toBeInTheDocument();

    // Click Aggregate
    const aggregateBtn = screen.getByRole('button', { name: /aggregate/i });
    fireEvent.click(aggregateBtn);

    expect(screen.getAllByText('Aggregate').length).toBeGreaterThanOrEqual(2);

    // Click Stacked back
    const stackedBtn = screen.getByRole('button', { name: /stacked/i });
    fireEvent.click(stackedBtn);

    expect(screen.getByText('Model Breakdown')).toBeInTheDocument();
  });

  it('switches between Cost and Tokens metric modes', () => {
    render(
      <CostChart
        trends={mockTrends}
        modelBreakdown={mockModelBreakdown}
        currency="USD"
      />
    );

    const tokensBtn = screen.getByRole('button', { name: /tokens/i });
    fireEvent.click(tokensBtn);

    // Center donut should show token count
    expect(screen.getByText('1.10M tokens')).toBeInTheDocument();
  });

  it('highlights model when hovered in the breakdown list', () => {
    render(
      <CostChart
        trends={mockTrends}
        modelBreakdown={mockModelBreakdown}
        currency="USD"
      />
    );

    const modelItem = screen.getByTestId('model-item-gemini-1.5-pro');
    fireEvent.mouseEnter(modelItem);

    // Center of donut should now display the focused model details
    expect(screen.getByText('59.3%')).toBeInTheDocument();

    fireEvent.mouseLeave(modelItem);
    expect(screen.getByText('Total Spend')).toBeInTheDocument();
  });
});
