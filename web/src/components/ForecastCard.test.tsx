import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { ForecastCard } from './ForecastCard';
import { UsageForecast } from '../api';

const mockForecastWithSpillover: UsageForecast = {
  horizon_days: 30,
  daily_velocity_wtu: 2400000,
  daily_pool_eligible_wtu: 2100000,
  daily_always_payg_wtu: 300000,
  daily_velocity_usd: 4.8,
  projected_30d_total_wtu: 72000000,
  projected_30d_total_usd: 144.0,
  projected_30d_included_usd: 112.5,
  projected_30d_payg_usd: 31.5,
  will_spillover_in_current_window: true,
  spillover_date: '2026-10-20',
  days_until_spillover: 2,
  forecast_days: Array.from({ length: 30 }, (_, idx) => {
    const dayNum = idx + 1;
    const isSpill = dayNum % 7 >= 5;
    return {
      date: `2026-10-${String((idx % 28) + 1).padStart(2, '0')}`,
      window_start: idx < 7 ? '2026-10-15' : '2026-10-22',
      window_end: idx < 7 ? '2026-10-21' : '2026-10-28',
      projected_wtu: 2400000,
      included_wtu: isSpill ? 1500000 : 2100000,
      payg_overage_wtu: isSpill ? 900000 : 300000,
      included_cost_usd: isSpill ? 3.0 : 4.2,
      payg_overage_usd: isSpill ? 1.8 : 0.6,
      cumulative_pool_wtu: Math.min(16250000, ((idx % 7) + 1) * 2100000),
      pool_limit_wtu: 16250000,
      spillover_triggered: isSpill,
    };
  }),
};

describe('ForecastCard', () => {
  it('renders nothing when forecast is undefined', () => {
    const { container } = render(<ForecastCard forecast={undefined} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders 30-day forecast KPIs, predicted spillover badge, and window-aware daily projection bars', () => {
    render(<ForecastCard forecast={mockForecastWithSpillover} currency="USD" />);

    expect(screen.getByText(/30-Day Window-Aware Usage & PAYG Forecast/i)).toBeInTheDocument();
    expect(screen.getByText(/Spillover on 2026-10-20/i)).toBeInTheDocument();
    expect(screen.getByText(/\$144\.00/)).toBeInTheDocument();
    expect(screen.getByText(/\$112\.50/)).toBeInTheDocument();
    expect(screen.getByText(/\$31\.50/)).toBeInTheDocument();
    expect(screen.getByText(/\$4\.80\/day/i)).toBeInTheDocument();
  });

  it('renders Within Weekly Pool badge when no spillover is predicted', () => {
    const safeForecast: UsageForecast = {
      ...mockForecastWithSpillover,
      will_spillover_in_current_window: false,
      spillover_date: '',
      days_until_spillover: -1,
    };
    render(<ForecastCard forecast={safeForecast} currency="USD" />);

    expect(screen.getByText(/Within Weekly Pool/i)).toBeInTheDocument();
  });

  it('toggles between USD and WTU display modes', () => {
    render(<ForecastCard forecast={mockForecastWithSpillover} currency="USD" />);

    const wtuBtn = screen.getByRole('button', { name: /WTU/i });
    fireEvent.click(wtuBtn);

    expect(screen.getAllByText(/72\.00M WTU/i).length).toBeGreaterThan(0);
  });
});
