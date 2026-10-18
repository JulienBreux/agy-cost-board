import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, within } from '@testing-library/react';
import { LeaderboardView } from './LeaderboardView';
import { LeaderboardEntry } from '../api';

const mockEntries: LeaderboardEntry[] = [
  {
    rank: 1,
    user_id: 'sophia.chen@google.com',
    total_cost: 420.5,
    spend_share_pct: 42.05,
    total_tokens: 18500000,
    token_share_pct: 37.0,
    active_days: 28,
    cost_per_1k_tokens: 0.0227,
    top_model: 'gemini-2.5-pro',
    model_breakdown: {
      'gemini-2.5-pro': { tokens: 12000000, cost: 350.0, share: 0.65 },
      'gemini-2.5-flash': { tokens: 6500000, cost: 70.5, share: 0.35 },
    },
    currency: 'EUR',
  },
  {
    rank: 2,
    user_id: 'marcus.vance@google.com',
    total_cost: 310.0,
    spend_share_pct: 31.0,
    total_tokens: 21000000,
    token_share_pct: 42.0,
    active_days: 25,
    cost_per_1k_tokens: 0.0148,
    top_model: 'gemini-2.5-flash',
    model_breakdown: {
      'gemini-2.5-flash': { tokens: 18000000, cost: 250.0, share: 0.86 },
      'gemini-2.5-pro': { tokens: 3000000, cost: 60.0, share: 0.14 },
    },
    currency: 'EUR',
  },
  {
    rank: 3,
    user_id: 'elena.rodriguez@google.com',
    total_cost: 190.25,
    spend_share_pct: 19.03,
    total_tokens: 7500000,
    token_share_pct: 15.0,
    active_days: 29,
    cost_per_1k_tokens: 0.0254,
    top_model: 'claude-3-7-sonnet',
    model_breakdown: {
      'claude-3-7-sonnet': { tokens: 7500000, cost: 190.25, share: 1.0 },
    },
    currency: 'EUR',
  },
  {
    rank: 4,
    user_id: 'david.kim@google.com',
    total_cost: 79.25,
    spend_share_pct: 7.93,
    total_tokens: 3000000,
    token_share_pct: 6.0,
    active_days: 12,
    cost_per_1k_tokens: 0.0264,
    top_model: 'gemini-2.5-flash-lite',
    model_breakdown: {
      'gemini-2.5-flash-lite': { tokens: 3000000, cost: 79.25, share: 1.0 },
    },
    currency: 'EUR',
  },
];

describe('LeaderboardView Component', () => {
  it('renders Top 3 Podium cards with native currency formatting and developer metrics', () => {
    const onSelectUser = vi.fn();
    render(<LeaderboardView entries={mockEntries} onSelectUser={onSelectUser} days={30} />);

    const podium1 = screen.getByTestId('podium-card-1');
    const podium2 = screen.getByTestId('podium-card-2');
    const podium3 = screen.getByTestId('podium-card-3');

    expect(within(podium1).getByText('sophia.chen@google.com')).toBeInTheDocument();
    expect(within(podium1).getByText('€420.50')).toBeInTheDocument();
    expect(within(podium1).getByText('gemini-2.5-pro')).toBeInTheDocument();

    expect(within(podium2).getByText('marcus.vance@google.com')).toBeInTheDocument();
    expect(within(podium2).getByText('€310.00')).toBeInTheDocument();

    expect(within(podium3).getByText('elena.rodriguez@google.com')).toBeInTheDocument();
    expect(within(podium3).getByText('€190.25')).toBeInTheDocument();

    // Clicking a podium card triggers onSelectUser
    fireEvent.click(podium1);
    expect(onSelectUser).toHaveBeenCalledWith('sophia.chen@google.com');
  });

  it('filters leaderboard table rows by developer search query', () => {
    const onSelectUser = vi.fn();
    render(<LeaderboardView entries={mockEntries} onSelectUser={onSelectUser} days={30} />);

    const searchInput = screen.getByPlaceholderText(/search developer/i);
    fireEvent.change(searchInput, { target: { value: 'david.kim' } });

    const table = screen.getByRole('table');
    expect(within(table).getByText('david.kim@google.com')).toBeInTheDocument();
    expect(within(table).queryByText('sophia.chen@google.com')).not.toBeInTheDocument();
  });

  it('sorts leaderboard table by Spend, Tokens, Active Days, and Efficiency', () => {
    const onSelectUser = vi.fn();
    render(<LeaderboardView entries={mockEntries} onSelectUser={onSelectUser} days={30} />);

    const table = screen.getByRole('table');
    const getRowUsers = () => {
      const rows = Array.from(table.querySelectorAll('tbody tr'));
      return rows.map((r) => r.getAttribute('data-user-id'));
    };

    // Default: Spend desc
    expect(getRowUsers()[0]).toBe('sophia.chen@google.com');

    // Sort by Tokens desc -> marcus.vance has 21M tokens (#1 by tokens)
    fireEvent.click(screen.getByRole('columnheader', { name: /tokens/i }));
    expect(getRowUsers()[0]).toBe('marcus.vance@google.com');

    // Sort by Active Days desc -> elena.rodriguez has 29 active days
    fireEvent.click(screen.getByRole('columnheader', { name: /active days/i }));
    expect(getRowUsers()[0]).toBe('elena.rodriguez@google.com');

    // Sort by Efficiency (Cost / 1K Tokens) desc -> david.kim has 0.0264
    fireEvent.click(screen.getByRole('columnheader', { name: /cost \/ 1k/i }));
    expect(getRowUsers()[0]).toBe('david.kim@google.com');
  });

  it('renders model mix bar and triggers onSelectUser when a table row is clicked', () => {
    const onSelectUser = vi.fn();
    render(<LeaderboardView entries={mockEntries} onSelectUser={onSelectUser} days={30} />);

    const modelMixBars = screen.getAllByTestId('model-mix-bar');
    expect(modelMixBars.length).toBe(4);

    const table = screen.getByRole('table');
    const row = within(table).getByText('david.kim@google.com');
    fireEvent.click(row);
    expect(onSelectUser).toHaveBeenCalledWith('david.kim@google.com');
  });

  it('safely renders raw Go backend payload with share_of_spend and numeric model_breakdown percentages without throwing', () => {
    const onSelectUser = vi.fn();
    const rawGoEntries = [
      {
        rank: 1,
        user_id: 'elena.rostova@google.com',
        total_cost: 460.24,
        share_of_spend: 19.32,
        total_tokens: 1088425,
        share_of_tokens: 18.97,
        active_days: 30,
        cost_per_1k_tokens: 0.4228,
        top_model: 'claude-5.5-sonnet-medium',
        model_breakdown: {
          'claude-5.5-sonnet-medium': 23.2,
          'gemini-4.0-pro': 20.7,
        },
        currency: 'EUR',
      } as unknown as LeaderboardEntry,
    ];

    render(<LeaderboardView entries={rawGoEntries} onSelectUser={onSelectUser} days={30} />);
    expect(screen.getByText('19.3% of org')).toBeInTheDocument();
    expect(screen.getAllByText('elena.rostova@google.com').length).toBeGreaterThan(0);
  });
});

